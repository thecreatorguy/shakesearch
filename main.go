package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/blevesearch/bleve"
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

const (
	TITLE_MARKER = "🙂"
	MAX_PAGE_LEN = 30
)

type Searcher struct {
	Index bleve.Index
}

type ShakeDocument struct {
	Work string `json:"work"`
	Text string `json:"text"`
	Lines int `json:"lines"`
}

type SearchResult struct {
	ID string `json:"id"`
	Fragments []string `json:"fragments"`
	ShakeDocument
}

type SearchResults struct {
	Total	int 			`json:"total"`
	Page	int 			`json:"page"`
	Length	int				`json:"length"`
	Results []SearchResult	`json:"results"`
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
			fmt.Println(err)
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

func (s *Searcher) Load(filename string) error {
	docMapping := bleve.NewDocumentStaticMapping()
	docMapping.AddFieldMappingsAt("work", bleve.NewTextFieldMapping())
	docMapping.AddFieldMappingsAt("text", bleve.NewTextFieldMapping())
	linesMapping := bleve.NewNumericFieldMapping()
	linesMapping.Index = false
	docMapping.AddFieldMappingsAt("lines", linesMapping)

	idxMapping := bleve.NewIndexMapping()
	idxMapping.DefaultMapping = docMapping

	// To preserve the original benefit, this index should not be persisted and remain only in memory
	index, err := bleve.NewMemOnly(idxMapping)
	if err != nil {
		return err
	}
	s.Index = index
	
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	works := strings.Split(string(dat), TITLE_MARKER)[1:]

	// Free the memory
	dat = nil

	var wg sync.WaitGroup
	wg.Add(len(works))
	headerBlockRe := regexp.MustCompile(`\n[A-Z.,1-9 ]+\n`)
	nonEmptyNewlineRe := regexp.MustCompile(`\n\s*`)
	batches := make([]*bleve.Batch, len(works))
	for i, w := range works {
		go func(i int, w string) {
			defer wg.Done()
			batches[i] = index.NewBatch()
			_, title, _ := bufio.ScanLines([]byte(w), false)
			// if err != nil {
			// 	return err
			// }
			
			indices := headerBlockRe.FindAllStringIndex(w, -1)
			startIdx := 0
			j := 0
			for _, idx := range indices {
				// Strip off the starting newline
				block := w[startIdx:idx[0]-1]
				startIdx = idx[0] + 1
				batches[i].Index(fmt.Sprintf("%s-%v", title, j), ShakeDocument{
					Work: string(title),
					Text: block,
					Lines: len(nonEmptyNewlineRe.FindAllStringIndex(block, -1)) + 1,
				})
				j++
			}
			block := w[startIdx:]
			batches[i].Index(fmt.Sprintf("%s-%v", title, j), ShakeDocument{
				Work: string(title),
				Text: block,
				Lines: len(nonEmptyNewlineRe.FindAllStringIndex(block, -1)) + 1,
			})
		}(i, w)
	}
	wg.Wait()

	combined := index.NewBatch()
	for _, b := range batches {
		combined.Merge(b)
	}
	return index.Batch(combined)
}

func (s *Searcher) Search(query string, page, length int) (SearchResults, error) {
	// Construct the query- the match query has fuzziness embedded
	q := bleve.NewMatchQuery(query)
	q.SetField("text")
	search := bleve.NewSearchRequestOptions(q, length, page*length, false)
	search.Fields = []string{"*"}
	search.Highlight = bleve.NewHighlightWithStyle("html") // html.Name
	r, err := s.Index.Search(search)
	if err != nil {
		return SearchResults{}, err
	}

	if r.Total == 0 {
		q.SetFuzziness(1)
		r, err = s.Index.Search(search)
		if err != nil {
			return SearchResults{}, err
		}

		if r.Total == 0 {
			q.SetFuzziness(2)
			r, err = s.Index.Search(search)
			if err != nil {
				return SearchResults{}, err
			}
		}
	}

	// Convert the results to the ShakeSearch result form
	results := make([]SearchResult, len(r.Hits))
	for i, h := range r.Hits {
		temp, err := json.Marshal(h.Fields)
		if err != nil {
			return SearchResults{}, err
		}
		err = json.Unmarshal(temp, &results[i])
		if err != nil {
			return SearchResults{}, err
		}
		results[i].ID = h.ID
		results[i].Fragments = h.Fragments["text"]
	}

	return SearchResults{
		Total: int(r.Total),
		Page: page,
		Length: length,
		Results: results,
	}, nil
}
