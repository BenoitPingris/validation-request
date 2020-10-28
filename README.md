https://github.com/BenoitPingris/validation-request/workflows/go/badge.svg

# validation-request

Middleware to validate incomming body using the [validator.v10](https://godoc.org/gopkg.in/go-playground/validator.v10) package

## Installation

- `go get -u github.com/BenoitPingris/validation-request`


## How to use

```go
type LoginRequest struct {
  Email string    `json:"email" validate:"email"`
  Password string `json:"password" validate:"min=6"`
}

func main() {
  r := chi.NewRouter()

  r.With(validation.Validate(LoginRequest{})).Post("/login", func(w http.ResponseWriter, r *http.Request) {
    payload := validate.FromContext(r.Context()).(*LoginRequest)
    w.Write([]byte(payload.Email))
  })
  
  http.ListenAndServe(":3001", r)
}
```

- You need to call the `Validate` methods as a middleware and passing a struct with `validate` tags to it.
- In the handler you can easily get the body from your request by calling the `FromContext` method by passing the request context as a parameter, and then cast it to the wished type
  - `body := validate.FromContext(r.Context()).(*BodyRequest)`
