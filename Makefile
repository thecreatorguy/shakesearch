.PHONY: run

shakesearch: searcher.go main.go
	go build -o shakesearch *.go

run: shakesearch
	./shakesearch