package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func AuthMiddleware(h http.Handler) http.Handler {
	log.Println("Creating AuthMiddleWare")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("AuthMiddleWare Invoked")
		h.ServeHTTP(w, r)
		// w.WriteHeader(400)
	})
}

func ResourceHandler(handlr func(w http.ResponseWriter, r *http.Request), verb string) http.Handler {
	
	log.Println("Creating ResourceHandler")
	wrappedHandlr := AuthMiddleware(http.HandlerFunc(handlr))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("ResourceHandler Invoked")
		if r.Method == verb {
			wrappedHandlr.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

	})

}

func Sample(w http.ResponseWriter, r *http.Request) {
	var f interface{}
	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &f)
	m := f.(map[string]interface{})
	sm, _ := json.Marshal(m)
	log.Println("In Sample", sm )
	// w.WriteHeader(203)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)

}

func main() {

	http.Handle("/sample", ResourceHandler(Sample, http.MethodGet))
	http.ListenAndServe(":8092", nil)
}
