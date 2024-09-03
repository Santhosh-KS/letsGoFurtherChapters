package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) server() error {

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Create a shutdownError channel. We will use this to receive any
	// errors returned by the graceful shutdown() function

	shutdownError := make(chan error)

	go func() {
		// Creae a quit channel which carries os.Signal values.

		quit := make(chan os.Signal, 1)
		// Use signal.Notify() to listen for incoming SIGINT and SIGERM signals and
		// relay them to the quit channel. Any other signals will
		// not be caught by signal.Notify() and will retian their default behavior.
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the signal from the quit channel. This code will block until
		// a signal is received.
		s := <-quit

		// Log a message to say that the signal has been caught. Notice that we also call the Stirng() method on the signal to get the signal name
		// include it in the log entry attributes.
		app.logger.Info("caught signal", "signal", s.String())

		// Create a context with a 30-second timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Call Shutdown() on our server, passing in the context we just made.
		// Shutdown() will return nil if the graceful shutdown was successful, or
		// an erro (which may happen because of a problem closing the listeners, or
		// because the shutdown didn't complete before the 30-second context deadline is
		// hit). We rely this return value to the shutdownError channel.
		shutdownError <- srv.Shutdown(ctx)

		/* // Exit the application with a 0 (success) status code.
		os.Exit(0) */
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	// Calling Shutdown() on our server will cause ListenAndServe() to immediately
	// return a http.ErrServerClosed error. So if we see this error, it is actually
	// a good thing and an indication that the graceful shutdown is started. So we check
	// specifically for this, only returing the error if it is NOT http.ErrServerClosed.

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, we wait to receive the return value from the Shutdown() on the
	// ShutdownError channel. If return value is an error, we know that there
	// was a problem with the graceful shutdown and we return the error.

	err = <-shutdownError
	if err != nil {
		return err
	}

	// At this point we know that graceful shutdown is completed and we log
	// a "stop server" message.

	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
