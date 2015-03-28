package main

import (
	"log"
	"net/http"
	"io/ioutil"
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

const MAIN_HTML = "../static_cui/cui/cui.html"

func htmlResponseHandler(html []byte) http.Handler {
	fn := func (w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		w.Write(html)
	}
	return http.HandlerFunc(fn)
}

func main() {
	var html []byte
	var err error
	if html, err = ioutil.ReadFile(MAIN_HTML); err != nil {
        log.Fatal(err)
    }

    mh := htmlResponseHandler(html)
    fs := http.FileServer(http.Dir("../static_cui/cui/"))

	mux := http.NewServeMux()
	mux.Handle("/", mh)
	mux.Handle("/static/", fs)
	
	log.Fatal(http.ListenAndServe(":8082", mux))
	//log.Fatal(http.ListenAndServe(":8088", http.HandlerFunc(handleHttp)))
}