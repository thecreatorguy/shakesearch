package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	searcher := Searcher{}
	fmt.Println("Loading in the works of Shakespeare!..")
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Loaded!")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))
	http.HandleFunc("/preview", handlePreview(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

// handleSearch returns a search handler with the given searcher, allowing for easy dependency injection.
func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		// Parse query
		input := r.URL.Query()
		query := input.Get("q")
		if len(query) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		page := 0
		if pagestr := input.Get("page"); pagestr != "" {
			page, err = strconv.Atoi(pagestr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("page number is not valid"))
				return
			}
		}

		pageLength := 10
		if lengthstr := input.Get("length"); lengthstr != "" {
			pageLength, err = strconv.Atoi(lengthstr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("page length is not valid"))
				return
			}
		}

		// Return results
		results, err := searcher.Search(query, page, pageLength)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("query failure"))
			return
		}

		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

// Preview is the struct that returns the preview of a given id in the preview handler
type Preview struct {
	Preview string `json:"preview"`
}

func handlePreview(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		input := r.URL.Query()
		id := input.Get("id")
		if len(id) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}

		// Return results
		preview, err := searcher.Preview(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("query failure"))
			return
		}

		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(Preview{preview})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}
