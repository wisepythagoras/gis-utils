package config

import (
	"errors"
	"fmt"
	"image/color"
	"regexp"
)

// ParseHexColor returns a color type which was parsed from a raw hex color
// string. Adapted from here: https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
func ParseHexColor(hexColor string) (*color.RGBA, error) {
	c := &color.RGBA{}
	c.A = 255

	if hexColor[0] != '#' {
		return c, errors.New("invalid color string")
	}

	if len(hexColor) == 7 {
		_, err := fmt.Sscanf(hexColor, "#%02x%02x%02x", &c.R, &c.G, &c.B)
		return c, err
	} else if len(hexColor) == 4 {
		_, err := fmt.Sscanf(hexColor, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		c.R *= 17
		c.G *= 17
		c.B *= 17
		return c, err
	} else {
		return nil, errors.New("invalid hex color length (must be 6 or 3)")
	}
}

// ParseRGBAColor parses an rgba color string.
func ParseRGBAColor(rgbaColor string) (*color.RGBA, error) {
	re := regexp.MustCompile(`^rgba\(\d+, \d+, \d+, \d+\)`)

	if len(re.FindString(rgbaColor)) == 0 {
		return nil, errors.New("invalid string color")
	}

	c := &color.RGBA{}
	_, err := fmt.Sscanf(rgbaColor, "rgba(%d, %d, %d, %d)", &c.R, &c.G, &c.B, &c.A)

	return c, err
}

func ParseColor(colorStr string) (*color.RGBA, error) {
	if colorStr[0] == '#' {
		return ParseHexColor(colorStr)
	} else {
		return ParseRGBAColor(colorStr)
	}
}
