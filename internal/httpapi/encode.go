package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"coldchain-alert/internal/domain"
)

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, err error) {
	code := "internal_error"
	if errors.Is(err, domain.ErrNotFound) {
		code = "not_found"
	}
	if errors.Is(err, domain.ErrUnauthorized) {
		code = "unauthorized"
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		code = "invalid_input"
	}
	writeJSON(writer, status, map[string]string{"error": code, "message": err.Error()})
}

func decodeJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func methodAllowed(writer http.ResponseWriter, request *http.Request, methods ...string) bool {
	for _, method := range methods {
		if request.Method == method {
			return true
		}
	}
	writer.Header().Set("Allow", methods[0])
	writeError(writer, http.StatusMethodNotAllowed, domain.ErrInvalidInput)
	return false
}
