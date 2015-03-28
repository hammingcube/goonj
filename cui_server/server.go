package main

import (
	"log"
	"net/http"
	"io/ioutil"
)

func handleHttp(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL)

	switch r.URL.Path {
    	case "/":
        	w.Write(cui_html)
    	case "/c/_start/":
        	w.Write([]byte("something"))
        case "/chk/clock/":
        	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
        	w.Write(getClock())
    	case "/c/_get_task/":
        	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
        	w.Write(getTask())
    }
/*
	if r.URL.Path == "/" {
		w.Write(cui_html)
	} else {
		w.Write([]byte("Hello"))
	}
	cmd := r.URL.Path[len("/c/"):]
	if cmd == "_get_task" {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.Write(getTask())
	} else {
		w.Write([]byte("Hello World\n"))
	}*/
}

const MAIN_HTML = "../static_cui/cui/templates/cui.html"

func htmlResponseHandler(html []byte) http.Handler {
	fn := func (w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		w.Write(html)
	}
	return http.HandlerFunc(fn)
}

var cui_html []byte

func main() {
	var err error
	cui_html, err = ioutil.ReadFile(MAIN_HTML)
	if err != nil {
        log.Fatal(err)
    }

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handleHttp))
	mux.Handle("/static/", http.FileServer(http.Dir("../static_cui/cui/")))
	

	log.Fatal(http.ListenAndServe(":8082", mux))
	//log.Fatal(http.ListenAndServe(":8088", http.HandlerFunc(handleHttp)))
}