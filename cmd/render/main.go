package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/tdewolff/canvas"
	"github.com/wisepythagoras/gis-utils/config"
	"github.com/wisepythagoras/gis-utils/gis"
	"github.com/wroge/wgs84"
)

func ReadPBF(f io.Reader) ([]*gis.RichWay, *gis.BBox) {
	nodeMap := make(map[osm.NodeID]*osm.Node)
	ways := make([]*gis.RichWay, 0)

	minLat := float64(999.0)
	minLon := float64(999.0)
	maxLat := float64(-999.0)
	maxLon := float64(-999.0)

	scanner := osmpbf.New(context.Background(), f, 8)
	scanner.FilterNode = func(n *osm.Node) bool { return true }

	defer scanner.Close()
	// scanner.Object()

	for scanner.Scan() {
		o := scanner.Object()
		t := o.ObjectID().Type()

		if t == "node" {
			// Add all of the nodes to the node map so that it's easily referenced.
			node := o.(*osm.Node)
			nodeMap[node.ID] = node

			if node.Lat < minLat {
				minLat = node.Lat
			} else if node.Lat > maxLat {
				maxLat = node.Lat
			}

			if node.Lon < minLon {
				minLon = node.Lon
			} else if node.Lon > maxLon {
				maxLon = node.Lon
			}
		} else if t == "way" {
			way := o.(*osm.Way)
			nodeIDs := make([]osm.NodeID, 0)
			points := make([]gis.Point, 0)

			for _, wn := range way.Nodes {
				nodeIDs = append(nodeIDs, wn.ID)
				node := nodeMap[wn.ID]

				if node != nil {
					x, y, _ := wgs84.LonLat().To(wgs84.WebMercator())(node.Lon, node.Lat, 0)

					point := gis.Point{
						Lat: node.Lat,
						Lon: node.Lon,
						X:   x,
						Y:   y,
					}

					points = append(points, point)
				}
			}

			newWay := &gis.RichWay{
				Way:     way,
				NodeIDs: nodeIDs,
				Points:  points,
			}

			ways = append(ways, newWay)
		}
	}

	scanErr := scanner.Err()

	if scanErr != nil {
		panic(scanErr)
	}

	bbox := &gis.BBox{
		SW: gis.Point{Lat: minLat, Lon: minLon},
		NE: gis.Point{Lat: maxLat, Lon: maxLon},
	}

	return ways, bbox
}

func main() {
	shapefilePtr := flag.String("shapefile", "", "The path to the land shapefile")
	outputPtr := flag.String("output", "out.png", "The output path")
	pbfPtr := flag.String("pbf", "", "The path to the OSM Protobuf file")
	stylesPtr := flag.String("styles", "", "The path to the style configuration file")
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

	pbf := &gis.PBF{}
	pbf.Init()

	if err := pbf.Load(f); err != nil {
		panic(err)
	}

	ways := pbf.Ways()
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
		Width:  320,
		Config: conf,
	}
	err = image.Init()

	if err != nil {
		fmt.Println(err)
		return
	}

	image.DrawShapePolygons(polygons)
	image.DrawPolygons(ways)
	image.DrawLines(ways)
	image.PNG(*outputPtr, canvas.DPI(600))

	filename := strings.TrimSuffix(*outputPtr, filepath.Ext(*outputPtr))
	image.SVG(fmt.Sprintf("%s.svg", filename))
}
