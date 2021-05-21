package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// DB provides all the methods needed for storage.
type DB interface {
	Set(string, interface{})
	Get(string) interface{}
	IsUpdated(string) bool
}

// handler holds all http methods.
type handler struct {
	db           DB
	notification chan notification
	noListener   chan struct{}
}

type notification struct {
	key   string
	value interface{}
}

func New(db DB) (http.Handler, error) {
	r := http.NewServeMux()

	notification := make(chan notification)
	noListener := make(chan struct{}, 1)
	noListener <- struct{}{} // initially no one is hearing.

	h := &handler{
		db:           db,
		notification: notification,
		noListener:   noListener,
	}
	r.Handle("/db/", errorMiddleware(h.handleDB))
	r.HandleFunc("/watch/", h.watchHandler)

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
	return &httpError{
		err:  err,
		code: code,
		msg:  msg,
	}
}

func (e *httpError) Error() string {
	return fmt.Sprintf("%v: %v [%d]", e.err, e.msg, e.code)
}

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

	stringValue := string(value)

	h.db.Set(key, stringValue)
	w.WriteHeader(http.StatusCreated)

	log.Printf("Stored key:%v and value:%v in database\n", key, stringValue)

	select {
	case v := <-h.noListener:
		// No listeners. Send value back so that set can proceed.
		h.noListener <- v
	case h.notification <- notification{key: key, value: stringValue}:
		// Notify watch and wait for it to process.
		fmt.Printf("set: key : %v, val = %v\n", key, stringValue)
	}

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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (h *handler) watchHandler(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer conn.Close()

	log.Println("Client Connected")
	err = conn.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}

	// Empty the nolistener channel
	<-h.noListener
	defer func() {
		// When watch returns, unblock calls to another
		// listerers or clients.
		h.noListener <- struct{}{}
	}()

	for {
		notification := <-h.notification
		// listen indefinitely for new messages coming
		// through on our WebSocket connection
		// read in a message

		msg := fmt.Sprintf("watch: key = %v, val = %v", notification.key, notification.value)
		if h.db.IsUpdated(notification.key) {
			updt := fmt.Sprintf("Updating %v to value = %v", notification.key, notification.value)
			conn.WriteMessage(1, []byte(updt))
		}

		err = conn.WriteMessage(1, []byte(msg))
		if err != nil {
			log.Print(err)
		}

	}
}

func getKey(s string) (string, error) {
	key := strings.TrimPrefix(s, "/db/")
	arr := strings.Split(key, "/")
	if len(arr) != 1 {
		return "", fmt.Errorf("key contains a slash /, %s", key)
	}
	return key, nil
}
