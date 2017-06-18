package main

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func HTTPServer() {
	http.HandleFunc("/status", statusHandler)
	log.Fatal(http.ListenAndServe(":6565", nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}
