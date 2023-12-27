package main

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	// curl -X PUT -d 'Hello, key-value store!' -v http://localhost:8080/v1/key-a
	r.HandleFunc("/v1/{key}", keyValuePutHandler).Methods("PUT")
	// curl -v http://localhost:8080/v1/key-a
	r.HandleFunc("/v1/{key}", keyValueGetHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func keyValuePutHandler(w http.ResponseWriter, r *http.Request) {
	// Получить ключ из запроса
	vars := mux.Vars(r)
	key := vars["key"]

	// тело запроса хранит значение
	value, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	// если возникла ошибка сообщаем о ней
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
	}
	// Сохраняем значение в хранилище как строку
	err = Put(key, string(value))
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
	}
	// Если все ок, отправляем ответ
	w.WriteHeader(http.StatusCreated)
}

func keyValueGetHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ключи из запроса
	vars := mux.Vars(r)
	key := vars["key"]

	// Получаем из хранилища данные для ключа
	value, err := Get(key)
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Сообщаем значение
	w.Write([]byte(value))
}
