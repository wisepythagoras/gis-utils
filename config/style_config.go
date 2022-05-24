package config

type FeatureQuery struct {
	Attribute string
	Value     string
}

type FeatureStyle struct {
	Queries     []FeatureQuery
	StrokeWidth float64 `yaml:"stroke_width"`
	StrokeColor string  `yaml:"stroke_color"`
	FillColor   string  `yaml:"fill_color"`
}

type LandStyle struct {
	FillColor   string  `yaml:"fill_color"`
	StrokeWidth float64 `yaml:"stroke_width"`
	StrokeColor string  `yaml:"stroke_color"`
}

type StyleConfig struct {
	FillColor string `yaml:"fill_color"`
	Land      LandStyle
	Styles    []FeatureStyle
}
