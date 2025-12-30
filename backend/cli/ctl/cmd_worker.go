package main

import (
	"net/http"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"
	"github.com/urfave/cli/v2"

	api_handler "myvendor.mytld/myproject/backend/api/handler"
	"myvendor.mytld/myproject/backend/domain/handler"
)

func newWorkerCmd() *cli.Command {
	return &cli.Command{
		Name:  "worker",
		Usage: "Run a worker to process job queue tasks",
		Flags: flattenFlags(
			jobqueueWorkerFlags(),
			sentryFlags(),
		),
		Action: workerAction,
	}
}

func workerAction(c *cli.Context) error {
	logger := slogutils.FromContext(c.Context)

	err := initializeSentry(c, "worker")
	if err != nil {
		return err
	}

	if c.Bool("sentry-log-integration") {
		logger = initializeSentrySlog(c)
	}

	logger = setLogProcess(c, "worker")

	db, err := connectDatabase(c, nil)
	if err != nil {
		return err
	}

	mailer, err := buildMailer(c)
	if err != nil {
		return err
	}

	config, err := getConfig(c)
	if err != nil {
		return err
	}

	timeSource, err := newCurrentTimeSource(c)
	if err != nil {
		return err
	}

	queue := createAsynqQueue(c)

	domainHandler := handler.NewHandler(db, config, handler.Deps{
		TimeSource: timeSource,
		Mailer:     mailer,
		Queue:      queue,
	})

	srv := createAsynqServer(c, domainHandler)
	err = srv.Start()
	if err != nil {
		return errors.Wrap(err, "starting server")
	}

	// Health check HTTP server
	outerMux := http.NewServeMux()
	outerMux.HandleFunc("/healthz", api_handler.NewHealthzHandler(db))
	httpServer := &http.Server{
		Addr:              ":8070",
		Handler:           outerMux,
		ReadHeaderTimeout: 60 * time.Second,
	}
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server error", slogutils.Err(err))
		}
	}()

	logger.Info("Worker started, waiting for tasks")

	// Wait for shutdown and gracefully stop
	srv.WaitForSignals()
	err = httpServer.Shutdown(c.Context)
	if err != nil {
		logger.Error("HTTP server shutdown error", slogutils.Err(err))
	}
	srv.Shutdown()

	logger.Info("Worker shut down")

	return nil
}
