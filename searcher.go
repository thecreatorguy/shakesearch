package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/blevesearch/bleve"
)

const (
	// TitleMarker is the marker that directly precedes the title of a work
	TitleMarker = "🙂"

	// MaxPageLength is the maximum number of lines to return as a page
	MaxPageLength = 25
)

var headerBlockRe = regexp.MustCompile(`\n[A-Z.,1-9 ]+\n`)

// Searcher stores the intelligently separated text of Shakespeare for searching
type Searcher struct {
	Index bleve.Index
	WorkLengths map[string]int
}

// ShakeDocument is a single document from the works of Shakespeare, which should
// be a single "block" of text preceded by a header
type ShakeDocument struct {
	Work string `json:"work"`
	Text string `json:"text"`
}

// SearchResult is a single result from the search function
type SearchResult struct {
	ID string `json:"id"`
	Fragments []string `json:"fragments"`
	ShakeDocument
}

// SearchResults is the collection of results from the search function
type SearchResults struct {
	Total	int 			`json:"total"`
	Page	int 			`json:"page"`
	Length	int				`json:"length"`
	Results []SearchResult	`json:"results"`
}

// Load takes the given file, processes it, and allows it to be searched with
// the Search() function. It is tailored to the Shakespeare works, where each
// is preceded by the constant utf-8 character "🙂". The documents that will
// be produced are each full "blocks" of text, which at the moment are sections
// of text that are preceded by a header of a character name, act, etc.
func (s *Searcher) Load(filename string) error {
	// Create the index meant to store the ShakeDocument structs. Static mapping ensures
	// that only the data we want is present and stored. To preserve the original format,
	// this index is created in-memory
	docMapping := bleve.NewDocumentStaticMapping()
	docMapping.AddFieldMappingsAt("work", bleve.NewTextFieldMapping())
	docMapping.AddFieldMappingsAt("text", bleve.NewTextFieldMapping())

	idxMapping := bleve.NewIndexMapping()
	idxMapping.DefaultMapping = docMapping

	index, err := bleve.NewMemOnly(idxMapping)
	if err != nil {
		return err
	}
	s.Index = index
	
	// Read the file into memory and split it into individual works based on the 🙂 character
	// This has room for improvement in memory efficiency, but it works for the size of
	// Shakespeare so it's fine to do it this way.
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	works := strings.Split(string(dat), TitleMarker)[1:]
	dat = nil  // Free the memory

	// On each work, concurrently generate documents that are separated by "block", or a
	// group of text that has a header line.
	var wg sync.WaitGroup
	wg.Add(len(works))
	batches := make([]*bleve.Batch, len(works))
	s.WorkLengths = make(map[string]int)
	var mu sync.Mutex
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
				})
				j++
			}
			block := w[startIdx:]
			batches[i].Index(fmt.Sprintf("%s-%v", title, j), ShakeDocument{
				Work: string(title),
				Text: block,
			})
			mu.Lock()
			s.WorkLengths[string(title)] = j + 1
			mu.Unlock()
		}(i, w)
	}
	wg.Wait()

	// I don't believe that the index is thread safe, so batch everything at once.
	combined := index.NewBatch()
	for _, b := range batches {
		combined.Merge(b)
	}
	return index.Batch(combined)
}

// Search searches through the internal contents for the given query, returning
// "length" results, starting from the "page"th page of length results.
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

	// Increase fuzziness as results are not found
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

// Preview gets the preview of a document by its ID, including surrounding text, padding
// the preview to at least 30 lines
func (s *Searcher) Preview(id string) (string, error) {
	split := strings.Split(id, "-")
	work := strings.Join(split[:len(split)-1], "-")
	i, err := strconv.Atoi(split[len(split)-1])
	if err != nil {
		return "", err
	}

	preview, err := s.GetDocumentText(id)
	if err != nil {
		return "", err
	}
	
	lb := i-1
	ub := i+1
	for strings.Count(preview, "\n") < MaxPageLength {
		if lb >= 0 {
			prevBlock, err := s.GetDocumentText(fmt.Sprintf("%v-%v", work, lb))
			if err != nil {
				return "", err
			}
			preview = prevBlock + "\n" + preview
			lb--
		}

		if ub < s.WorkLengths[work] && strings.Count(preview, "\n") < MaxPageLength {
			nextBlock, err := s.GetDocumentText(fmt.Sprintf("%v-%v", work, ub))
			if err != nil {
				return "", err
			}
			preview = nextBlock + "\n" + preview
			ub++
		}
	}

	return preview, nil
}

// GetDocumentText gets the text of a document by its ID
func (s *Searcher) GetDocumentText(id string) (string, error) {
	d, err := s.Index.Document(id)
	if err != nil {
		return "", err
	}
	for _, f := range d.Fields {
		if f.Name() == "text" {
			return string(f.Value()), nil
		}
	}

	return "", nil
}