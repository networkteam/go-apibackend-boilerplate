package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	cli_handler "github.com/apex/log/handlers/cli"
	"github.com/friendsofgo/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/joho/godotenv"
	"github.com/networkteam/apexlogutils"
	apexlogutils_pgx "github.com/networkteam/apexlogutils/pgx/v5"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/domain/types"
	"myvendor.mytld/myproject/backend/mail"
	"myvendor.mytld/myproject/backend/mail/smtp"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func main() {
	loadDotenv()

	defaultConfig := domain.DefaultConfig()
	app := &cli.App{
		Name:  "ctl",
		Usage: "App CLI control",
		Flags: []cli.Flag{
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
		},
		Before: func(c *cli.Context) error {
			verbosity := apexlogutils.Verbosity(c.Int("verbosity"))
			log.SetLevel(apexlogutils.ToApexLogLevel(verbosity))
			// Use a CLI friendly handler by default, server sets its own handler depending on terminal / ANSI
			log.SetHandler(cli_handler.New(os.Stderr))

			// Pretend the CLI has a SystemAdministrator role (without setting an account)
			c.Context = authentication.WithAuthContext(c.Context, authentication.AuthContext{
				Authenticated: true,
				Role:          types.RoleSystemAdministrator,
			})

			return nil
		},
		Commands: []*cli.Command{
			newServerCmd(),
			newMigrateCmd(),
			newAccountCmd(),
			newFixturesCmd(),
			newTestCmd(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.
			WithField("component", "cli").
			Fatalf("Error executing command: %v", err)
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

	log.
		WithField("component", "cli").
		Infof("Trying to load env from %v", strings.Join(filenames, ", "))

	for _, filename := range filenames {
		err := godotenv.Load(filename)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			log.
				WithField("component", "cli").
				WithError(err).
				Fatalf("Error loading envs from %s", filename)
		}
		log.
			WithField("component", "cli").
			Infof("Loaded env from %s", filename)
	}
}

func connectDatabase(c *cli.Context) (*sql.DB, error) {
	postgresDSN := c.String("postgres-dsn")
	log.
		WithField("component", "cli").
		WithField("postgresDSN", postgresDSN).
		Debug("Connecting to database")

	connConfig, err := pgx.ParseConfig(postgresDSN)
	if err != nil {
		return nil, errors.Wrap(err, "parsing PostgreSQL connection string")
	}
	verbosity := apexlogutils.Verbosity(c.Int("verbosity"))
	connConfig.Tracer = &tracelog.TraceLog{
		Logger:   apexlogutils_pgx.NewLogger(log.Log),
		LogLevel: apexlogutils_pgx.ToPgxLogLevel(verbosity),
	}
	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "opening database connection")
	}
	return db, nil
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
	// Add more config options here
	return config, nil
}
