package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/wisepythagoras/gis-utils/gis"
)

func parseBBox(bboxStr string) (*gis.BBox, error) {
	parts := strings.Split(bboxStr, ",")

	if len(parts) < 4 {
		return nil, errors.New("invalid bounding box")
	}

	bbox := &gis.BBox{}

	for i, part := range parts {
		part = strings.TrimSpace(part)
		coord, err := strconv.ParseFloat(part, 64)

		if err != nil {
			return nil, err
		}

		if i == 0 {
			bbox.NE.Lon = coord
		} else if i == 1 {
			bbox.NE.Lat = coord
		} else if i == 2 {
			bbox.SW.Lon = coord
		} else if i == 3 {
			bbox.SW.Lat = coord
		}
	}

	if bbox.NE.Lon < bbox.SW.Lon || bbox.NE.Lat < bbox.SW.Lat {
		return nil, errors.New("the ordering of the bounding box coordinates is invalid")
	}

	return bbox, nil
}

func main() {
	shapefilePtr := flag.String("shapefile", "", "The path to the shapefile that you need to clip")
	outputPtr := flag.String("output", "", "The output path")
	bboxPtr := flag.String("bbox", "", "The bounding box of the area to clip (NE Lon,NE Lat,SW Lon,SW Lat")
	flag.Parse()

	if len(*shapefilePtr) == 0 {
		fmt.Println("A path to a shapefile is required (use -shapefile path/to/shapefile).")
		os.Exit(1)
	}

	bbox, err := parseBBox(*bboxPtr)

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	geojson, _ := bbox.ToGeoJSONStr()
	fmt.Println(string(geojson))
	outputPath := *outputPtr

	// Construct a default filename for the output, in case one was not passed.
	if len(outputPath) == 0 {
		_, filename := path.Split(*shapefilePtr)
		filename = strings.TrimSuffix(filename, filepath.Ext(filename))
		outputPath = fmt.Sprintf("%s_%s.shp", filename, *bboxPtr)
	}

	shapefile := &gis.Shapefile{Filename: *shapefilePtr}
	shapefile.Load()
	features, err := shapefile.Clip(bbox)

	fmt.Println(len(features), "features found within the bounding box.")

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	shapefile.SaveClippedShapefile(outputPath)

	fmt.Printf("The clipped shapefile was saved as %s\n", outputPath)
}
