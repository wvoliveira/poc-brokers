package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func SendResponse(w http.ResponseWriter, data interface{}, err error) {
	w.Header().Add("Content-Type", "application/json")
	response := Response{}
	response.Status = "successful"
	response.Data = data

	if err != nil {
		response.Status = "error"
		response.Message = err.Error()
		response.Data = nil

		switch err {
		case ErrNotFound:
			w.WriteHeader(404)
		default:
			w.WriteHeader(500)
		}
	}

	payload, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response object: %s\n", err.Error())
	}

	w.Write(payload)
}
