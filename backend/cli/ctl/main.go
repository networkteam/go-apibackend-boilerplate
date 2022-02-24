package main

import (
	"database/sql"
	"os"

	"github.com/apex/log"
	cli_handler "github.com/apex/log/handlers/cli"
	"github.com/friendsofgo/errors"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/networkteam/apexlogutils"
	apexlogutils_pgx "github.com/networkteam/apexlogutils/pgx"
	"github.com/urfave/cli/v2"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/mail"
	"myvendor.mytld/myproject/backend/mail/smtp"
	"myvendor.mytld/myproject/backend/security/authentication"
)

func main() {
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
		},
		Before: func(c *cli.Context) error {
			verbosity := apexlogutils.Verbosity(c.Int("verbosity"))
			log.SetLevel(apexlogutils.ToApexLogLevel(verbosity))
			// Use a CLI friendly handler by default, server sets its own handler depending on terminal / ANSI
			log.SetHandler(cli_handler.New(os.Stderr))

			// Pretend the CLI has a SystemAdministrator role (without setting an account)
			c.Context = authentication.WithAuthContext(c.Context, authentication.AuthContext{
				Authenticated: true,
				Role:          domain.RoleSystemAdministrator,
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
	connConfig.Logger = apexlogutils_pgx.NewLogger(log.Log)
	verbosity := apexlogutils.Verbosity(c.Int("verbosity"))
	connConfig.LogLevel = apexlogutils_pgx.ToPgxLogLevel(verbosity)
	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "opening database connection")
	}
	return db, nil
}

func buildMailer(c *cli.Context) *mail.Mailer {
	sender := smtp.NewSender(
		c.String("smtp-host"),
		c.Int("smtp-port"),
		c.String("smtp-user"),
		c.String("smtp-password"),
	)
	config := mail.DefaultConfig(getConfig(c))
	mailer := mail.NewMailer(sender, config)
	return mailer
}

func getConfig(c *cli.Context) domain.Config {
	config := domain.DefaultConfig()
	config.AppBaseUrl = c.String("app-base-url")
	return config
}
