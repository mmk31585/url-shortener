package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	ErrEmptyBody     = errors.New("request body is empty")
	ErrMalformedJSON = errors.New("request body contains malformed json")
	ErrUnknownField  = errors.New("request body contains an unknown field")
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	const maxBytes = 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(data); err != nil {
		var typeErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError
		switch {
		case errors.Is(err, io.EOF):
			return ErrEmptyBody
		case strings.Contains(err.Error(), "unknown field"):
			return fmt.Errorf("%w: %v", ErrUnknownField, err)
		case errors.As(err, &typeErr):
			return fmt.Errorf("%w: field %q must be %s", ErrMalformedJSON, typeErr.Field, typeErr.Type)
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("%w: %v", ErrMalformedJSON, err)
		default:
			return fmt.Errorf("%w: %v", ErrMalformedJSON, err)
		}
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: multiple json values", ErrMalformedJSON)
	}

	return nil
}

func ReadAndValidate(w http.ResponseWriter, r *http.Request, data any) error {
	if err := ReadJSON(w, r, data); err != nil {
		return err
	}
	if err := Validate.Struct(data); err != nil {
		return err
	}
	return nil
}

func FormatValidationError(err error) string {
	if err == nil {
		return ""
	}
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return err.Error()
	}
	messages := make([]string, 0, len(validationErrors))
	for _, fieldErr := range validationErrors {
		messages = append(messages, strings.ToLower(fieldErr.Field())+" is "+fieldErr.Tag())
	}
	return strings.Join(messages, ", ")
}
