package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
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

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Decode the request body into the target destination.

	err := json.NewDecoder(r.Body).Decode(dst)

	if err != nil {
		// If there is an error during decoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// Use the errors.As() function to check whether the error has the type
		// *json.SyntaxErro. If it does, then return a plain-english error message
		// which includes the ocation of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError)
			// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
			// for syntax errors in the JSON. So we check for this using errors.Is() and
			// return a generic error message. There is an open issue regarding this a
			// https://github.com/golang/go/issues/25956
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

			// Likewise, catch any *json.UnmarshalTypeError errors. These occour when the
			// JSON value is the wrong type for the target destination. If the eeor relates to a
			// specific field, then we include that in our error message to mae it
			// easier for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)",
				unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty. We
		// check for this with errors.Is() and return a plain-english error message instead.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// A json.InvalidUnmarshallError error will be returned if we pass something
		// that is not a non-nil pointer to Decode(). We catch this and panic,
		// rather than returning an error to ur handler. At the end of this chpater
		// we will talk about panicking versus returing errors, and discuss why it's an
		// appropriate thing to do in this specific situation.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		// for anything else, return the error message as-is
		default:
			return err
		}
	}
	return nil
}
