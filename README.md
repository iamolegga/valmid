# valmid

HTTP request **val**idation **mid**dleware for Go 1.22+.

Combines [httpin](https://github.com/ggicci/httpin) for request binding and [go-playground/validator](https://github.com/go-playground/validator) for struct validation.

## Install

```bash
go get github.com/iamolegga/valmid
```

## Usage

```go
package main

import (
    "encoding/json"
    "net/http"

    "github.com/iamolegga/valmid"
)

type UserBody struct {
    Name  string `json:"name" validate:"required,min=3"`
    Email string `json:"email" validate:"required,email"`
}

type CreateUserInput struct {
    ID   int       `in:"path=id" validate:"gt=0"`
    Body *UserBody `in:"body=json" validate:"required"`
}

func main() {
    mux := http.NewServeMux()
    mux.Handle("POST /users/{id}",
        valmid.Middleware[CreateUserInput]()(
            http.HandlerFunc(createUser),
        ),
    )
    http.ListenAndServe(":8080", mux)
}

func createUser(w http.ResponseWriter, r *http.Request) {
    input := valmid.Get[CreateUserInput](r)
    json.NewEncoder(w).Encode(input)
}
```

## Input Sources

```go
in:"path=id"              // URL path parameter (r.PathValue)
in:"query=page"           // Query string
in:"header=Authorization" // HTTP header
in:"form=field"           // Form field
in:"body=json"            // JSON body (binds to nested struct)
```

Combine sources and set defaults:

```go
in:"query=token;header=X-Token"  // Try query, fallback to header
in:"query=page;default=1"        // Default value
in:"query=id;required"           // Required at binding level
```

## Validation

Uses [go-playground/validator](https://github.com/go-playground/validator) tags:

```go
validate:"required"
validate:"email"
validate:"min=3,max=100"
validate:"gt=0"
validate:"oneof=admin user"
```

## Error Handling

```go
// Per-middleware
valmid.Middleware[Input](
    valmid.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
    }),
)

// Global
valmid.SetErrorHandler(handler)
```

## Custom Validator

```go
v := validator.New()
v.RegisterValidation("customrule", customFunc)
valmid.SetValidator(v)
```

## License

MIT
