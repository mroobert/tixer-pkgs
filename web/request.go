package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/julienschmidt/httprouter"
)

var (
	ErrIdTooLong           = errors.New("ID is too long")
	ErrIdInvalidUTF8       = errors.New("ID contains invalid UTF-8 characters")
	ErrIdProhibitedChar    = errors.New("ID contains prohibited character")
	ErrIdProhibitedPattern = errors.New("ID matches prohibited pattern")
)

// ReadIDParam standardizes the way we read & validate url ID parameters across the Tixer services.
// The validation rules are relative to firestore.
func ReadIDParam(r *http.Request) (string, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id := params.ByName("id")

	if len(id) > 1500 {
		return "", fmt.Errorf(" %q: %w", id, ErrIdTooLong)
	}

	if !utf8.ValidString(id) {
		return "", fmt.Errorf(" %q: %w", id, ErrIdInvalidUTF8)
	}

	if strings.Contains(id, "/") || strings.Contains(id, "\\") {
		return "", fmt.Errorf(" %q: %w", id, ErrIdProhibitedChar)
	}

	if id == "." || id == ".." || strings.HasPrefix(id, "__") || strings.HasSuffix(id, "__") {
		return "", fmt.Errorf(" %q: %w", id, ErrIdProhibitedPattern)
	}

	return id, nil
}

// ReadJSON is a helper function to standardize the way we read JSON request data across the Tixer services.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {

		// Return the location of the parsing problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
		// for syntax errors in the JSON. So we check for this using errors.Is() and
		// return a generic error message. There is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// These occur when the JSON value is the wrong type for the target destination.
		// If the error relates to a specific field, then we include that in our error message to make it
		// easier for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		// An io.EOF error will be returned by Decode() if the request body is empty. We
		// check for this with errors.Is() and return a plain-english error message
		// instead.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		// A json.InvalidUnmarshalError error will be returned if we pass something
		// that is not a non-nil pointer to Decode(). We catch this and panic,
		// rather than returning an error to our handler.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// If the request body only contained a single JSON value this will
	// return an io.EOF error. So if we get anything else, we know that there is
	// additional data in the request body and we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
