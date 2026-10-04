package main

import (
	"context"
	"encoding/json"
	"net/http"
)

func receiptHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid order", http.StatusBadRequest)
		return
	}
	result, err := sendReceipt(context.Background(), http.DefaultClient, order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() { http.HandleFunc("/orders/receipt", receiptHandler); http.ListenAndServe(":8080", nil) }
