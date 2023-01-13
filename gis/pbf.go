package gis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/samber/lo"
	"github.com/wroge/wgs84"
)

type PBF struct {
	nodeMap   map[osm.NodeID]*osm.Node
	ways      []*RichWay
	relations []*RichWay
	bbox      *BBox
	Verbose   bool
}

func (pbf *PBF) Init() {
	pbf.nodeMap = make(map[osm.NodeID]*osm.Node)
	pbf.ways = make([]*RichWay, 0)
	pbf.relations = make([]*RichWay, 0)
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

			// Compute the bounding box from the nodes in the PBF file.
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

				if point := pbf.pointFromNodeID(wn.ID); point != nil {
					points = append(points, *point)
				}
			}

			newWay := &RichWay{
				Way:     way,
				NodeIDs: nodeIDs,
				Points:  points,
			}

			pbf.ways = append(pbf.ways, newWay)
		} else if t == "relation" {
			// TODO: I only support polygon relations. Should other types be supported?
			relation := o.(*osm.Relation)
			nodeIDs := make([]osm.NodeID, 0)
			points := make([]Point, 0)

			if relation.Polygon() {
				if pbf.Verbose {
					j, _ := json.Marshal(relation)
					fmt.Println(string(j))
				}

				nodes := make([]osm.WayNode, 0)

				for _, member := range relation.Members {
					if member.Type == "node" {
						nodeID := member.ElementID().NodeID()
						nodeIDs = append(nodeIDs, nodeID)
						nodes, points = pbf.updateNodesAndPoints(nodeID, nodes, points)
					} else if member.Type == "way" {
						way, found := lo.Find(pbf.ways, func(way *RichWay) bool {
							return way.Way.ID == member.ElementID().WayID()
						})

						if found {
							nodeIDs = append(nodeIDs, way.NodeIDs...)

							for _, nodeID := range way.NodeIDs {
								nodes, points = pbf.updateNodesAndPoints(nodeID, nodes, points)
							}
						}
					}
				}

				way := &osm.Way{
					ID:      osm.WayID(relation.ID),
					Visible: true,
					Nodes:   nodes,
					Tags:    relation.Tags,
				}

				newWay := &RichWay{
					Way:     way,
					NodeIDs: nodeIDs,
					Points:  points,
				}

				pbf.relations = append(pbf.relations, newWay)
			}
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

func (pbf *PBF) wayNodeFromNodeID(nodeID osm.NodeID) *osm.WayNode {
	if node, ok := pbf.nodeMap[nodeID]; ok {
		return &osm.WayNode{
			ID:  nodeID,
			Lat: node.Lat,
			Lon: node.Lon,
		}
	}

	return nil
}

func (pbf *PBF) pointFromNodeID(nodeID osm.NodeID) *Point {
	if node, ok := pbf.nodeMap[nodeID]; ok {
		x, y, _ := wgs84.LonLat().To(wgs84.WebMercator())(node.Lon, node.Lat, 0)

		return &Point{
			Lat: node.Lat,
			Lon: node.Lon,
			X:   x,
			Y:   y,
		}
	}

	return nil
}

func (pbf *PBF) updateNodesAndPoints(nodeID osm.NodeID, nodes []osm.WayNode, points []Point) ([]osm.WayNode, []Point) {
	if wayNode := pbf.wayNodeFromNodeID(nodeID); wayNode != nil {
		nodes = append(nodes, *wayNode)

		if point := pbf.pointFromNodeID(nodeID); point != nil {
			points = append(points, *point)
		}
	}

	return nodes, points
}

func (pbf *PBF) BBox() *BBox {
	return pbf.bbox
}

func (pbf *PBF) Ways() []*RichWay {
	return pbf.ways
}

func (pbf *PBF) Relations() []*RichWay {
	return pbf.relations
}
