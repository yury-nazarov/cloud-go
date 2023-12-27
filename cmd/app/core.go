package main

import (
	"errors"
)

var (
	store          = make(map[string]string)
	ErrorNoSuchKey = errors.New("no such key")
)

// Put добавить элемент
func Put(key string, value string) error {
	store[key] = value
	return nil
}

// Get получить элемент по ключу
func Get(key string) (string, error) {
	value, ok := store[key]

	if !ok {
		return "", ErrorNoSuchKey
	}
	return value, nil
}

// Delete удалить элемент по ключу
func Delete(key string) error {
	delete(store, key)
	return nil
}
