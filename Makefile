.PHONY: all clip render

all: clip render

clip:
	$(shell cd cmd/clip-shapefile; go build .)
	mv cmd/clip-shapefile/clip-shapefile .

render:
	$(shell cd cmd/render; go build .)
	mv cmd/render/render .
