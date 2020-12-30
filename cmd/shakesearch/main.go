package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/thecreatorguy/shakesearch/pkg/shakesearch"
)

func main() {
	r := mux.NewRouter()

	searcher := shakesearch.Searcher{}
	fmt.Println("Loading in the works of Shakespeare!..")
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Loaded!")

	shakesearch.AddRoutes(r, searcher, "./views/index.go.html", "/", "/assets/", "./static")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}
	fmt.Printf("Listening on port %s...", port)
	server := &http.Server{
		Handler:        r,
		Addr: 			fmt.Sprintf(":%s", port),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(server.ListenAndServe())
}