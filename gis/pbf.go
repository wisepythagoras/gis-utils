package gis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/wroge/wgs84"
)

type PBF struct {
	nodeMap map[osm.NodeID]*osm.Node
	ways    []*RichWay
	bbox    *BBox
	Verbose bool
}

func (pbf *PBF) Init() {
	pbf.nodeMap = make(map[osm.NodeID]*osm.Node)
	pbf.ways = make([]*RichWay, 0)
}

func (pbf *PBF) Load(f io.Reader) error {
	minLat := math.Inf(1)
	minLon := math.Inf(1)
	maxLat := math.Inf(-1)
	maxLon := math.Inf(-1)

	scanner := osmpbf.New(context.Background(), f, 8)
	scanner.FilterNode = func(n *osm.Node) bool { return true }

	defer scanner.Close()

	for scanner.Scan() {
		o := scanner.Object()
		t := o.ObjectID().Type()

		if t == "node" {
			// Add all of the nodes to the node map so that it's easily referenced.
			node := o.(*osm.Node)
			pbf.nodeMap[node.ID] = node

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
			points := make([]Point, 0)

			if pbf.Verbose {
				j, _ := json.Marshal(way)
				fmt.Println(string(j))
			}

			for _, wn := range way.Nodes {
				nodeIDs = append(nodeIDs, wn.ID)
				node := pbf.nodeMap[wn.ID]

				if node != nil {
					x, y, _ := wgs84.LonLat().To(wgs84.WebMercator())(node.Lon, node.Lat, 0)

					point := Point{
						Lat: node.Lat,
						Lon: node.Lon,
						X:   x,
						Y:   y,
					}

					points = append(points, point)
				}
			}

			newWay := &RichWay{
				Way:     way,
				NodeIDs: nodeIDs,
				Points:  points,
			}

			pbf.ways = append(pbf.ways, newWay)
		}
	}

	err := scanner.Err()

	if err != nil {
		return err
	}

	pbf.bbox = &BBox{
		SW: Point{Lat: minLat, Lon: minLon},
		NE: Point{Lat: maxLat, Lon: maxLon},
	}

	return nil
}

func (pbf *PBF) BBox() *BBox {
	return pbf.bbox
}

func (pbf *PBF) Ways() []*RichWay {
	return pbf.ways
}
