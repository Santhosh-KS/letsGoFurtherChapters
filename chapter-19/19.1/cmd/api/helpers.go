package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"greelight.techkunstler.com/internal/validator"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	// Use http.MaxBytesReader() to limit the size of the request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder, and call the DisallowUnknowFields() method on it
	// before decoding. This means that if the JSON from the client now includes any
	// field which cannot be mapped to the target destination, the decoder will return an error
	// instead of just ignoring the field

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the target destination.
	err := dec.Decode(dst)

	if err != nil {
		// If there is an error during decoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		// Add a ndwmaxBytesError variable.
		var maxBytesError *http.MaxBytesError

		switch {
		// Use the errors.As() function to check whether the error has the type
		// *json.SyntaxErro. If it does, then return a plain-english error message
		// which includes the ocation of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %x)", syntaxError)
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

		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format "json: unknown
		// field "<name>"". We can check for this . extract the field name from the error,
		// and interpolate it into our custom error message. Note that there's an open
		// issue at https://github.com//golang/go/issues/29035 regarding turning this
		// into a distinct error type in the future.

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// Use the errors.As() function to check whether the error has the type
		// *http.MaxBytesError. If it does, then it means the request body exceed our
		//size limit of 1MB and we return  aclear error message.
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

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
	// Way to handle multiple json values passed:
	// Example: curl -d '{"title": "Moana"}{"title": "Top Gun"}' localhost:4000/v1/movies
	// Call Decode() again, using a pointer to an empty anonymous struct as the destination. If the request body only contained a single JSON value this
	// will retun an io.EOF error. So if we get anything else, we know that there
	// is additional data in the request body and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON object")
	}
	return nil
}

// The readString() helper returns a string value from the query string, or fht provided
// default value if no matching key could be found.

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// Extract the value for a given key from the query string. If no key exists this
	// will return an empty string ""

	s := qs.Get(key)

	// If no key exists (or the value is empty) then return the default value.

	if s == "" {
		return defaultValue
	}

	return s
}

// The readCSV() helper reads a string value from the query string and then splits it
// into a slice on the comma character. If no matching key could be found,
// it returns the provided default value.

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// Extract the value from the query string.

	csv := qs.Get(key)

	// If no key exists (or the value is empty) then return the default value.

	if csv == "" {
		return defaultValue
	}

	// Otherwise parse the value into a []string slice and return it

	return strings.Split(csv, ",")
}

// The readInt() helper reads a string value from the query string and converts it into an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record an
// error message in the provided Validator instance.

func (app *application) readInt(qs url.Values,
	key string, defaultValue int, v *validator.Validator) int {
	// Extract the value from the query string.

	s := qs.Get(key)

	// If no key exists (or the value is empty) then return the default value.

	if s == "" {
		return defaultValue
	}

	// Try to convert the value to an int. If this fails, add an error mesage to the
	// validator instance and return the default value.

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *application) background(fn func()) {
	// Launch a background goroutine.
	app.wg.Add(1)

	go func() {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()
		fn()
	}()
}
