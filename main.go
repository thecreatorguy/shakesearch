package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

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

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

type SearchResults struct {
	Total	int 		`json:"total"`
	Page	int 		`json:"page"`
	Length	int			`json:"length"`
	Results []string 	`json:"results"`
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
		results := searcher.Search(query, page, pageLength)
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

func (s *Searcher) Load(filename string) error {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New([]byte(strings.ToLower(string(dat))))
	return nil
}

func (s *Searcher) Search(query string, page, length int) SearchResults {
	idxs := s.SuffixArray.Lookup([]byte(strings.ToLower(query)), -1)
	results := []string{}
	// TODO: change results into a more sensical solution- this cuts off words
	for _, idx := range idxs {
		results = append(results, s.CompleteWorks[idx-250:idx+250])
	}

	// Compute the lower and upper bound
	lb := page * length
	if lb < 0 {
		lb = 0
	} else if lb > len(results) {
		lb = len(results)
	}
	ub := (page+1) * length
	if ub < 0 {
		ub = 0
	} else if ub > len(results) {
		ub = len(results)
	}

	return SearchResults{
		Total: len(results),
		Page: page,
		Length: length,
		Results: results[lb:ub],
	}
}
