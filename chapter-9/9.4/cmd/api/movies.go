package main

import (
	// "encoding/json"
	"errors"
	"fmt"
	"greelight.techkunstler.com/internal/data"
	"greelight.techkunstler.com/internal/validator"
	"net/http"
	// "time"
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
		Genres  []string     `json:"genres"`
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

	// Copy the values from the input struct to a new Movie struct.
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Initialize a new Validator instance.
	v := validator.New()

	// Use the Valid() method to see if any of the checks failed. If they did, then use the failedValidationResponse() helper to send a response to the clien,
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our movies model, passing in a pointer to the validated movie struct.
	// Thiswill create a recor in the database and update the
	// movie struct with the system-generated information.
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending a HTTP response, we want to include a location header to let the client know which URL they can find the newly created resource at. We make an empty
	// http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Write a JSON response with a 201 Created status code, themovie data in the
	// response body, and the location header.

	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// passing in the v.Errors map.
	fmt.Fprintf(w, "%+v\n", input)
}

// Add a showMoivew handler for the "GET /v1/movies/:id" endpoint. For now, we retrive the
// the interpolated "id" parameter from the current URL and include it in a placeholder response.

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific movie. We also need to use
	// the errors.Is() function to check if it returns a data.ErrRecordNotFound error,
	// in which case we send a 404 Not Found response to the client.

	movie, err := app.models.Movies.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	/* // Create a new instance of the Movie struct, containing the ID we extracted from
	// the URL and some dummy data. Also notice that we deliberately haven't set a
	// value for the Year field

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	} */

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

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the existing movie record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)

		}
		return
	}

	// Use pointers for the Title, year and Runtime fields.
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Read the JSON request body data into the input struct.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// If the input.Title Value is nil then we know that no corresponding "title" key
	// value pair was provided in the JSON request body. So we move on and leave the movie
	// record unchanged. Otherwise, we update the movie record with the new title
	// value. Importantly, because input.Title is a now a pointer to a string, we need to dereference the pointer using
	// * operator to get the underlying value, before assigning it to our movie record.

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// Validate the updted movie record, ending the client a 422 Unprocessable Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	/* // Pass the updated movie record to our new Update() method.
	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	} */

	// Intercept any errEditConflict error and call the new editConflictResponse() helper
	err = app.models.Movies.Update(movie)
	if err != nil {

		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write the updated movie record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete the movie from the database, sending a 404 not found respons
	// to the client if there isn't a matching record.

	err = app.models.Movies.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status code along with a success message.

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// To keep things consistent with our other handlers, we'll define an input struct
	// to hold the expected values from the request query string.
	// Embed the new Filters struct
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	// Initialize a new Validator instance.

	v := validator.New()

	// Cal r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()

	// Use our helpers to extract the title and genres query string values, falling back
	// to defaults of an empty string and empty slice respectively. If they are not
	// provided by theclient.

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// Get the page and page_size query string values as integers. notice that we set the default
	// page value to 1 and default page_size to 20, and that we pass the
	// validator instance as the final argument here.

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Extract the sort query string value, falling back to "id" if it is not
	// provided. by the client (which will impy an ascending sort on movie ID).

	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafeList = []string{"id", "title", "year",
		"runtime", "-id", "-title",
		"-year", "-runtime"}
	// Check the validator instance for any errors and use the failedValidationResponse()
	// helper to send the client a response if necessary.
	// Execute the validateion checks on the Filters struct and send a response containing the errors if necessary
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the GetAll() method to retrievethe movies, passing in the various filter
	// parameters.

	movies, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the move data.
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// Dump the contents of the input struct in a HTTP response.

	fmt.Fprintf(w, "%+v\n", input)
}
