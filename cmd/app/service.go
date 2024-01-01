package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var logger TransactionLogger

func initializeTransactionLog() error {
	var err error

	logger, err = NewFileTransactionLogger("transaction.log")
	if err != nil {
		return fmt.Errorf("filed to create event logger: %w", err)
	}

	events, errors := logger.ReadEvents()
	e, ok := Event{}, true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case EventDelete:
				err = Delete(e.Key)
			case EventPut:
				err = Put(e.Key, e.Value)
			}
		}
	}
	logger.Run()
	return err
}

func main() {
	r := mux.NewRouter()
	// curl -X PUT -d 'Hello, key-value store!' -v http://localhost:8080/v1/key-a
	r.HandleFunc("/v1/{key}", keyValuePutHandler).Methods("PUT")
	// curl -v http://localhost:8080/v1/key-a
	r.HandleFunc("/v1/{key}", keyValueGetHandler).Methods("GET")
	// curl -X DELETE -v http://localhost:8080/v1/key-a
	r.HandleFunc("/v1/{key}", keyValueDeleteHandler).Methods("DELETE")

	if err := initializeTransactionLog(); err != nil {
		log.Fatal("can not init Transaction file", err.Error())
	}
	listen := ":8080"
	log.Println("app was started on: ", listen)
	log.Fatal(http.ListenAndServe(listen, r))
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
		return
	}
	// Логируем
	logger.WritePut(key, string(value))
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
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func keyValueDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := Delete(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Логируем
	logger.WriteDelete(key)
	w.WriteHeader(http.StatusOK)
}
