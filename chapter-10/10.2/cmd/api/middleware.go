package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
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

	// Define a client struct to hold the rate limiter and last seen time for each clinet.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	//  Declare a mutex and a map to hold clien'ts IP addresses and
	// rate limiters
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries from the clients map
	// once every minute.

	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutext to prevent any rate limiter check from happening
			// while the cleanup is in progress.

			mu.Lock()

			// Loop through all clients. If they haven't been seen within the last
			// three minutes, delete the corresponding entry from the map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	// The function we are returing is a closure, which 'closes over' the limiter variable. */

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Extract the client's IP add from the request.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Lock the mutext to prevent this code from being executed concurrently.
		mu.Lock()

		// Check to see if the IP address already exists in the map.
		// If it doesn't, then initialize a new rate limiter and
		// add the IP add and limiter to the map.

		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}

		}

		clients[ip].lastSeen = time.Now()
		// Call limiter.Allow() to see if the request is permitted, and if its not
		// then we call the rateLimitExceeded Response() helper to return a 429 Too many
		// Requests response
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		// Very importatntly, unlock the mutext before calling the next handler in
		// the chain. Notice that we DON't use defer to unlock the mutex, as that
		// would mean that mutext isn't unlocked until all the handlers downstream of this
		// middleware have also returned.
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
