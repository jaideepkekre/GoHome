package main

import (
	"encoding/json"
	// "fmt"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/gorilla/mux"
	
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/sample", Sample).Methods("POST")
	log.Fatal(http.ListenAndServe(":8001", router))
}



func Sample(w http.ResponseWriter, r *http.Request) {
	var f interface{}
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &f)
	m := f.(map[string]interface{})
	something, _ := json.Marshal(m)
	print(something)
}
