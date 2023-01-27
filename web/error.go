package web

import (
	"net/http"

	"golang.org/x/exp/slog"
)

// ServerErrorResponse method will be used when our application encounters an
// unexpected problem at runtime to send a 500 Internal Server Error.
func ServerErrorResponse(log *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	logError(log, r, err)
	message := "the server encountered a problem and could not process your request"
	errorResponse(log, w, r, http.StatusInternalServerError, message)
}

// BadRequestResponse method will be used to send a 400 Bad Request.
func BadRequestResponse(log *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	errorResponse(log, w, r, http.StatusBadRequest, err.Error())
}

// NotFoundResponse method will be used to send a 404 Not Found.
func NotFoundResponse(log *slog.Logger, w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	errorResponse(log, w, r, http.StatusNotFound, message)
}

// FailedValidationResponse method will be used to send a 422 Unprocessable Entity.
func FailedValidationResponse(log *slog.Logger, w http.ResponseWriter, r *http.Request, errors map[string]string) {
	errorResponse(log, w, r, http.StatusUnprocessableEntity, errors)
}

// InvalidAuthenticationResponse method will be used to send a 401 Auhorization Required.
func InvalidAuthenticationResponse(log *slog.Logger, w http.ResponseWriter, r *http.Request) {
	message := "invalid or missing authentication credentials"
	errorResponse(log, w, r, http.StatusUnauthorized, message)
}

// logError method is a generic helper for logging an error message.
func logError(log *slog.Logger, r *http.Request, err error) {
	log.Error("internal error", err, "request_method", r.Method, "request_url", r.URL.String())
}

// errorResponse method is a generic helper for sending JSON-formatted error
// messages to the client with a given status code.
func errorResponse(log *slog.Logger, w http.ResponseWriter, r *http.Request, status int, message any) {
	envelope := Envelope{"error": message}

	err := WriteJSON(w, status, envelope, nil)
	if err != nil {
		logError(log, r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
