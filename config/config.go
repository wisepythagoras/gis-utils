package config

import (
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"

	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
)

const NOT_LOADED_ERR = "no loaded styles were found"
const NO_STYLE_ERR = "no corresponding style found"

type Config struct {
	UseMap      bool
	Verbose     bool
	styleConfig *StyleConfig
	styleMap    map[string]map[string]*FeatureStyle
}

func (c *Config) ParseFile(filename string) error {
	if len(filename) == 0 {
		return errors.New("no configuration file or bytes found")
	}

	source, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	return c.Parse(source)
}

func (c *Config) Parse(source []byte) error {
	var styleConfig StyleConfig

	err := yaml.Unmarshal(source, &styleConfig)

	if err != nil {
		return err
	}

	if c.UseMap {
		styleMap := make(map[string]map[string]*FeatureStyle)

		for i, style := range styleConfig.Styles {
			for _, query := range style.Queries {
				if styleMap[query.Attribute] == nil {
					styleMap[query.Attribute] = make(map[string]*FeatureStyle)
				}

				styleMap[query.Attribute][query.Value] = &styleConfig.Styles[i]

				if c.Verbose {
					fmt.Println(query.Attribute, query.Value, style)
				}
			}
		}

		c.styleMap = styleMap
	}

	c.styleConfig = &styleConfig

	return nil
}

func (c *Config) Query(attribute, value string) (*FeatureStyle, error) {
	if c.styleConfig == nil {
		return nil, errors.New(NOT_LOADED_ERR)
	}

	if c.UseMap {
		return c.queryMap(attribute, value)
	}

	style, ok := lo.Find(c.styleConfig.Styles, func(fs FeatureStyle) bool {
		return lo.Some(fs.Queries, []FeatureQuery{{Attribute: attribute, Value: value}})
	})

	if !ok {
		return nil, errors.New(NO_STYLE_ERR)
	}

	return &style, nil
}

func (c *Config) queryMap(attribute, value string) (*FeatureStyle, error) {
	if c.styleConfig == nil {
		return nil, errors.New(NOT_LOADED_ERR)
	}

	if attrMap, ok := c.styleMap[attribute]; ok {
		if style, ok := attrMap[value]; ok {
			return style, nil
		}

		return nil, errors.New(NO_STYLE_ERR)
	} else {
		return nil, errors.New(NO_STYLE_ERR)
	}
}

func (c *Config) GetStyles() *StyleConfig {
	return c.styleConfig
}

func (c *Config) ShowAll() bool {
	if c.styleConfig == nil {
		return false
	}

	return c.styleConfig.ShowAll
}

func (c *Config) GetFillColor() (*color.RGBA, error) {
	if c.styleConfig == nil {
		return nil, errors.New(NOT_LOADED_ERR)
	}

	if c.styleConfig.FillColor == "" {
		return &color.RGBA{26, 100, 153, 255}, nil
	}

	return ParseColor(c.styleConfig.FillColor)
}

func (c *Config) GetLandFillColor() (*color.RGBA, error) {
	if c.styleConfig == nil {
		return nil, errors.New(NOT_LOADED_ERR)
	}

	if c.styleConfig.Land.FillColor == "" {
		return &color.RGBA{255, 255, 255, 255}, nil
	}

	return ParseColor(c.styleConfig.Land.FillColor)
}

func (c *Config) GetLandStrokeColor() (*color.RGBA, error) {
	if c.styleConfig == nil {
		return nil, errors.New(NOT_LOADED_ERR)
	}

	if c.styleConfig.Land.StrokeColor == "" {
		return &color.RGBA{255, 255, 255, 255}, nil
	}

	return ParseColor(c.styleConfig.Land.StrokeColor)
}

func (c *Config) GetLandStrokeWidth() (float64, error) {
	if c.styleConfig == nil {
		return 0.0, errors.New(NOT_LOADED_ERR)
	}

	return c.styleConfig.Land.StrokeWidth, nil
}
