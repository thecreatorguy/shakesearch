package shakesearch

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// AddRoutes adds the shakesearch web routes to the given router, searching through the works of
// Shakespeare with the given searcher, and with the static files located at the given path.
// NOTE: I don't currently know what the best way to add static files to a project like this is,
// so my plan is to import the project as a git module and then link the files in directly by
// that path, and the searcher can be loaded through the same technique.
func AddRoutes(r *mux.Router, searcher Searcher, indexTemplatePath, rootPath, assetsPrefix, assetsDir string) {
	fs := http.StripPrefix(assetsPrefix, http.FileServer(http.Dir(assetsDir)))
	r.PathPrefix(assetsPrefix).Handler(fs)

	r.HandleFunc(rootPath, handleRoot(indexTemplatePath, assetsPrefix)).Methods("GET")
	r.HandleFunc("/search", handleSearch(searcher)).Methods("GET")
	r.HandleFunc("/preview", handlePreview(searcher)).Methods("GET")
}

// handleRoot returns a handler that returns the index page with the correct assets path filled in
func handleRoot(indexTemplatePath, assetsPrefix string) func(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	pageTemplate := template.Must(template.ParseFiles(indexTemplatePath))
	err := pageTemplate.ExecuteTemplate(&buf, "index", struct{AssetsPrefix string}{assetsPrefix})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	index := buf.Bytes()
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(index)
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
