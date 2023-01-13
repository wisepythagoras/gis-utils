package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tdewolff/canvas"
	"github.com/wisepythagoras/gis-utils/config"
	"github.com/wisepythagoras/gis-utils/gis"
)

func main() {
	shapefilePtr := flag.String("shapefile", "", "The path to the land shapefile")
	outputPtr := flag.String("output", "out.png", "The output path")
	pbfPtr := flag.String("pbf", "", "The path to the OSM Protobuf file")
	stylesPtr := flag.String("styles", "", "The path to the style configuration file")
	widthPtr := flag.Float64("width", 320, "The width of the output image")
	verbosePtr := flag.Bool("verbose", false, "Whether to print debug information or not")
	flag.Parse()

	if len(*shapefilePtr) == 0 {
		fmt.Println("A path to a shapefile is required (use -shapefile path/to/land.shp).")
		os.Exit(1)
	} else if len(*pbfPtr) == 0 {
		fmt.Println("A path to a *.pbf is required (use -pbf path/to/file.pbf).")
		os.Exit(1)
	} else if len(*stylesPtr) == 0 {
		fmt.Println("A style configuration file is required (use -styles path/to/styles.yaml).")
		os.Exit(1)
	}

	conf := &config.Config{UseMap: true}
	err := conf.ParseFile(*stylesPtr)

	if err != nil {
		panic(err)
	}

	f, err := os.Open(*pbfPtr)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	pbf := &gis.PBF{Verbose: *verbosePtr}
	pbf.Init()

	if err := pbf.Load(f); err != nil {
		panic(err)
	}

	ways := pbf.Ways()
	relations := pbf.Relations()
	// TODO: Add a bbox override for custom bbox from input.
	bbox := pbf.BBox()

	shapefile := &gis.Shapefile{Filename: *shapefilePtr}

	err = shapefile.Load()

	if err != nil {
		panic(err)
	}

	polygons, err := shapefile.Clip(bbox)

	if err != nil {
		panic(err)
	}

	image := &gis.Image{
		BBox:   bbox,
		Width:  *widthPtr,
		Config: conf,
	}
	err = image.Init()

	if err != nil {
		fmt.Println(err)
		return
	}

	image.DrawShapePolygons(polygons)
	image.DrawWays(ways)
	image.DrawWays(relations)
	image.PNG(*outputPtr, canvas.DPI(600))

	filename := strings.TrimSuffix(*outputPtr, filepath.Ext(*outputPtr))
	image.SVG(fmt.Sprintf("%s.svg", filename))
}
