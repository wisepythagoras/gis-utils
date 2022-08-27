package gis

import (
	"encoding/json"
	"math"

	"github.com/tomchavakis/geojson"
	"github.com/tomchavakis/geojson/feature"
	"github.com/tomchavakis/geojson/geometry"
)

type BBox struct {
	SW Point
	NE Point
}

func (b *BBox) ToGeoJSONStr() ([]byte, error) {
	poly := geometry.Geometry{
		GeoJSONType: geojson.Polygon,
		Coordinates: [][][]float64{
			{
				{
					b.SW.Lon,
					b.SW.Lat,
				},
				{
					b.SW.Lon,
					b.NE.Lat,
				},
				{
					b.NE.Lon,
					b.NE.Lat,
				},
				{
					b.NE.Lon,
					b.SW.Lat,
				},
				{
					b.SW.Lon,
					b.SW.Lat,
				},
			},
		},
	}

	f, err := feature.New(poly, []float64{}, nil, "")

	if err != nil {
		return nil, err
	}

	return json.Marshal(f)
}

// Adapted from: https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
func GetTileBBox(x, y, z uint32) *BBox {
	bbox := &BBox{}

	bbox.SW = Point{
		Lat: Tile2Lat(y+1, z),
		Lon: Tile2Lon(x, z),
	}
	bbox.NE = Point{
		Lat: Tile2Lat(y, z),
		Lon: Tile2Lon(x+1, z),
	}

	return bbox
}

func Tile2Lon(x, z uint32) float64 {
	return float64(x)/math.Pow(2.0, float64(z))*360.0 - 180
}

func Tile2Lat(y, z uint32) float64 {
	n := math.Pi - (2.0*math.Pi*float64(y))/math.Pow(2.0, float64(z))
	return (180 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n))))
}
