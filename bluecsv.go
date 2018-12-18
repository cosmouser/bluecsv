package main

import (
	"./actions"
	"./config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	http.HandleFunc("/", actions.HomeHandler)
	http.HandleFunc("/faq", actions.FaqHandler)
	http.HandleFunc("/login", actions.Authenticate)
	http.HandleFunc("/logout", actions.Logout)
	http.HandleFunc("/form", actions.UploadHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(config.Port, nil)
}
