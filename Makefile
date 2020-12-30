.PHONY: run shakesearch

shakesearch: 
	go build -o build ./...

run: shakesearch
	build/shakesearch