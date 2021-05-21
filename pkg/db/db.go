package db

import (
	"sync"
)

type DB struct {
	mu       sync.RWMutex
	store    map[string]interface{}
	isUpdate map[string]bool
}

// NewDB returns a new initialised DB.
func NewDB() *DB {
	return &DB{
		store:    map[string]interface{}{},
		isUpdate: map[string]bool{},
	}
}

func (db *DB) Set(key string, val interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.store[key] != nil {
		db.isUpdate[key] = true
	}

	db.store[key] = val

}

func (db *DB) Get(key string) interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.store[key]
}

func (db *DB) IsUpdated(key string) bool {
	return db.isUpdate[key]
}
