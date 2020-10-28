package validation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
)

type A struct {
	Value int `validate:"min=5"`
}

func makeBody(b interface{}) *bytes.Reader {
	requestByte, _ := json.Marshal(b)
	return bytes.NewReader(requestByte)
}

func TestValidateError(t *testing.T) {
	tests := map[string]struct {
		request func() *http.Request
		expect  int
	}{
		"No body provided": {
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/", nil)
				return req
			},
			http.StatusUnprocessableEntity,
		},
		"Body with invalid value": {
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/", makeBody(A{-2}))
				return req
			},
			http.StatusUnprocessableEntity,
		},
		"Invalid body (not a JSON)": {
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/", strings.NewReader("dummy"))
				return req
			},
			http.StatusBadRequest,
		},
		"Invalid body": {
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/", strings.NewReader(`{ "c": "test" }`))
				return req
			},
			http.StatusUnprocessableEntity,
		},
	}
	r := chi.NewRouter()

	r.Use(Validate(A{}))
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	for _, test := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, test.request())

		if w.Code != test.expect {
			t.Errorf("Invalid status code: expected %d got %d", test.expect, w.Code)
		}
	}
}

type B struct {
	Email string `json:"email" validate:"email,required"`
	Value int    `json:"value" validate:"min=5,max=10"`
}

func TestValidateSuccess(t *testing.T) {
	tests := map[string]struct {
		request func() *http.Request
		expect  string
	}{
		"Valid body": {
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/", strings.NewReader(`{ "email": "email@example.com", "value": 6 }`))
				return req
			},
			`{"email":"email@example.com","value":6}`,
		},
	}
	r := chi.NewRouter()

	r.Use(Validate(B{}))
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		body := FromContext(r.Context()).(*B)

		obj, _ := json.Marshal(body)
		w.Write(obj)
	})

	for _, test := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, test.request())
		if w.Code != http.StatusOK {
			t.Errorf("Invalid status code: expected 200 got %d", w.Code)
		}
		if w.Body.String() != test.expect {
			t.Errorf("Invalid response body: expected %s got %s", test.expect, w.Body.String())
		}
	}
}

func TestValidateErrorMessage(t *testing.T) {
	tests := map[string]struct {
		request func() *http.Request
		expect  string
	}{
		"Invalid email": {
			func() *http.Request {
				req, _ := http.NewRequest("POST", "/", strings.NewReader(`{"email": "notanemail","value":0}`))
				return req
			},
			`{"email":"Invalid email format.","value":"Value is min 5"}`,
		},
	}

	r := chi.NewRouter()

	r.Use(Validate(B{}))
	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	for _, test := range tests {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, test.request())
		if w.Body.String() != test.expect {
			t.Errorf("Invalid response body: expected %s got %s", test.expect, w.Body.String())
		}
	}
}
