package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/httputil"
)

const MAIN_HTML = "../static_cui/cui/templates/cui.html"

var cui_html []byte

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
	case "/chk/save/":
		val := struct {
			Task     string
			Ticket   string
			ProgLang string
			Solution string
		}{
			Task:     r.FormValue("task"),
			Ticket:   r.FormValue("ticket"),
			ProgLang: r.FormValue("prg_lang"),
			Solution: r.FormValue("solution"),
		}
		j, _ := json.Marshal(val)
		fmt.Printf("%s\n", string(j))
	}

}

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
}
