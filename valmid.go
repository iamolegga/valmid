package valmid

import (
	"context"
	"net/http"
	"sync"

	"github.com/ggicci/httpin"
	httpin_integration "github.com/ggicci/httpin/integration"
	"github.com/go-playground/validator/v10"
)

func init() {
	// Register stdlib r.PathValue() for Go 1.22+ routing
	httpin_integration.UseHttpPathVariable("path")
}

// ErrorHandlerFunc handles validation/binding errors.
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

var (
	defaultErrorHandler ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	defaultValidator = validator.New()
	mu               sync.RWMutex
)

// SetErrorHandler sets the default error handler for all middlewares.
// Can be overridden per-middleware with WithErrorHandler option.
func SetErrorHandler(h ErrorHandlerFunc) {
	mu.Lock()
	defer mu.Unlock()
	defaultErrorHandler = h
}

// SetValidator sets a custom validator instance (global).
// Useful for registering custom validation rules.
func SetValidator(v *validator.Validate) {
	mu.Lock()
	defer mu.Unlock()
	defaultValidator = v
}

func getErrorHandler() ErrorHandlerFunc {
	mu.RLock()
	defer mu.RUnlock()
	return defaultErrorHandler
}

func getValidator() *validator.Validate {
	mu.RLock()
	defer mu.RUnlock()
	return defaultValidator
}

// options holds per-middleware configuration.
type options struct {
	errorHandler ErrorHandlerFunc
}

// Option configures the middleware.
type Option func(*options)

// WithErrorHandler sets a custom error handler for this middleware instance.
func WithErrorHandler(h ErrorHandlerFunc) Option {
	return func(o *options) {
		o.errorHandler = h
	}
}

// Middleware creates a validation middleware for the given input type.
// T must be a struct type with httpin and validate tags.
func Middleware[T any](opts ...Option) func(http.Handler) http.Handler {
	o := &options{
		errorHandler: nil, // will use default if nil
	}
	for _, opt := range opts {
		opt(o)
	}

	var t T
	core, err := httpin.New(t)
	if err != nil {
		panic("valmid: failed to create httpin core: " + err.Error())
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Decode request using httpin
			input, err := core.Decode(r)
			if err != nil {
				handleError(w, r, o, err)
				return
			}

			// Validate using go-playground/validator
			v := getValidator()
			if err := v.Struct(input); err != nil {
				handleError(w, r, o, err)
				return
			}

			// Store in context and proceed
			ctx := context.WithValue(r.Context(), httpin.Input, input)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func handleError(w http.ResponseWriter, r *http.Request, o *options, err error) {
	handler := o.errorHandler
	if handler == nil {
		handler = getErrorHandler()
	}
	handler(w, r, err)
}

// Get retrieves the validated input from request context.
// Returns zero value if validation middleware was not applied.
func Get[T any](r *http.Request) T {
	v, ok := r.Context().Value(httpin.Input).(*T)
	if !ok {
		var zero T
		return zero
	}
	return *v
}
