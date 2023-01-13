package config

type FeatureQuery struct {
	Attribute string
	Value     string
}

type FeatureStyle struct {
	Queries      []FeatureQuery
	WayIdQueries []int64        `yaml:"way_id_queries"`
	Exclude      []FeatureQuery `yaml:"exclude"`
	StrokeWidth  float64        `yaml:"stroke_width"`
	StrokeColor  string         `yaml:"stroke_color"`
	FillColor    string         `yaml:"fill_color"`
	ZIndex       int            `yaml:"z_index"`
	Dashed       bool
}

// ShouldExclude takes in a map of tags (from an OSM Way) and returns whether the style should
// be excluded or not.
func (fs *FeatureStyle) ShouldExclude(tagMap map[string]string) bool {
	for _, exclusion := range fs.Exclude {
		if v, ok := tagMap[exclusion.Attribute]; ok && v == exclusion.Value {
			return true
		}
	}

	return false
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
