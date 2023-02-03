package gis

import (
	"errors"

	"github.com/jonas-p/go-shp"
)

type Shapefile struct {
	Filename string
	reader   *shp.Reader
	polygons []*ShapePolygon
}

func (shapefile *Shapefile) Load() error {
	reader, err := shp.Open(shapefile.Filename)

	if err != nil {
		return err
	}

	shapefile.reader = reader

	return nil
}

func (shapefile *Shapefile) Clip(bbox *BBox) ([]*ShapePolygon, error) {
	if shapefile.reader == nil {
		return nil, errors.New("no shapefile was loaded")
	}

	shapefile.polygons = make([]*ShapePolygon, 0)
	// convert := wgs84.LonLat().To(wgs84.WebMercator())

	for shapefile.reader.Next() {
		_, p := shapefile.reader.Shape()

		var intermediate interface{} = p
		polygon := intermediate.(*shp.Polygon)

		points := make([]Point, 0)
		insideBBox := false

		for i, point := range polygon.Points {
			var lat, lon, xmin, ymin, xmax, ymax float64
			// var lon float64

			lat = point.Y
			lon = point.X

			xmin = bbox.SW.Lon
			ymin = bbox.SW.Lat
			xmax = bbox.NE.Lon
			ymax = bbox.NE.Lat

			// lon, lat, _ = wgs84.WebMercator().To(wgs84.LonLat())(point.X, point.Y, 0)
			// fmt.Println(point.X, point.Y, lat, lon)

			// xmin, ymin, _ := convert(bbox.SW.Lon, bbox.SW.Lat, 0)
			// xmax, ymax, _ := convert(bbox.NE.Lon, bbox.NE.Lat, 0)

			if lon >= bbox.SW.Lon && lon <= bbox.NE.Lon && lat <= bbox.NE.Lat && lat >= bbox.SW.Lat {
				insideBBox = true
			}

			if point.X < xmin {
				point.X = xmin
			} else if point.X > xmax {
				point.X = xmax
			}

			if point.Y < ymin {
				point.Y = ymin
			} else if point.Y > ymax {
				point.Y = ymax
			}

			if lon < bbox.SW.Lon {
				lon = bbox.SW.Lon
			} else if lon > bbox.NE.Lon {
				lon = bbox.NE.Lon
			}

			if lat < bbox.SW.Lat {
				lat = bbox.SW.Lat
			} else if lat > bbox.NE.Lat {
				lat = bbox.NE.Lat
			}

			points = append(points, Point{Lat: lat, Lon: lon})
			polygon.Points[i] = point
		}

		if insideBBox {
			newPolygon := &ShapePolygon{
				Points: points,
				Raw:    p,
			}

			shapefile.polygons = append(shapefile.polygons, newPolygon)
		}
	}

	return shapefile.polygons, nil
}

func (shapefile *Shapefile) Iter(callback func(int, *shp.Polygon) error) error {
	if shapefile.reader == nil {
		return errors.New("no shapefile was loaded")
	}

	for shapefile.reader.Next() {
		i, p := shapefile.reader.Shape()
		var intermediate interface{} = p
		polygon := intermediate.(*shp.Polygon)

		if err := callback(i, polygon); err != nil {
			return err
		}
	}

	return nil
}

// GetPolygons just retruns the list of polygons that were captured from a shapefile.
func (shapefile *Shapefile) GetPolygons() []*ShapePolygon {
	return shapefile.polygons
}

func (shapefile *Shapefile) SaveClippedShapefile(filename string) {
	shape, _ := shp.Create(filename, shp.POLYGON)
	defer shape.Close()

	for _, polygon := range shapefile.polygons {
		shape.Write(polygon.Raw.(shp.Shape))
	}
}
