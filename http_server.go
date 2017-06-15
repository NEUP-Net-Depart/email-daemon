package main

import (
	"fmt"
	"net/http"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

func HTTPServer() {
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/notify", notifyHandler)
	log.Fatal(http.ListenAndServe(":6565", nil))
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func notifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := r.ParseForm()
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userID, err := strconv.Atoi(r.Form.Get("user_id"))
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Infof("Clear userID = %d mail sent flag", userID)
	if _, ok := MailSentMap[userID]; ok {
		mu.Lock()
		MailSentMap[userID] = 0
		mu.Unlock()
	}
	w.WriteHeader(http.StatusOK)
	return
}
