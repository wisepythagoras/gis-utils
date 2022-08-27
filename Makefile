.PHONY: all clip render tiles

all: clip render tiles

clip:
	$(shell cd cmd/clip-shapefile; go build .)
	mv cmd/clip-shapefile/clip-shapefile .

render:
	$(shell cd cmd/render; go build .)
	mv cmd/render/render .

tiles:
	$(shell cd cmd/tiles; go build .)
	mv cmd/tiles/tiles .
