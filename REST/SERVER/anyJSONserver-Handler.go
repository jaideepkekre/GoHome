package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.Handle("/", Middleware(http.HandlerFunc(sample)))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8001", r))
}

func Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		print("hi")
		h.ServeHTTP(w, r)
	})
}

func sample(w http.ResponseWriter, r *http.Request) {

	var f interface{}
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &f)
	m := f.(map[string]interface{})
	something, _ := json.Marshal(m)
	print(something)
}
