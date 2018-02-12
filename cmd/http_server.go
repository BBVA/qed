package main

import (
	"log"
	"net/http"
	"verifiabledata/api"
)

func main() {

	http.HandleFunc("/health-check", api.HealthCheckHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
