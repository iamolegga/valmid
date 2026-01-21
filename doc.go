// Package valmid provides HTTP request validation middleware for Go 1.22+.
//
// It combines [httpin] for request binding and [go-playground/validator] for
// struct validation, working with standard net/http middleware pattern.
//
// # Basic Usage
//
//	type UserBody struct {
//	    Name  string `json:"name" validate:"required,min=3"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	type CreateUserInput struct {
//	    ID   int       `in:"path=id" validate:"gt=0"`
//	    Body *UserBody `in:"body=json" validate:"required"`
//	}
//
//	mux := http.NewServeMux()
//	mux.Handle("POST /users/{id}",
//	    valmid.Middleware[CreateUserInput]()(
//	        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	            input := valmid.Get[CreateUserInput](r)
//	            // input is validated and ready to use
//	        }),
//	    ),
//	)
//
// # Supported Input Sources
//
// Use httpin tags to bind data from different sources:
//
//	in:"path=id"              // URL path parameter (Go 1.22+ r.PathValue)
//	in:"query=page"           // Query string parameter
//	in:"header=Authorization" // HTTP header
//	in:"form=field"           // Form field (application/x-www-form-urlencoded)
//	in:"body=json"            // JSON body (binds to nested struct)
//
// Multiple sources and defaults:
//
//	in:"query=token;header=X-Token"  // Try query first, then header
//	in:"query=page;default=1"        // Default value if missing
//	in:"query=id;required"           // httpin-level required check
//
// # JSON Body Binding
//
// JSON bodies bind to a nested struct field, not individual fields:
//
//	type Body struct {
//	    Name string `json:"name" validate:"required"`
//	}
//
//	type Input struct {
//	    Body *Body `in:"body=json"`
//	}
//
// # Validation
//
// Use go-playground/validator tags for validation:
//
//	validate:"required"           // Field must be present
//	validate:"email"              // Must be valid email
//	validate:"min=3,max=100"      // String length or numeric bounds
//	validate:"gt=0"               // Greater than zero
//	validate:"oneof=admin user"   // Must be one of values
//
// Nested structs are validated automatically.
//
// # Error Handling
//
// Default error handler returns HTTP 400 with error message.
// Customize per-middleware or globally:
//
//	// Per-middleware
//	valmid.Middleware[Input](
//	    valmid.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
//	        w.WriteHeader(http.StatusUnprocessableEntity)
//	        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
//	    }),
//	)
//
//	// Global default
//	valmid.SetErrorHandler(myErrorHandler)
//
// # Custom Validator
//
// Register custom validation rules:
//
//	v := validator.New()
//	v.RegisterValidation("customrule", customFunc)
//	valmid.SetValidator(v)
//
// [httpin]: https://github.com/ggicci/httpin
// [go-playground/validator]: https://github.com/go-playground/validator
package valmid
