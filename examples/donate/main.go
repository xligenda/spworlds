package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/xligenda/spworlds"
)

const PORT = ":8080"

func main() {
	id := mustEnv("SPW_ID")
	token := mustEnv("SPW_TOKEN")

	spw := spworlds.NewClient(id, token)
	app := newApp(spw, mustEnv("REDIRECT_URL"), mustEnv("WEBHOOK_URL"))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /donate/{amount}", app.donateLink)
	mux.HandleFunc("POST /webhook", app.donateWebhook)
	mux.HandleFunc("GET /latest", app.latestDonate)

	srv := &http.Server{
		Addr:         PORT,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("Server starting on port " + PORT)
	log.Fatal(srv.ListenAndServe())
}
