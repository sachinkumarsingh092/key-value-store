package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// DB provides all the methods needed for storage.
type DB interface {
	Set(string, interface{})
	Get(string) interface{}
}

// handler holds all http methods.
type handler struct {
	db DB
}

func New(db DB) (http.Handler, error) {
	r := http.NewServeMux()

	h := &handler{db}
	r.Handle("/db/", errorMiddleware(h.handleDB))

	return r, nil
}

// httpError contains http status code
type httpError struct {
	err  error
	code int
	msg  string
}

// errorf constructor
func errorf(err error, code int, msg string) error {
	return &httpError{err: err, code: code, msg: msg}
}

func (e *httpError) Error() string { return fmt.Sprintf("%v: %v [%d]", e.err, e.msg, e.code) }

// errorMiddleware wraps a normal handler and converts errors to corresponding http status codes
type errorMiddleware func(http.ResponseWriter, *http.Request) error

func (fn errorMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err == nil {
		return
	}

	msg := err.Error()
	jsonError := struct {
		Msg string `json:"msg"`
	}{msg}

	w.Header().Set("Content-Type", "application/json")

	if herr, ok := err.(*httpError); ok {
		w.WriteHeader(herr.code)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	body, err := json.Marshal(jsonError)
	if err != nil {
		fmt.Printf("Could not encode error data: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
	if _, err = w.Write(body); err != nil {
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

func (h *handler) handleDB(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case http.MethodGet:
		return h.getHandler(w, r)
	case http.MethodPost:
		return h.setHandler(w, r)
	default:
		return errorf(fmt.Errorf(""), http.StatusMethodNotAllowed, r.Method+": method not allowed")
	}
}

func (h *handler) setHandler(w http.ResponseWriter, r *http.Request) error {
	key, err := getKey(r.URL.Path)
	if err != nil {
		return errorf(err, http.StatusBadRequest, "requested key not valid")
	}

	if r.Header.Get("Content-Type") != "application/octet-stream" {
		return errorf(fmt.Errorf(""), http.StatusBadRequest, "Mime-Type not supported, application/octet-stream is supported")
	}

	value, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	h.db.Set(key, string(value))
	w.WriteHeader(http.StatusCreated)

	log.Printf("Stored key:%v and value:%v in database\n", key, string(value))

	return nil
}

func (h *handler) getHandler(w http.ResponseWriter, r *http.Request) error {
	key, err := getKey(r.URL.Path)
	if err != nil {
		return errorf(err, http.StatusBadRequest, "requested key not valid")
	}
	value := h.db.Get(key)
	if err != nil {
		return errorf(err, http.StatusInternalServerError, "error GET the requested key")
	}
	if value == nil {
		return errorf(fmt.Errorf(""), http.StatusNotFound, "key does not exist")
	}

	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
	}

	fmt.Fprintf(w, "%s", value.(string))

	return nil
}

// func (h *handler) watchHandler(w http.ResponseWriter, r *http.Request) {
// }

func getKey(s string) (string, error) {
	key := strings.TrimPrefix(s, "/db/")
	arr := strings.Split(key, "/")
	if len(arr) != 1 {
		return "", fmt.Errorf("key contains a slash /, %s", key)
	}
	return key, nil
}
