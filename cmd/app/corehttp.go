package main

// import (
// 	"log"
// 	"net/http"
// )

// func main() {
// 	http.HandleFunc("/", HelloGoHandler)
// 	// Обработчик - любой тип определяющий метод ServeHTTP
// 	// Если в качестве обработчика передать nil, то будет использован DefaultServerMux
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func HelloGoHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("hello net/http"))
// }
