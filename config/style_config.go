package config

import (
	"github.com/paulmach/osm"
	"github.com/samber/lo"
)

type FeatureQuery struct {
	Attribute string
	Value     string
}

type FeatureStyle struct {
	Queries       []FeatureQuery
	WayIdQueries  []int64        `yaml:"way_id_queries"`
	WayIdExcludes []int64        `yaml:"way_id_excludes"`
	Exclude       []FeatureQuery `yaml:"exclude"`
	StrokeWidth   float64        `yaml:"stroke_width"`
	StrokeColor   string         `yaml:"stroke_color"`
	FillColor     string         `yaml:"fill_color"`
	ZIndex        int            `yaml:"z_index"`
	Dashed        bool
}

// ShouldExclude takes in a map of tags (from an OSM Way) and returns whether the style should
// be excluded or not.
func (fs *FeatureStyle) ShouldExclude(tagMap map[string]string, wayID osm.WayID) bool {
	for _, exclusion := range fs.Exclude {
		if v, ok := tagMap[exclusion.Attribute]; ok && v == exclusion.Value {
			return true
		}
	}

	_, ok := lo.Find(fs.WayIdExcludes, func(id int64) bool {
		return id == int64(wayID)
	})

	return ok
}

type LandStyle struct {
	FillColor   string  `yaml:"fill_color"`
	StrokeWidth float64 `yaml:"stroke_width"`
	StrokeColor string  `yaml:"stroke_color"`
}

type StyleConfig struct {
	FillColor string `yaml:"fill_color"`
	Land      LandStyle
	ShowAll   bool `yaml:"show_all"`
	Styles    []FeatureStyle
}
