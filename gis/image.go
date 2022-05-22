package gis

import (
	"errors"
	"image/color"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	"github.com/wroge/wgs84"
)

type Image struct {
	BBox      *BBox
	Width     float64
	mapCanvas *canvas.Canvas
	context   *canvas.Context
}

func (img *Image) Init() error {
	if img.BBox == nil {
		return errors.New("no bounding box found")
	}

	// Here we convert the WGS84 projection to Webmercator, because WGS84 looks a little weird when
	// drawn on the map.
	convert := wgs84.LonLat().To(wgs84.WebMercator())
	xmin, ymin, _ := convert(img.BBox.SW.Lon, img.BBox.SW.Lat, 0)
	xmax, ymax, _ := convert(img.BBox.NE.Lon, img.BBox.NE.Lat, 0)

	height := img.Width * (ymax - ymin) / (xmax - xmin)
	mapCanvas := canvas.New(img.Width, height)
	context := canvas.NewContext(mapCanvas)

	// The background color is set here. This should load from a configuration file.
	context.SetFillColor(color.RGBA{222, 236, 240, 255})
	context.DrawPath(0.0, 0.0, canvas.Rectangle(img.Width, height))

	// Set the coordinate scaling, so that we can just start adding points from our shapefiles and
	// protobuf files with the Webmercator projection.
	xscale := img.Width / (xmax - xmin)
	yscale := height / (ymax - ymin)
	context.SetView(canvas.Identity.Translate(0.0, 0.0).Scale(xscale, yscale).Translate(-xmin, -ymin))

	img.context = context
	img.mapCanvas = mapCanvas

	return nil
}

func (img *Image) DrawShapePolygons(polygons []*ShapePolygon) {
	convert := wgs84.LonLat().To(wgs84.WebMercator())

	img.context.SetStrokeWidth(2.0)

	for _, polygon := range polygons {
		path := &canvas.Path{}

		for i, point := range polygon.Points {
			// Change the projection before creating any shapes on the image.
			X, Y, _ := convert(point.Lon, point.Lat, 0)

			if i == 0 {
				path.MoveTo(X, Y)
			} else {
				path.LineTo(X, Y)
			}
		}

		path.Close()

		// The color of the polygon is going to be painted here. This, also, should come from a styles
		// or configuration file.s
		img.context.SetStrokeColor(color.RGBA{205, 205, 205, 255})
		img.context.SetFillColor(color.RGBA{255, 255, 255, 255})
		img.context.DrawPath(0, 0, path)
	}
}

func (img *Image) DrawLines(lines []*RichWay) {
	for _, line := range lines {
		highway := line.Way.TagMap()["highway"]
		area := line.Way.TagMap()["area"]

		if len(highway) == 0 || len(area) > 0 {
			continue
		}

		path := &canvas.Path{}

		for i, point := range line.Points {
			if i == 0 {
				path.MoveTo(point.X, point.Y)
			} else {
				path.LineTo(point.X, point.Y)
			}
		}

		img.context.SetStrokeWidth(4.0)

		if highway == "footway" {
			img.context.SetStrokeColor(color.RGBA{34, 0, 255, 255})
		} else if highway == "path" {
			img.context.SetStrokeColor(color.RGBA{34, 255, 34, 255})
		} else if highway == "residential" {
			img.context.SetStrokeColor(color.RGBA{255, 100, 34, 255})
		} else if highway == "pedestrian" {
			img.context.SetStrokeColor(color.RGBA{255, 0, 0, 255})
		} else {
			img.context.SetStrokeColor(color.RGBA{34, 34, 34, 255})
		}

		img.context.SetFillColor(color.Transparent)
		img.context.DrawPath(0, 0, path)
	}
}

func (img *Image) Save(filename string, resolution canvas.Resolution) {
	renderers.Write(filename, img.mapCanvas, resolution)
}

// func geoJSON(lat, lon float64) {
// 	delta := 0.0201

// 	bounds := &osm.Bounds{
// 		MinLat: lat - delta, MaxLat: lat + delta,
// 		MinLon: lon - delta, MaxLon: lon + delta,
// 	}

// 	ctx := context.Background()
// 	o, _ := osmapi.Map(ctx, bounds)

// 	fc, err := osmgeojson.Convert(o)

// 	gj, _ := json.MarshalIndent(fc, "", " ")
// 	fmt.Println(string(gj), err)
// }
