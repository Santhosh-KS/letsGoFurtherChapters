package main

import (
	// "encoding/json"
	"fmt"
	"greelight.techkunstler.com/internal/data"
	"net/http"
	"time"
)

// Add a createMovieHandler for the "POST /v1/movies" endpoint. For now we simply
// return a plain-text placeholder response.

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Decrae an anonymus struct to hold the information that we expect to be in the
	// HTTP request body (note that the field names and types in the struct are subset
	//of the move sturct that we created earlier).
	// This struct will be our *target decode destination*

	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"` // Make this field a data.Runtime type.
		Genres  []string     `json:"genres`
	}

	/* // Initialize a new json.Decoder instance which reads from the request body, and then use the Decode() method to decode the body contents in to the input struct.
	// Importantly, notice that when we call Decode() we pass a *pointer* to the input struct as the target decode destination. If there was an error during decoding, we
	// also use our generic errorResponse() helper to send tht client a 400 Bad Request respone
	// containing the error message.

	err := json.NewDecoder(r.Body).Decode(&input)

	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	// Dump the contents of the input struct in a HTTP response. */
	err := app.readJSON(w, r, &input)
	if err != nil {
		// app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		app.badRequestResponse(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", input)
}

// Add a showMoivew handler for the "GET /v1/movies/:id" endpoint. For now, we retrive the
// the interpolated "id" parameter from the current URL and include it in a placeholder response.

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		// http.NotFound(w, r)
		app.notFoundResponse(w, r)
		return
	}

	// Create a new instance of the Movie struct, containing the ID we extracted from
	// the URL and some dummy data. Also notice thta we deliberately haven't set a
	// value for the Year field

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	// Encode the struct to JSON and send it as the HTTP response.
	// Create an envelope {"movie": movie} instance and pass it to writeJSON(), instead of
	// passing the plain movie struct.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		/* app.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError) */
		app.serverErrorResponse(w, r, err)
	}

}
