package gis

import (
	"github.com/paulmach/osm"
)

type RichWay struct {
	Way     *osm.Way
	NodeIDs []osm.NodeID
	Points  []Point
}
