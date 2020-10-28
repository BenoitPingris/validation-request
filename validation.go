package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type contextKey struct {
	name string
}

var payloadCtxKey = &contextKey{"Payload"}

var v *validator.Validate

func init() {
	v = validator.New()
}

func buildError(err error) []byte {
	errors := map[string]string{}
	for _, err := range err.(validator.ValidationErrors) {
		if err.Tag() == "email" {
			errors[strings.ToLower(err.Field())] = "Invalid email format."
			continue
		}
		errors[strings.ToLower(err.Field())] = fmt.Sprintf("%s is %s %s", err.Field(), err.Tag(), err.Param())
	}
	body, _ := json.Marshal(errors)
	return body
}

// Validate is a middleware that checks if the incomming body matches the validation rules passed as `rules`
func Validate(rules interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b := reflect.New(reflect.TypeOf(rules)).Interface()
			if r.Body == nil {
				http.Error(w, "Invalid body.", http.StatusUnprocessableEntity)
				return
			}
			if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := v.Struct(b); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write(buildError(err))
				return
			}
			next.ServeHTTP(w, r.WithContext(NewContext(r.Context(), b)))
		})
	}
}

// NewContext returns a new context with the payload in it
func NewContext(c context.Context, payload interface{}) context.Context {
	return context.WithValue(c, payloadCtxKey, payload)
}

// FromContext returns the payload contained in the context of the request
func FromContext(c context.Context) interface{} {
	return c.Value(payloadCtxKey)
}
