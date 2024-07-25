package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

// Define a config struct to hold all the configuration settins for our application.
// For now, the only configuration settings will be the network port that we want the server to listen on,
// and the current operationg environmnet for the application (development, staging, production, etc.)
// We will read in these configuration settings from command-line flags when the application starts.

type config struct {
	port int
	env  string
}

// Define an application struct o hold the dependencies for our HTTP handlers, helpers, and middleware. At the moment this only contains a copy of the config struct and a logger, but it will grwo to include a lot more our build progresses.

type application struct {
	config config
	logger *slog.Logger
}

func main() {

	// Declarean instance of the config struct.

	var cfg config

	// Read tthe value of the port and env command-line flags into the config struct. We default to using the port number 4000 and the environment "development" if no corresponding flags are provided.

	flag.IntVar(&cfg.port, "port", 4000, "API Server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()
	// Initialize a new structured logger which writes log entries to the standard out stream
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := &application{
		config: cfg,
		logger: logger,
	}
	// Declare a new servemux and add a /v1/healthcheck route which dispatches reques to the healthcheckhandler method (which we will create in a moment).
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// Declare a HTTP server which listens on the port provided in the config struct, uses the
	// servemux we created above as the handler, has some sensible timeout settings and writes any log messages to the structured logger at error level.

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// Start the HTTP server.

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.env)
	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)

}
