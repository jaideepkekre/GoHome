package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func headers(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func Sample(w http.ResponseWriter, r *http.Request) {
	var f interface{}
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &f)
	m := f.(map[string]interface{})
	something, _ := json.Marshal(m)
	print(something)
	w.WriteHeader(203)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

}

func main() {

	http.HandleFunc("/sample", Sample)
	http.HandleFunc("/headers", headers)
	http.ListenAndServe(":8092", nil)
}
