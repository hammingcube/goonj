package main

import (
	"log"
	"net/http"
)

func handleHttp(w http.ResponseWriter, r *http.Request) {
	cmd := r.URL.Path[len("/c/"):]
	if cmd == "_get_task" {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write(getTask())
	} else {
		w.Write([]byte("Hello World\n"))
	}
}

func main() {
	// Simple static webserver:
	http.Handle("/c/", http.HandlerFunc(handleHttp))
	http.Handle("/cui/", http.StripPrefix("/cui/", http.FileServer(http.Dir("../static_cui/cui/"))))
	log.Fatal(http.ListenAndServe(":8082", nil))
	//log.Fatal(http.ListenAndServe(":8088", http.HandlerFunc(handleHttp)))
}