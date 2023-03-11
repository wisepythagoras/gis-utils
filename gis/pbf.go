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
				Points:  [][]Point{points},
			}

			pbf.ways = append(pbf.ways, newWay)
		} else if t == "relation" {
			// TODO: I only support polygon relations. Should other types be supported?
			relation := o.(*osm.Relation)
			nodeIDs := make([]osm.NodeID, 0)

			if relation.Polygon() && relation.Visible {
				if pbf.Verbose {
					j, _ := json.Marshal(relation)
					fmt.Println(string(j))
				}

				nodes := make([]osm.WayNode, 0)
				rings := make([][]Point, 0)
				outerRings := make([][]Point, 0)
				outer := make([]Point, 0)

				sortedMembers, wayMap := pbf.sortRelationMembers(relation.Members)

				if pbf.Verbose || relation.ID == 166150 {
					j, _ := json.Marshal(sortedMembers)
					fmt.Println(" ->", len(sortedMembers), len(relation.Members), string(j))
					j, _ = json.Marshal(relation.Members)
					fmt.Println("   ", string(j))
				}

				for _, member := range sortedMembers {
					points := make([]Point, 0)

					if way, found := wayMap[member.ElementID().WayID()]; found {
						if pbf.Verbose {
							j, _ := json.Marshal(way)
							fmt.Println("  ->", string(j))
						}

						nodeIDs = append(nodeIDs, way.NodeIDs...)

						for _, nodeID := range way.NodeIDs {
							nodes, points = pbf.updateNodesAndPoints(nodeID, nodes, points)
						}
					}

					if member.Role == "outer" {
						outer = append(outer, points...)
					} else {
						rings = append(rings, points)
					}
				}

				// Leaving this here for reference. It's the old and dumb way of placing relations as polygons.
				// for _, member := range relation.Members {
				// 	points := make([]Point, 0)

				// 	if member.Type == "node" {
				// 		nodeID := member.ElementID().NodeID()
				// 		nodeIDs = append(nodeIDs, nodeID)
				// 		nodes, points = pbf.updateNodesAndPoints(nodeID, nodes, points)

				// 		if pbf.Verbose {
				// 			fmt.Println("  ->", nodeID, member.Lat, member.Lon)
				// 		}
				// 	} else if member.Type == "way" {
				// 		way, found := lo.Find(pbf.ways, func(way *RichWay) bool {
				// 			return way.Way.ID == member.ElementID().WayID()
				// 		})

				// 		if found {
				// 			if pbf.Verbose {
				// 				j, _ := json.Marshal(way)
				// 				fmt.Println("  ->", string(j))
				// 			}

				// 			nodeIDs = append(nodeIDs, way.NodeIDs...)

				// 			for _, nodeID := range way.NodeIDs {
				// 				nodes, points = pbf.updateNodesAndPoints(nodeID, nodes, points)
				// 			}
				// 		}
				// 	}

				// 	if member.Role == "outer" {
				// 		outer = append(outer, points...)
				// 	} else {
				// 		rings = append(rings, points)
				// 	}
				// }

				if len(outer) > 0 {
					temp := make([]Point, 0)

					for _, p := range outer {
						temp = append(temp, p)

						if len(temp) > 2 {
							first := temp[0]
							last := temp[len(temp)-1]

							if first.Lat == last.Lat && first.Lon == last.Lon {
								outerRings = append(outerRings, temp)
								temp = make([]Point, 0)
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
					Points:  append(outerRings, rings...),
				}

				if pbf.Verbose {
					j, _ := json.Marshal(newWay)
					fmt.Println(" ->", string(j))
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

func (pbf *PBF) findWay(wayMap map[osm.WayID]*RichWay, r osm.Member) (*RichWay, map[osm.WayID]*RichWay) {
	var way *RichWay = nil
	found := false

	if way, found = wayMap[r.ElementID().WayID()]; !found {
		way, found = lo.Find(pbf.ways, func(way *RichWay) bool {
			return way.Way.ID == r.ElementID().WayID()
		})

		if !found {
			return nil, wayMap
		}

		wayMap[way.Way.ID] = way
	}

	return way, wayMap
}

// TODO: This probably needs some tweaking, because some relations still don't show.
func (pbf *PBF) sortRelationMembers(members osm.Members) (osm.Members, map[osm.WayID]*RichWay) {
	var member *osm.Member = nil
	var memberWay *RichWay = nil
	outerMembers := []osm.Member{}
	innerMembers := []osm.Member{}
	wayMap := make(map[osm.WayID]*RichWay)
	isFirst := true

	remaining := append(osm.Members{}, members...)

	for i := 0; i < len(members); i++ {
		m := members[i]

		if m.Type != "way" {
			continue
		} else if m.Role != "outer" {
			innerMembers = append(innerMembers, m)
			_, wayMap = pbf.findWay(wayMap, m)
			continue
		}

		// Here we're looking at outer ways. These are what might need sorting, because sometimes the
		// ways that comprise them will not be implemented in order or clockwise.

		if member == nil {
			member = &m
			remaining = append(osm.Members{}, members[i+1:]...)
			// Nil check here?
			memberWay, wayMap = pbf.findWay(wayMap, m)
			outerMembers = append(outerMembers, m)
			continue
		}

		newRemaining := make([]osm.Member, 0)

		for j := 0; j < len(remaining); j++ {
			r := remaining[j]
			var way *RichWay = nil

			if r.Type != "way" {
				continue
			}

			way, wayMap = pbf.findWay(wayMap, r)

			if way == nil {
				continue
			} else if r.Role != "outer" {
				innerMembers = append(innerMembers, r)
				_, wayMap = pbf.findWay(wayMap, r)
				continue
			}

			first := way.Points[0][0]
			last := way.Points[0][len(way.Points[0])-1]
			mFirst := memberWay.Points[0][0]
			mLast := memberWay.Points[0][len(memberWay.Points[0])-1]

			if (mLast.Lat == first.Lat && mLast.Lon == first.Lon ||
				mLast.Lat == last.Lat && mLast.Lon == last.Lon) || ((mFirst.Lat == first.Lat && mFirst.Lon == first.Lon ||
				mFirst.Lat == last.Lat && mFirst.Lon == last.Lon) && !isFirst) {
				member = &r

				if (mFirst.Lat == first.Lat && mFirst.Lon == first.Lon ||
					mFirst.Lat == last.Lat && mFirst.Lon == last.Lon) && !isFirst {
					way.Points[0] = lo.Reverse(way.Points[0])
				}

				memberWay = way
				outerMembers = append(outerMembers, r)
				newRemaining = append(newRemaining, remaining[j+1:]...)
				isFirst = false
				break
			} else {
				newRemaining = append(newRemaining, r)
			}
		}

		remaining = newRemaining
	}

	sorted := append(outerMembers, innerMembers...)
	remaining = lo.Filter(remaining, func(m osm.Member, _ int) bool {
		_, found := lo.Find(sorted, func(sm osm.Member) bool {
			return m.Ref == sm.Ref
		})

		return m.Type != "node" && !found
	})

	return append(sorted, remaining...), wayMap
}
