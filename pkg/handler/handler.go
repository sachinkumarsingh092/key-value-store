package handler

import (
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
	r.HandleFunc("/db/set", h.setHandler)
	r.HandleFunc("/db/get", h.getHandler)

	return r, nil
}

func (h *handler) setHandler(w http.ResponseWriter, r *http.Request) {
	key, err := getKey(r.URL.Path)
	if err != nil {
		fmt.Printf("%v requested key not valid", err)
	}

	// if r.Header.Get("Content-Type") != "application/octet-stream" {
	// 	return fmt.Errorf("Mime-Type not supported, application/octet-stream is supported")
	// }

	value, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	log.Printf("Stored key:%v and value:%v in database\n", key, string(value))
	h.db.Set(key, string(value))
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *handler) getHandler(w http.ResponseWriter, r *http.Request) {
	// key, err := getKey(r.URL.Path)
	// if err != nil {
	// 	fmt.Print("requested key not valid")
	// }
	value := h.db.Get("set")

	if value == nil {
		fmt.Print("key does not exist")
	}

	if r.URL.Query().Get("format") == "json" {
		w.Header().Set("Content-Type", "application/json")
	}

	fmt.Fprintf(w, "%s\n", value.(string))
}

func getKey(s string) (string, error) {
	key := strings.TrimPrefix(s, "/db/")
	arr := strings.Split(key, "/")
	if len(arr) != 1 {
		return "", fmt.Errorf("key contains a slash /, %s", key)
	}
	return key, nil
}

// func (h *handler) handleDB(w http.ResponseWriter, r *http.Request) func(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case http.MethodGet:
// 		return h.getHandler(w, r)
// 	case http.MethodPost:
// 		return h.setHandler(w, r)
// 	default:
// 		return errorf(fmt.Errorf(""), http.StatusMethodNotAllowed, r.Method+": method not allowed")
// 	}
// }
