package valmid_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/iamolegga/valmid"
)

type Body struct {
	Name string `json:"name" validate:"required,min=3"`
}

type Input struct {
	ID     int    `in:"path=id" validate:"gt=0"`
	Page   int    `in:"query=page;default=1"`
	Token  string `in:"header=X-Token" validate:"required"`
	Body   *Body  `in:"body=json"`
	Field  string `in:"form=field"`
}

type BadInput struct {
	Field string `in:"unknown=bad"`
}

func TestMiddleware(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("POST /users/{id}", valmid.Middleware[Input]()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			input := valmid.Get[Input](r)
			if input.ID != 42 || input.Page != 2 || input.Token != "secret" || input.Body.Name != "John" {
				t.Errorf("unexpected input: %+v", input)
			}
		}),
	))

	req := httptest.NewRequest("POST", "/users/42?page=2", strings.NewReader(`{"name":"John"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", "secret")
	mux.ServeHTTP(httptest.NewRecorder(), req)
}

func TestMiddleware_BindingError(t *testing.T) {
	handler := valmid.Middleware[Input]()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMiddleware_ValidationError(t *testing.T) {
	handler := valmid.Middleware[Input]()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"Jo"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token", "secret")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMiddleware_WithErrorHandler(t *testing.T) {
	called := false
	handler := valmid.Middleware[Input](
		valmid.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			called = true
			w.WriteHeader(http.StatusTeapot)
		}),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called || rec.Code != http.StatusTeapot {
		t.Errorf("custom error handler not called correctly")
	}
}

func TestMiddleware_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid input type")
		}
	}()
	valmid.Middleware[BadInput]()
}

func TestSetErrorHandler(t *testing.T) {
	called := false
	valmid.SetErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
		called = true
		w.WriteHeader(http.StatusTeapot)
	})
	defer valmid.SetErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, err.Error(), http.StatusBadRequest)
	})

	handler := valmid.Middleware[Input]()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called || rec.Code != http.StatusTeapot {
		t.Errorf("global error handler not called correctly")
	}
}

func TestSetValidator(t *testing.T) {
	v := validator.New()
	valmid.SetValidator(v)

	// just verify it doesn't panic and uses the custom validator
	_ = valmid.Middleware[Input]()
}

type MultiQueryInput struct {
	Tags []string `in:"query=tag" validate:"required,min=1,dive,required"`
}

func TestMiddleware_MultipleQueryParams(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("GET /items", valmid.Middleware[MultiQueryInput]()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			input := valmid.Get[MultiQueryInput](r)
			if len(input.Tags) != 3 || input.Tags[0] != "a" || input.Tags[1] != "b" || input.Tags[2] != "c" {
				t.Errorf("unexpected tags: %+v", input.Tags)
			}
		}),
	))

	req := httptest.NewRequest("GET", "/items?tag=a&tag=b&tag=c", nil)
	mux.ServeHTTP(httptest.NewRecorder(), req)
}

func TestGet_NoMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	input := valmid.Get[Input](req)
	if input.ID != 0 || input.Token != "" {
		t.Error("expected zero value")
	}
}
