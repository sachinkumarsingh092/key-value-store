package db

import (
	"log"
	"sync"
	"time"
)

const (
	bufferSize = 100
	rpcTimeout = 100 * time.Millisecond
)

type response struct {
	pid int
	val interface{}
}

type request struct {
	pid      int
	iotype   int
	key      string
	val      interface{}
	respChan chan response
}

type DB struct {
	pid     int
	mu      sync.RWMutex
	store   map[string]interface{}
	reqChan chan request
}

// NewDB returns a new initialised DB.
func NewDB(pid int) *DB {
	return &DB{
		pid:   pid,
		store: map[string]interface{}{},
	}
}

func (db *DB) Set(key string, val interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.store[key] = val
}

func (db *DB) Get(key string) interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.store[key]
}

func (db *DB) Watch() {
	log.Printf("%d starts broadcast", db.pid)
}
