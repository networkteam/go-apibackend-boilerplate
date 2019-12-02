package main

import (
	"context"
	"database/sql"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/friendsofgo/errors"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/security/authentication"
)

var rootFlags struct {
	postgresDSN string
	verbosity   int
}

var rootCtx struct {
	db  *sql.DB
	ctx context.Context
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootFlags.postgresDSN, "postgres-dsn", "dbname=myproject-dev sslmode=disable", "PostgreSQL connection DSN")
	rootCmd.PersistentFlags().IntVarP(&rootFlags.verbosity, "verbosity", "v", 3, "Verbosity: 0=fatal, 1=error, 2=warn, 3=info, 4=debug")
}

var rootCmd = &cobra.Command{
	Use:   "myproject-ctl",
	Short: "CLI control for myproject",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		log.SetLevel(logLevel(rootFlags.verbosity))
		log.SetHandler(cli.New(os.Stderr))
		if os.Getenv("POSTGRES_DSN") != "" {
			rootFlags.postgresDSN = os.Getenv("POSTGRES_DSN")
		}

		log.WithField("postgresDSN", rootFlags.postgresDSN).Debug("Connecting to db")
		db, err := sql.Open("postgres", rootFlags.postgresDSN)
		if err != nil {
			return errors.Wrap(err, "could not open db")
		}
		if err := db.Ping(); err != nil {
			return errors.Wrap(err, "could not connect to db")
		}

		rootCtx.db = db
		rootCtx.ctx = buildCtx()

		return nil
	},
}

func logLevel(verbosity int) log.Level {
	if verbosity >= 4 {
		return log.DebugLevel
	}

	switch verbosity {
	case 3:
		return log.InfoLevel
	case 2:
		return log.WarnLevel
	case 1:
		return log.ErrorLevel
	}

	return log.FatalLevel
}

func buildCtx() context.Context {
	ctx := context.Background()
	authCtx := authentication.AuthContext{
		Authenticated: true,
		Role:          domain.RoleSystemAdministrator,
	}
	return authentication.WithAuthContext(ctx, authCtx)
}
