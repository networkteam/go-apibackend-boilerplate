package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/playground"
	logger "github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
	"github.com/friendsofgo/errors"
	"github.com/getsentry/sentry-go"
	"github.com/hashicorp/go-multierror"
	"github.com/mattn/go-isatty"
	"github.com/networkteam/apexlogutils"
	"github.com/networkteam/apexlogutils/httplog"
	apexlogutils_middleware "github.com/networkteam/apexlogutils/middleware"
	"github.com/robfig/cron"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/api"
	api_handler "myvendor.mytld/myproject/backend/api/handler"
	http_api "myvendor.mytld/myproject/backend/api/http"
)

const shutdownTimeout = 5 * time.Second

func newServerCmd() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Run the backend server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "address",
				Usage:   "Listen on this address",
				EnvVars: []string{"BACKEND_ADDRESS"},
				Value:   "0.0.0.0:8080",
			},
			&cli.BoolFlag{
				Name:  "playground",
				Usage: "Enable GraphQL playground",
				Value: false,
			},
			&cli.BoolFlag{
				Name:    "disable-ansi",
				Usage:   "Force disable ANSI log output and output log in logfmt format",
				EnvVars: []string{"BACKEND_DISABLE_ANSI"},
				Value:   false,
			},

			&cli.StringFlag{
				Name:    "sentry-dsn",
				Usage:   "Sentry DSN (will be disabled if empty)",
				EnvVars: []string{"SENTRY_DSN"},
			},
			&cli.StringFlag{
				Name:    "sentry-environment",
				Usage:   "Sentry environment",
				EnvVars: []string{"SENTRY_ENVIRONMENT"},
				Value:   "development",
			},
			&cli.StringFlag{
				Name:    "sentry-release",
				Usage:   "Release version for Sentry",
				EnvVars: []string{"SENTRY_RELEASE"},
			},
		},
		Before: func(c *cli.Context) error {
			setServerLogHandler(c)

			return nil
		},
		Action: func(c *cli.Context) error {
			// This action is where the server is set up and dependencies are wired
			// -- make sure to keep it clean and with clear intention what is done here

			log := logger.FromContext(c.Context)

			// Initialize sentry
			defer sentry.Recover()
			err := initializeSentry(c, "backend")
			if err != nil {
				return err
			}

			db, err := connectDatabase(c)
			if err != nil {
				return err
			}

			err = db.Ping()
			if err != nil {
				return errors.Wrap(err, "pinging database")
			}

			mailer := buildMailer(c)

			timeSource, err := newCurrentTimeSource(c)
			if err != nil {
				return err
			}

			// Set up signal handling, should be called before starting background processing
			setupCancelOnSignal(c)

			shutdownCronJobs, err := startCronJobs(c, db)
			if err != nil {
				return err
			}

			mux := http.NewServeMux()

			deps := api.ResolverDependencies{
				DB:         db,
				TimeSource: timeSource,
				Mailer:     mailer,
				Config:     getConfig(c),
			}
			graphqlHandler := api_handler.NewGraphqlHandler(deps, api_handler.HandlerConfig{
				EnableTracing:  false,
				EnableLogging:  true,
				DisableRecover: false,
			})

			playgroundEnabled := c.Bool("playground")
			if playgroundEnabled {
				mux.Handle("/", playground.Handler("GraphQL playground", "/query"))
			}

			mux.Handle("/query", http_api.MiddlewareStack(deps, graphqlHandler))
			mux.HandleFunc("/healthz", api_handler.NewHealthzHandler(db))

			rootHandler := apexlogutils_middleware.RequestID(
				httplog.New(
					mux,
					// Do not log health checks, it would be too verbose
					httplog.ExcludePathPrefix("/healthz"),
				),
			)

			address := c.String("address")
			log.Infof("Serving GraphQL endpoint at http://%s/query", address)
			if playgroundEnabled {
				log.Infof("Connects to http://%s/ for GraphQL playground", address)
			}

			return serve(c, rootHandler, func(c *cli.Context) error {
				shutdownCronJobs()
				return nil
			})
		},
	}
}

func serve(c *cli.Context, handler http.Handler, onShutdown func(c *cli.Context) error) (err error) {
	log := logger.FromContext(c.Context)

	address := c.String("address")
	srv := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// Fatal will exit the program if the server failed to listen
			log.
				WithError(err).
				Fatalf("Failed to listen and serve")
		}
	}()

	<-c.Context.Done()

	log.Debugf("Server stopped")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer func() {
		cancel()
	}()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		return errors.Wrap(err, "shutting down server")
	}

	log.Debugf("Server exited properly")

	if err == http.ErrServerClosed {
		err = nil
	}

	if shutdownErr := onShutdown(c); shutdownErr != nil {
		err = multierror.Append(err, shutdownErr)
	}

	log.Infof("Everything shut down, goodbye")

	return err
}

func setupCancelOnSignal(c *cli.Context) {
	log := logger.FromContext(c.Context)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch,
		// kill -SIGINT XXXX or Ctrl+c
		syscall.SIGINT,
		// kill -SIGTERM XXXX
		syscall.SIGTERM,
	)

	var cancel context.CancelFunc
	c.Context, cancel = context.WithCancel(c.Context)
	go func() {
		sig := <-ch
		log.Infof("Received signal: %v", sig)
		cancel()
	}()
}

func startCronJobs(c *cli.Context, db *sql.DB) (func(), error) {
	log := logger.FromContext(c.Context)

	cronJobs := cron.New()

	cronJobs.Start()

	return func() {
		log.Debugf("Stopping cron jobs")
		cronJobs.Stop()
		log.Debugf("All cron jobs stopped")
	}, nil
}

func initializeSentry(c *cli.Context, component string) error {
	log := logger.FromContext(c.Context)

	sentryDSN := c.String("sentry-dsn")
	sentryEnvironment := c.String("sentry-environment")
	sentryRelease := c.String("sentry-release")

	if sentryDSN == "" {
		log.Info("No Sentry DSN set: Sentry disabled")

		return nil
	}

	sentryOptions := sentry.ClientOptions{
		Dsn:         sentryDSN,
		Environment: sentryEnvironment,
		Release:     sentryRelease,
		DebugWriter: os.Stderr,
	}

	log.
		WithField("dsn", sentryDSN).
		WithField("environment", sentryEnvironment).
		WithField("release", sentryRelease).
		Info("Initializing Sentry")

	if sentryEnvironment != "production" {
		sentryOptions.Debug = true
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTags(map[string]string{"component": component})
	})

	err := sentry.Init(sentryOptions)
	if err != nil {
		return errors.Wrap(err, "initializing Sentry")
	}

	return nil
}

func setServerLogHandler(c *cli.Context) {
	if isatty.IsTerminal(os.Stdout.Fd()) && !c.Bool("disable-ansi") {
		logger.SetHandler(apexlogutils.NewComponentTextHandler(os.Stderr))
	} else {
		logger.SetHandler(logfmt.New(os.Stderr))
	}

	// Use a logger instance with predeclared component field
	log := logger.WithField("component", "cli.server")
	// Add logger to context.Context of cli.Context, so individual
	c.Context = logger.NewContext(c.Context, log)
}
