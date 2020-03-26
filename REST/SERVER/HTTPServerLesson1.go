package main

import (
	"encoding/json"
	"net/http"
)




func rHandler(w http.ResponseWriter, r *http.Request)  {	
	response := make(map[string]interface{}) 
	response["hello"] = "world"
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(response)
	
	
}


func main() {
	// Step 1: Create a mux, don't use the default Mux as it's global
	mux := http.NewServeMux()
	// Convert your function to a HandlerFunc Type and add to mux using this shortcut
	mux.HandleFunc("/",rHandler)
	// Assign the mux a server
	http.ListenAndServe(":3000", mux)
}