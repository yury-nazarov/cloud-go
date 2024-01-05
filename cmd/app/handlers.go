package main

import (
	"io"
	"net/http"

	"github.com/go-errors/errors"
	"github.com/gorilla/mux"
)

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
