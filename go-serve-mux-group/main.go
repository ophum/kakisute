package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func main() {

	// grouping api, web
	mux := http.NewServeMux()
	mux.Handle(routerGroup("/api", func(mux *http.ServeMux) {
		mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
			SetError(r, errors.New("failed to ping1"))
			SetError(r, errors.New("failed to ping2"))
			SetError(r, errors.New("failed to ping3"))
		})
		mux.HandleFunc("POST /hello", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("world"))
		})
	}))

	log.Println("serve :8080")
	if err := http.ListenAndServe(":8080", logger(errorHandler(mux))); err != nil {
		panic(err)
	}
}

type errorsKey struct{}

func SetError(r *http.Request, err error) {
	errs, ok := r.Context().Value(errorsKey{}).(*[]error)
	if !ok {
		return
	}

	*errs = append(*errs, err)
}

func errorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errs []error
		r = r.WithContext(context.WithValue(r.Context(), errorsKey{}, &errs))
		next.ServeHTTP(w, r)
		log.Println(r.Context().Value(requestIDKey{}))
		reqID, ok := r.Context().Value(requestIDKey{}).(uuid.UUID)
		if !ok {
			reqID = uuid.UUID{}
		}

		if len(errs) == 0 {
			return
		}

		for i := len(errs) - 1; i >= 0; i-- {
			log.Printf("[%s] error %d: %+v", reqID, i, errs[i])
		}
	})
}

type requestIDKey struct{}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := uuid.Must(uuid.NewRandom())
		r = r.WithContext(context.WithValue(r.Context(), requestIDKey{}, requestID))
		next.ServeHTTP(w, r)
		end := time.Now()
		log.Println(r.Method, r.RequestURI, end.Sub(start), "reqID:", requestID)
	})
}

func routerGroup(prefix string, register func(mux *http.ServeMux)) (string, http.Handler) {
	mux := http.NewServeMux()
	register(mux)
	return prefix + "/{path...}", http.StripPrefix(prefix, mux)
}
