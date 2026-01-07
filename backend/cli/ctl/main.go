package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/multitracer"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/joho/godotenv"
	"github.com/networkteam/devlog"
	"github.com/networkteam/slogutils"
	slogutils_tracelog "github.com/networkteam/slogutils/adapter/pgx/v5/tracelog"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/mail"
	"myvendor.mytld/myproject/backend/mail/smtp"
	"myvendor.mytld/myproject/backend/persistence/repository"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func main() {
	initialHandler := setupInitialLogging()
	loadDotenv()

	defaultConfig := domain.DefaultConfig()
	app := &cli.App{
		Name:  "ctl",
		Usage: "App CLI control",
		Flags: flattenFlags(
			[]cli.Flag{
				&cli.IntFlag{
					Name:    "verbosity",
					Usage:   "Verbosity: 0=fatal, 1=error, 2=warn, 3=info, 4=debug",
					Aliases: []string{"v"},
					EnvVars: []string{"BACKEND_VERBOSITY"},
					Value:   3,
				},
				&cli.StringFlag{
					Name:    "postgres-dsn",
					Usage:   "PostgreSQL connection DSN",
					Value:   "dbname=myproject-dev sslmode=disable",
					EnvVars: []string{"BACKEND_POSTGRES_DSN"},
				},
				&cli.DurationFlag{
					Name:    "db-wait-timeout",
					Usage:   "Duration to wait for database to become available (0 to disable)",
					Value:   15 * time.Second,
					EnvVars: []string{"BACKEND_DB_WAIT_TIMEOUT"},
				},

				&cli.IntFlag{
					Name:    "hash-cost",
					Usage:   "Hash cost for password hashing with bcrypt (between 4 and 31, higher is slower)",
					Value:   defaultConfig.HashCost,
					EnvVars: []string{"BACKEND_HASH_COST"},
				},

				&cli.StringFlag{
					Name:    "app-base-url",
					Usage:   "Application base URL",
					Value:   "http://localhost:3000/",
					EnvVars: []string{"BACKEND_APP_BASE_URL"},
				},

				&cli.StringFlag{
					Name:    "jwt-secret",
					Usage:   "JWT secret",
					EnvVars: []string{"BACKEND_JWT_SECRET"},
				},

				&cli.StringFlag{
					Name:    "smtp-host",
					Usage:   "Host of SMTP for outgoing mails",
					Value:   "localhost",
					EnvVars: []string{"BACKEND_SMTP_HOST"},
				},
				&cli.IntFlag{
					Name:    "smtp-port",
					Usage:   "SMTP Port for outgoing mails",
					Value:   1025,
					EnvVars: []string{"BACKEND_SMTP_PORT"},
				},
				&cli.StringFlag{
					Name:    "smtp-user",
					Usage:   "SMTP User for outgoing mails",
					EnvVars: []string{"BACKEND_SMTP_USER"},
				},
				&cli.StringFlag{
					Name:    "smtp-password",
					Usage:   "SMTP Password for outgoing mails",
					EnvVars: []string{"BACKEND_SMTP_PASSWORD"},
				},
				&cli.StringFlag{
					Name:    "smtp-tls-policy",
					Usage:   "TLS policy for outgoing mails (Values: opportunistic, mandatory, non)",
					EnvVars: []string{"BACKEND_SMTP_TLS_POLICY"},
					Value:   "non",
				},
				&cli.StringFlag{
					Name:    "mail-default-from",
					Usage:   "Default sender address for outgoing mails",
					EnvVars: []string{"MAIL_DEFAULT_FROM"},
					Value:   "app@example.com",
				},

				&cli.StringFlag{
					Name:    "test-time-delta",
					Usage:   "Time delta for testing (format: 1y2m3d4h). Examples: '1y' (1 year forward), '-2m' (2 months back), '30d12h' (30 days 12 hours forward)",
					Value:   "",
					EnvVars: []string{"TEST_TIME_DELTA"},
				},

				&cli.BoolFlag{
					Name:    "disable-ansi",
					Usage:   "Force disable ANSI log output and output log in logfmt format",
					EnvVars: []string{"BACKEND_DISABLE_ANSI"},
					Value:   false,
				},
				&cli.BoolFlag{
					Name:    "force-ansi",
					Usage:   "Force enable ANSI log output",
					EnvVars: []string{"BACKEND_FORCE_ANSI"},
					Value:   false,
				},
			},
			jobqueueMainFlags(),
		),
		Before: func(c *cli.Context) error {
			handler := setLogHandler(c)

			// There might be some log records already buffered, emit them now
			err := initialHandler.EmitTo(handler)
			if err != nil {
				slog.Warn("Error emitting initial logs", "component", "cli", slogutils.Err(err))
			}

			// Pretend the CLI has a SystemAdministrator role (without setting an account)
			c.Context = authentication.WithAuthContext(c.Context, authentication.AuthContext{
				Authenticated: true,
				Role:          types.RoleSystemAdministrator,
			})

			return nil
		},
		Commands: []*cli.Command{
			newServerCmd(),
			newWorkerCmd(),
			newMigrateCmd(),
			newAccountCmd(),
			newFixturesCmd(),
			newTestCmd(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		slog.Error("Error executing command", "component", "cli", slogutils.Err(err))
		os.Exit(1)
	}
}

func loadDotenv() {
	backendEnv := os.Getenv("BACKEND_ENV")
	if backendEnv == "" {
		backendEnv = "production"
	}

	// We load _all_ existing files for the app environment (development or production).
	// So we can override and set additional variables in *.local env files.

	filenames := []string{".env.local", ".env"}
	filenames = append([]string{fmt.Sprintf(".env.%s.local", backendEnv), fmt.Sprintf(".env.%s", backendEnv)}, filenames...)

	slog.Log(context.Background(), slogutils.LevelTrace, "Trying to load env from files", "component", "cli", "filenames", filenames)

	loadedEnvFiles := make([]string, 0, len(filenames))
	for _, filename := range filenames {
		err := godotenv.Load(filename)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			slog.Error("Error loading envs", "component", "cli", slogutils.Err(err), "filename", filename)
			continue
		}
		loadedEnvFiles = append(loadedEnvFiles, filename)
	}
	slog.Debug("Loaded env from files", "component", "cli", "filenames", loadedEnvFiles)
}

func connectDatabase(c *cli.Context, dlog *devlog.Instance) (*sql.DB, error) {
	postgresDSN := c.String("postgres-dsn")
	slog.Debug(
		"Connecting to database",
		"component", "cli",
		"postgresDSN", postgresDSN,
	)

	connConfig, err := pgx.ParseConfig(postgresDSN)
	if err != nil {
		return nil, errors.Wrap(err, "parsing PostgreSQL connection string")
	}
	verbosity := c.Int("verbosity")
	connConfig.Tracer = &tracelog.TraceLog{
		Logger: slogutils_tracelog.NewLogger(
			slog.Default().With("component", "db.driver"),
			slogutils_tracelog.WithRemapLevel(tracelog.LogLevelInfo, slogutils.LevelTrace),
			slogutils_tracelog.WithRemapErrorLevel(func(_ tracelog.LogLevel, err error) slog.Level {
				if repository.IsConstraintViolationError(err) {
					return slogutils.LevelTrace
				}
				return slog.LevelError
			}),
		),
		LogLevel: verbosityToTracelogLevel(verbosity),
	}
	if dlog != nil {
		connConfig.Tracer = multitracer.New(
			connConfig.Tracer,
			newDevlogQueryTracer(dlog.CollectDBQuery()),
		)
	}
	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "opening database connection")
	}

	waitTimeout := c.Duration("db-wait-timeout")
	if waitTimeout > 0 {
		err = waitForDatabase(c.Context, db, waitTimeout)
		if err != nil {
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

func waitForDatabase(ctx context.Context, db *sql.DB, timeout time.Duration) error {
	start := time.Now()
	deadline := start.Add(timeout)
	backoff := 100 * time.Millisecond

	slog.Info("Waiting for database to become available...", "component", "cli")

	for time.Now().Before(deadline) {
		err := db.PingContext(ctx)
		if err == nil {
			slog.Info("Database is available", "component", "cli", "duration", time.Since(start))
			return nil
		}
		time.Sleep(backoff)
		backoff = min(backoff*2, time.Second)
	}

	return errors.New("timeout waiting for database")
}

func buildMailer(c *cli.Context) (*mail.Mailer, error) {
	sender, err := smtp.NewSender(
		c.String("smtp-host"),
		c.Int("smtp-port"),
		c.String("smtp-user"),
		c.String("smtp-password"),
		c.String("smtp-tls-policy"),
	)
	if err != nil {
		return nil, err
	}
	defaultConfig, err := getConfig(c)
	if err != nil {
		return nil, err
	}
	config := mail.DefaultConfig(defaultConfig)
	config.DefaultFrom = c.String("mail-default-from")
	mailer := mail.NewMailer(sender, config)
	return mailer, nil
}

//nolint:unparam // parsing other flags could return an error
func getConfig(c *cli.Context) (domain.Config, error) {
	config := domain.DefaultConfig()
	config.AppBaseURL = c.String("app-base-url")
	config.HashCost = c.Int("hash-cost")
	config.JWTSecret = c.String("jwt-secret")
	// Add more config options here
	return config, nil
}
