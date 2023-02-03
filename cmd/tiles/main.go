package main

import (
	"fmt"
	"math"

	"github.com/paulmach/orb/maptile"
	"github.com/wisepythagoras/gis-utils/gis"
)

func deg2rad(d float64) float64 {
	return d * math.Pi / 180.0
}

// https://stackoverflow.com/questions/65494988/get-map-tiles-bounding-box
// https://stackoverflow.com/questions/62908635/leaflet-latlng-coordinates-to-xy-map
func getTileURL(lat, lon float64, zoom int) (int, float64, float64) {
	var xtile = (math.Floor((lon + 180) / 360 * float64((int(1) << zoom))))
	var ytile = (math.Floor((1 - math.Log(math.Tan(deg2rad(lat))+1/math.Cos(deg2rad(lat)))/math.Pi) / 2 * float64((int(1) << zoom))))
	return zoom, xtile, ytile
}

func main() {
	lon := 9.0
	lat := 52.0

	z0, x0, y0 := getTileURL(lat, lon, 10)
	z1, x1, y1 := getTileURL(lat, lon, 13)
	z2, x2, y2 := getTileURL(lat, lon, 18)

	tile := maptile.At([2]float64{lon, lat}, 10)

	fmt.Println(z0, x0, y0)
	fmt.Println(z1, x1, y1)
	fmt.Println(z2, x2, y2)

	fmt.Println(tile.Z, tile.X, tile.Y)

	tileBBox := gis.GetTileBBox(74774, 50967, 17)
	geoJSON, err := tileBBox.ToGeoJSONStr()

	fmt.Println(string(geoJSON), err)

	tileBBox = gis.GetTileBBox(uint32(x1), uint32(y1), uint32(z1))
	geoJSON, err = tileBBox.ToGeoJSONStr()

	fmt.Println(string(geoJSON), err)
}
