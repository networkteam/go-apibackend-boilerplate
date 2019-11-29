package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/handler"
	"github.com/apex/log"
	"github.com/apex/log/handlers/logfmt"
	"github.com/apex/log/handlers/text"
	"github.com/friendsofgo/errors"
	sentry "github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	isatty "github.com/mattn/go-isatty"
	"github.com/robfig/cron"
	"github.com/spf13/cobra"

	"myvendor.mytld/myproject/backend/api"
	api_handler "myvendor.mytld/myproject/backend/api/handler"
	domain_handler "myvendor.mytld/myproject/backend/handler"
	"myvendor.mytld/myproject/backend/service/hub"
	"myvendor.mytld/myproject/backend/service/notification"
)

var serverFlags struct {
	port                              int
	enableTracing                     bool
	enablePlayground                  bool
	goRushApiUrl                      string
	appAccountTokenCleanupJobInterval string
	expiredCheckTasksJobInterval      string
	scheduledCheckTasksJobInterval    string
	localesDir                        string
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVar(&serverFlags.port, "port", 8080, "Port for the server")
	serverCmd.Flags().BoolVar(&serverFlags.enableTracing, "trace", false, "Enable Apollo tracing for GraphQL")
	serverCmd.Flags().BoolVar(&serverFlags.enablePlayground, "playground", false, "Enable GraphQL Playground")
	serverCmd.Flags().StringVar(&serverFlags.goRushApiUrl, "gorush-api-url", "http://localhost:8088/api", "Api Url for GoRush PushNotification Service")
	serverCmd.Flags().StringVar(&serverFlags.localesDir, "locales-dir", "./locales", "Path to locales directory")
	serverCmd.Flags().StringVar(&serverFlags.appAccountTokenCleanupJobInterval, "app-account-token-cleanup-job-interval", "1m", "Cron interval for running the app account token cleanup job")
	serverCmd.Flags().StringVar(&serverFlags.expiredCheckTasksJobInterval, "expired-check-tasks-job-interval", "1m", "Cron interval for checking expired check tasks job")
	serverCmd.Flags().StringVar(&serverFlags.scheduledCheckTasksJobInterval, "scheduled-check-tasks-job-interval", "1m", "Cron interval for checking scheduled check tasks job")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Runs the API server",
	PreRun: func(_ *cobra.Command, _ []string) {
		setLogHandler()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize sentry
		defer sentry.Recover()
		initializeSentry("backend")

		sentryHandler := sentryhttp.New(sentryhttp.Options{})

		// Do not display usage on errors after arguments are validated
		// See https://github.com/spf13/cobra/issues/340
		cmd.SilenceUsage = true

		timeSource := newCurrentTimeSource()

		pushNotificationService := notification.NewPushNotificationService(serverFlags.goRushApiUrl)

		h := hub.NewHub()

		graphqlHandler := api_handler.NewGraphqlHandler(
			api.ResolverDependencies{
				Db:         rootCtx.db,
				TimeSource: timeSource,
				Hub:        h,
				Notifier:   pushNotificationService,
			},
			api_handler.HandlerConfig{
				EnableTracing: serverFlags.enableTracing,
				EnableLogging: rootFlags.verbosity > 2,
			},
		)

		if serverFlags.enablePlayground {
			http.Handle("/", handler.Playground("GraphQL playground", "/query"))
		}

		http.Handle("/query", graphqlHandler)
		http.HandleFunc("/healthz", sentryHandler.HandleFunc(api_handler.NewHealthzHandler(rootCtx.db)))

		cronJob := cron.New()

		appAccountTokenCleanupJob := domain_handler.NewAppAccountTokenCleanupJob(rootCtx.db, timeSource)
		if err := cronJob.AddJob(fmt.Sprintf("@every %s", serverFlags.appAccountTokenCleanupJobInterval), appAccountTokenCleanupJob); err != nil {
			return errors.Wrap(err, "adding cleanup job")
		}

		cronJob.Start()

		log.Infof("Serving GraphQL endpoint at http://localhost:%d/query", serverFlags.port)
		if serverFlags.enablePlayground {
			log.Infof("Connect to http://localhost:%d/ for GraphQL playground", serverFlags.port)
		}

		return http.ListenAndServe(fmt.Sprintf(":%d", serverFlags.port), nil)
	},
}

func initializeSentry(component string) {
	sentryEnvironment := os.Getenv("SENTRY_ENVIRONMENT")
	sentryDSN := os.Getenv("SENTRY_DSN")
	sentryRelease := os.Getenv("SENTRY_RELEASE")

	sentryOptions := sentry.ClientOptions{
		Dsn:         sentryDSN,
		Environment: sentryEnvironment,
		Release:     sentryRelease,
		DebugWriter: os.Stderr,
	}

	if sentryDSN == "" {
		log.Info("No SENTRY_DSN set. Sentry disabled.")

		return
	}

	log.WithField("SENTRY_ENVIRONMENT", sentryEnvironment).Info("Using")
	log.WithField("SENTRY_DSN", sentryDSN).Info("Using")
	log.WithField("SENTRY_RELEASE", os.Getenv("SENTRY_RELEASE")).Info("Using")

	if sentryEnvironment != "production" {
		sentryOptions.Debug = true
	}

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTags(map[string]string{"component": component})
	})

	err := sentry.Init(sentryOptions)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		log.WithError(err).Error("Could not initialize Sentry")
		os.Exit(1)
	}
}

func setLogHandler() {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		log.SetHandler(text.New(os.Stderr))
	} else {
		log.SetHandler(logfmt.New(os.Stderr))
	}
}
