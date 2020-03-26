package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func rHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	response["hello"] = "world"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
	log.Println("executing response..!")

}

func main() {
	// 1: Create a mux, don't use the default Mux as it's global
	mux := http.NewServeMux()
	// 2: Convert your function to a HandlerFunc Type and add to mux using this shortcut
	hdl := http.HandlerFunc(rHandler)
	//3 Assign the Handle to the mux
	mux.Handle("/", hdl)
	// 4:Assign the mux a server
	http.ListenAndServe(":3000", mux)
}
