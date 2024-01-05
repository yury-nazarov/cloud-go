package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var listen = ":8080"
var logger TransactionLogger


func main() {
	if err := initializeTransactionLog(); err != nil {
		log.Fatal("can not init Transaction file: ", err.Error())
	}

	r := mux.NewRouter()
	// curl -X PUT -d 'Hello, key-value store!' -v http://localhost:8080/v1/key-a
	r.HandleFunc("/v1/{key}", keyValuePutHandler).Methods("PUT")
	// curl -v http://localhost:8080/v1/key-a
	r.HandleFunc("/v1/{key}", keyValueGetHandler).Methods("GET")
	// curl -X DELETE -v http://localhost:8080/v1/key-a
	r.HandleFunc("/v1/{key}", keyValueDeleteHandler).Methods("DELETE")


	fmt.Printf("%+v\n\n", store.m)

	log.Println("app was started on: ", listen)
	log.Fatal(http.ListenAndServe(listen, r))
}



func initializeTransactionLog() error {
	var err error

	logger, err = NewFileTransactionLogger("transaction.log")
	if err != nil {
		return fmt.Errorf("filed to create event logger: %w", err)
	}
	// Читаем журнал транзакций, загружая в RAM данные
	//events, errors := logger.ReadEvents()
	events, errors := logger.ReadEvents()

	e, ok := Event{}, true

	// Пока не получим ошибку - на пример при Put или Delete
	// или не закроется канал
	for ok && err == nil {
		select {
		// ok=true пока канал не закрыт
		case err, ok = <-errors:

		case e, ok = <-events:
			switch e.EventType {
			case EventDelete:
				err = Delete(e.Key)
				fmt.Println("Transaction upload. Delete key:", e.Key)
			case EventPut:
				err = Put(e.Key, e.Value)
				fmt.Printf("Transaction upload. Put key: %s, value: %s\n", e.Key, e.Value)
			}
		}
	}

	logger.Run()
	return err
}
