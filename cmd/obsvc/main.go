package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/handlers"

	"github.com/falun/obsvc/api"
	"github.com/falun/obsvc/collector"
)

func main() {
	configFile := flag.String("config", "", "specify yaml config")
	flag.Parse()
	rand.Seed(time.Now().Unix())

	if configFile == nil || *configFile == "" {
		log.Fatalf("config is required")
	}

	store := collector.NewStore()
	apiHandler := api.New(store, []api.CollectorHandler{})

	http.ListenAndServe(":8080", handlers.CORS()(apiHandler))
}
