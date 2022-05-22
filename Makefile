all: clip

clip:
	$(shell cd cmd/clip-shapefile; go build .)
	mv cmd/clip-shapefile/clip-shapefile .
