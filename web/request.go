package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/mroobert/tixer-pkgs/validate"
)

// ReadIDParam standardizes the way we read & validate url ID parameters across the Tixer services.
func ReadIDParam(r *http.Request) (uuid.UUID, error) {
	params := httprouter.ParamsFromContext(r.Context())
	input := params.ByName("id")

	id, err := uuid.Parse(input)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// ReadUUID standardizes the way we read & validate url UUID parameters across the Tixer services.
func ReadUUID(qs url.Values, key string, defaultValue uuid.UUID, vld *validate.Validator) uuid.UUID {
	v := qs.Get(key)

	if v == "" {
		return defaultValue
	}

	id, err := uuid.Parse(v)
	if err != nil {
		vld.AddError(key, fmt.Sprintf("'%s' is not a valid UUID: %s", v, err.Error()))
		return uuid.Nil
	}

	return id
}

// ReadInt standardizes the way we read url integer parameters across the Tixer services.
func ReadInt(qs url.Values, key string, defaultValue int, vld *validate.Validator) int {
	v := qs.Get(key)

	if v == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		vld.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

// ReadString standardizes the way we read url string parameters across the Tixer services.
func ReadString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
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
