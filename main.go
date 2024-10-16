package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"service": "active"}
	json.NewEncoder(w).Encode(response)
}

func main() {
	r := mux.NewRouter()
	fmt.Println("listening to port 8000")
	r.HandleFunc("/", rootHandler).Methods("GET")

	http.ListenAndServe(":8000", r)
}
