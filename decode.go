package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func DecodeFind(r *http.Request) (id string) {
	vars := mux.Vars(r)
	id = vars["id"]
	return
}

func DecodeFindAll() (query string, page, size int) {
	return
}
