package main

import (
	"log"
	"net/http"
)

func main() {
	// Simple static webserver:
	http.Handle("/cui/", http.StripPrefix("/cui/", http.FileServer(http.Dir("../static_cui/cui/"))))
	log.Fatal(http.ListenAndServe(":8082", nil))
}