package db

import (
	"sync"
)

type DB struct {
	pid   int
	mu    sync.RWMutex
	store map[string]interface{}
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
