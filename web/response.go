package web

import (
	"encoding/json"
	"net/http"
)

// Used as an envelope type in the http response.
// This isn't strictly necessary, but we can say it's a good idea:
//
// 1. For any humans who see the response out of context, it is a bit easier to understand what the data relates to.
// 2. It reduces the risk of errors on the client side, because it’s harder to accidentally process one response thinking
// that it is something different.
type Envelope map[string]any

// WriteJSON is a helper function to standardize the way that we write JSON responses across the Tixer services.
func WriteJSON(w http.ResponseWriter, status int, data Envelope, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// At this point, we know that we won't encounter any more errors before writing the
	// response, so it's safe to add any headers that we want to include.
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}
