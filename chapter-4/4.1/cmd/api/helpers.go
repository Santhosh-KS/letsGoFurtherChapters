package main

import (
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

// Define an envelope type.
type envelope map[string]any

func (app *application) readIDParam(r *http.Request) (int64, error) {

	// When httprouter is parsing a request, any interpolated URL parameters will be stored
	// in the request context. We can use the ParamsFromContext() function to retrive al slice
	// containing these parameter names and values.

	params := httprouter.ParamsFromContext(r.Context())

	// We can then use the ByName() method to get the value of the "id" parameter from the slice. In our project all movies will have a unique positive integer ID, but the value returned by ByName() is always a string. So we try to convert it to a base 10 integer (with
	// a bit size of 64) if the parameter couldn't be converted, or is less than 1, we know the ID is invalid so we use the http.NotFound() function to return a 404 NotFound response.

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil

}

// Change the data parameter to have the type envelope instead of any.
func (app *application) writeJSON(w http.ResponseWriter,
	status int, data envelope, headers http.Header) error {

	// Pass the mat to the json.Marshal() function. This returns a []byte slice
	// containing the encoded JSON. If there was an error, we log it and send the
	// client a generic error message.

	// Use the json.MarshalIndent() function so that whitespace is added to the encoded.
	// JSON. Here we use no line prefix ("") and tab indents ("\t") for each element.
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w,
			"The server encountered a problem and could not process your request",
			http.StatusInternalServerError)
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	// At this point we know that encoding the data worked without any problems, so we
	// can safely set any necessary HTTP headers for a successful response.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}
