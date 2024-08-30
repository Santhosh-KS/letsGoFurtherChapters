package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deffered function (which will always be run in the event of panic
		// as go unwinds the stack)
		defer func() {
			// use the builtin recover function to check fi there has been a panic or
			// not

			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a
				// response has been sent.

				w.Header().Set("Connection", "close")

				// The value retured by recover() has the type any,
				// so we use fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using our custom logger type at ERROR level and send the client a 500 iinternal server error response.

				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Initialize a new rate limiter which allows an average of 2 requests per second,
	// with a maximum of 4 requests in a single 'burst'.

	limiter := rate.NewLimiter(2, 4)

	// The function we are returing is a closure, which 'closes over' the limiter variable.

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call limiter.Allow() to see if the request is permitted, and if its not
		// then we call the rateLimitExceeded Response() helper to return a 429 Too many
		// Requests response
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
