package main

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/friendsofgo/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/networkteam/devlog"
	"github.com/networkteam/devlog/collector"
	"github.com/networkteam/slogutils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
	sloghttp "github.com/samber/slog-http"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/graph/public"
	api_handler "myvendor.mytld/myproject/backend/api/handler"
	"myvendor.mytld/myproject/backend/api/handler/testapi"
	http_api "myvendor.mytld/myproject/backend/api/http"
	http_middleware "myvendor.mytld/myproject/backend/api/http/middleware"
	domain_handler "myvendor.mytld/myproject/backend/domain/handler"
)

const shutdownTimeout = 5 * time.Second

func newServerCmd() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Run the backend server",
		Flags: flattenFlags(
			[]cli.Flag{
				&cli.StringFlag{
					Name:    "address",
					Usage:   "Listen on this address",
					EnvVars: []string{"BACKEND_ADDRESS"},
					Value:   "0.0.0.0:8080",
				},
				&cli.StringFlag{
					Name:    "websocket-allow-origin",
					Usage:   "Allow websocket connections from this origin, if empty only the origin matching the host of the request is allowed",
					EnvVars: []string{"BACKEND_WEBSOCKET_ALLOW_ORIGIN"},
				},
				&cli.BoolFlag{
					Name:  "playground",
					Usage: "Enable GraphQL playground",
					Value: false,
				},

				&cli.BoolFlag{
					Name:    "pprof",
					Usage:   "Enable pprof for profiling the server",
					EnvVars: []string{"PPROF_ENABLED"},
				},

				&cli.BoolFlag{
					Name:    "open-telemetry-enabled",
					Usage:   "Enable open telemetry",
					EnvVars: []string{"OPEN_TELEMETRY_ENABLED"},
				},

				&cli.BoolFlag{
					Name:    "enable-test-api",
					Usage:   "Enable a REST API for test purposes (must not be enabled in production!)",
					EnvVars: []string{"ENABLE_TEST_API"},
				},

				&cli.DurationFlag{
					Name:    "sensitive-operation-constant-time",
					Usage:   "Constant time duration to wait for sensitive operations (e.g. login / request password reset / perform password reset / registration), to prevent timing attacks",
					EnvVars: []string{"SENSITIVE_OPERATION_CONSTANT_TIME"},
					Value:   700 * time.Millisecond,
				},

				&cli.StringFlag{
					Name:    "delete-expired-access-tokens-cron",
					Usage:   "Cron schedule for deleting expired access tokens",
					Value:   "@every 1h",
					EnvVars: []string{"BACKEND_DELETE_EXPIRED_ACCESS_TOKENS_CRON"},
				},
			},

			devlogFlags(),
			sentryFlags(),
		),
		Action: serverAction,
	}
}

func serverAction(c *cli.Context) (err error) {
	// This action is where the server is set up and dependencies are wired
	// -- make sure to keep it clean and with clear intention what is done here

	logger := slogutils.FromContext(c.Context)

	var dlog *devlog.Instance
	if c.Bool("devlog") {
		dlog = devlog.NewWithOptions(devlog.Options{
			HTTPServerOptions: devlogHTTPServerOptions(),
		})

		// Collect debug logs with devlog
		defaultHandler := logger.Handler()
		logger = setDefaultLoggerAndContext(c, slog.New(
			slogmulti.Fanout(
				dlog.CollectSlogLogs(collector.CollectSlogLogsOptions{
					Level: slog.LevelInfo,
				}),
				defaultHandler,
			),
		))

		// Add devlog RoundTripper to DefaultTransport to trace outgoing HTTP requests
		http.DefaultTransport = dlog.CollectHTTPClient(http.DefaultTransport)
	}

	err = initializeSentry(c, "backend")
	if err != nil {
		return err
	}

	if c.Bool("sentry-log-integration") {
		logger = initializeSentrySlog(c)
	}

	// Configure logger with process field _after_ Sentry slog integration to ensure process field is included in Sentry logs
	logger = setLogProcess(c, "server")

	db, err := connectDatabase(c, dlog)
	if err != nil {
		return err
	}

	mailer, err := buildMailer(c)
	if err != nil {
		return err
	}

	timeSource, err := newCurrentTimeSource(c)
	if err != nil {
		return err
	}

	// Set up signal handling, should be called before starting background processing
	setupCancelOnSignal(c)

	config, err := getConfig(c)
	if err != nil {
		return err
	}

	// Set up OpenTelemetry with global providers
	otelShutdown, err := setupOTelSDK(c, config)
	if err != nil {
		return err
	}
	defer func() {
		err = stderrors.Join(err, otelShutdown(context.Background()))
	}()

	queue := createAsynqQueue(c)

	deps := api.ResolverDependencies{
		DB:            db,
		TimeSource:    timeSource,
		Config:        config,
		Mailer:        mailer,
		Queue:         queue,
		MeterProvider: otel.GetMeterProvider(),
	}

	shutdownCronJobs, err := startCronJobs(c, deps.Handler())
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	apiHandlerConfig := api_handler.Config{
		EnableTracing:                  false,
		EnableLogging:                  true,
		EnableOpenTelemetry:            c.Bool("open-telemetry-enabled"),
		DisableRecover:                 false,
		WebsocketAllowOrigin:           c.String("websocket-allow-origin"),
		SensitiveOperationConstantTime: c.Duration("sensitive-operation-constant-time"),
	}
	publicExecutableSchema := public.BuildExecutableSchema(deps, apiHandlerConfig)
	publicGraphqlHandler := api_handler.NewGraphqlHandler(deps, apiHandlerConfig, publicExecutableSchema)

	playgroundEnabled := c.Bool("playground")
	if playgroundEnabled {
		mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
	}

	if c.Bool("open-telemetry-enabled") {
		publicGraphqlHandler = otelhttp.NewHandler(publicGraphqlHandler, "/query")
	}

	mux.Handle("/query", http_api.MiddlewareStackWithAuth(deps, publicGraphqlHandler))
	mux.HandleFunc("/healthz", api_handler.NewHealthzHandler(db))
	mux.Handle("/metrics", promhttp.Handler())

	rootHandler := http_middleware.RequestID(
		sloghttp.NewWithConfig(logger.WithGroup("http"), sloghttp.Config{
			DefaultLevel:     slog.LevelInfo,
			ServerErrorLevel: slog.LevelError,
			ClientErrorLevel: slog.LevelWarn,
			WithRequestID:    true,
		})(
			http_middleware.RequestIDLogger(
				http_middleware.RecoverAndLog(mux),
			),
		),
	)

	outerMux := http.NewServeMux()

	// Wrap the root handler with devlog middleware for HTTP server request tracing
	if c.Bool("devlog") {
		rootHandler = dlog.CollectHTTPServer(rootHandler)
	}
	outerMux.Handle("/", rootHandler)
	if c.Bool("devlog") {
		err = addDevlogDashboard(c, outerMux, dlog)
		if err != nil {
			return errors.Wrap(err, "adding devlog dashboard")
		}
	}
	// Add pprof handlers if enabled
	if c.Bool("pprof") {
		outerMux.HandleFunc("/debug/pprof/", pprof.Index)
		outerMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		outerMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		outerMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		outerMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	outerMux.HandleFunc("/healthz", api_handler.NewHealthzHandler(db))
	if c.Bool("enable-test-api") {
		logger.Warn("Test API enabled, don't do this in production!")
		outerMux.Handle("/test-api/", http.StripPrefix("/test-api", testapi.NewAPI(db)))
	}

	address := c.String("address")
	logger.Info("Serving GraphQL endpoint", "url", fmt.Sprintf("http://%s/query", address))
	if playgroundEnabled {
		logger.Info("Running GraphQL playground", "url", fmt.Sprintf("http://%s/", address))
	}
	if c.Bool("devlog") {
		logger.Info("Running devlog dashboard", "url", fmt.Sprintf("http://%s/_devlog/", address))
	}

	err = serve(c, rootHandler, func(_ *cli.Context) error {
		shutdownCronJobs()
		return nil
	})
	return err
}

func serve(c *cli.Context, handler http.Handler, onShutdown func(c *cli.Context) error) (err error) {
	logger := slogutils.FromContext(c.Context)

	address := c.String("address")
	srv := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 60 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return c.Context
		},
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Failed to listen and serve", slogutils.Err(err))
			os.Exit(1)
		}
	}()

	<-c.Context.Done()

	logger.Debug("Server stopped")

	// use background context because the original context is already cancelled
	ctxShutDown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		return errors.Wrap(err, "shutting down server")
	}

	logger.Debug("Server exited properly")

	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}

	if shutdownErr := onShutdown(c); shutdownErr != nil {
		err = multierror.Append(err, shutdownErr)
	}

	logger.Info("Everything shut down, goodbye")

	return err
}

func setupCancelOnSignal(c *cli.Context) {
	logger := slogutils.FromContext(c.Context)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		// kill -SIGINT XXXX or Ctrl+c
		syscall.SIGINT,
		// kill -SIGTERM XXXX
		syscall.SIGTERM,
	)

	var cancel context.CancelFunc
	c.Context, cancel = context.WithCancel(c.Context)
	go func() {
		sig := <-signals
		logger.Info("Received signal", "signal", sig)
		cancel()
	}()
}

func startCronJobs(c *cli.Context, _ *domain_handler.Handler) (func(), error) {
	logger := slogutils.FromContext(c.Context)

	cronJobs := cron.New()

	/*
		err := cronJobs.AddJob(
			c.String("delete-expired-access-tokens-cron"),
			domain_handler.CommandHandlerJob(c.Context, handler.AuthSessionDeleteExpired, command.AuthSessionDeleteExpiredCmd{}),
		)
		if err != nil {
			return nil, err
		}
	*/

	cronJobs.Start()

	return func() {
		logger.Debug("Stopping cron jobs")
		cronJobs.Stop()
		logger.Debug("All cron jobs stopped")
	}, nil
}
