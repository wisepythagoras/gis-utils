package gis

import (
	"errors"
	"image/color"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers"
	"github.com/wisepythagoras/gis-utils/config"
	"github.com/wroge/wgs84"
)

type Image struct {
	BBox      *BBox
	Width     float64
	Config    *config.Config
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

	fillColor := &color.RGBA{222, 236, 240, 255}

	if img.Config != nil {
		fillColor, _ = img.Config.GetFillColor()
	}

	// The background color is set here. This should load from a configuration file.
	context.SetFillColor(fillColor)
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

		strokeWidth := 2.0
		strokeColor := &color.RGBA{205, 205, 205, 255}
		fillColor := &color.RGBA{255, 255, 255, 255}

		if img.Config != nil {
			strokeWidth, _ = img.Config.GetLandStrokeWidth()
			strokeColor, _ = img.Config.GetLandStrokeColor()
			fillColor, _ = img.Config.GetLandFillColor()
		}

		// The color of the polygon is going to be painted here. This, also, should come from a styles
		// or configuration file.s
		img.context.SetStrokeColor(strokeColor)
		img.context.SetFillColor(fillColor)
		img.context.SetStrokeWidth(strokeWidth)
		img.context.DrawPath(0, 0, path)
	}
}

func (img *Image) DrawPolygons(ways []*RichWay) {
	for _, way := range ways {
		highway := way.Way.TagMap()["highway"]
		waterway := way.Way.TagMap()["waterway"]
		footway := way.Way.TagMap()["footway"]
		route := way.Way.TagMap()["route"]
		area := way.Way.TagMap()["area"]

		// Do not style lines here.
		if len(highway) > 0 ||
			len(route) > 0 ||
			len(waterway) > 0 ||
			(len(footway) > 0 &&
				len(area) == 0) {
			continue
		}

		var style *config.FeatureStyle

		if img.Config != nil {
			style = img.getStyleFromTags(way)
		}

		if style != nil {
			path := &canvas.Path{}

			for i, point := range way.Points {
				if i == 0 {
					path.MoveTo(point.X, point.Y)
				} else {
					path.LineTo(point.X, point.Y)
				}
			}

			strokeWidth := 0.0
			strokeColor := &color.RGBA{0, 0, 0, 255}
			fillColor := &color.RGBA{0, 0, 0, 255}

			if style.StrokeWidth > 0 {
				strokeWidth = style.StrokeWidth
			}

			if style.StrokeColor != "" {
				strokeColor, _ = config.ParseColor(style.StrokeColor)
			}

			if style.FillColor != "" {
				fillColor, _ = config.ParseColor(style.FillColor)
			}

			img.context.SetStrokeWidth(strokeWidth)
			img.context.SetStrokeColor(*strokeColor)
			img.context.SetFillColor(*fillColor)
			img.context.DrawPath(0, 0, path)
			img.context.ResetStyle()
		}
	}
}

func (img *Image) DrawLines(ways []*RichWay) {
	for _, way := range ways {
		highway := way.Way.TagMap()["highway"]
		waterway := way.Way.TagMap()["waterway"]
		footway := way.Way.TagMap()["footway"]
		route := way.Way.TagMap()["route"]
		area := way.Way.TagMap()["area"]

		if (len(highway) == 0 &&
			len(footway) == 0 &&
			len(waterway) == 0 &&
			len(route) == 0) ||
			len(area) > 0 {
			continue
		}

		path := &canvas.Path{}

		for i, point := range way.Points {
			if i == 0 {
				path.MoveTo(point.X, point.Y)
			} else {
				path.LineTo(point.X, point.Y)
			}
		}

		var style *config.FeatureStyle

		if img.Config != nil {
			style = img.getStyleFromTags(way)
		}

		if style != nil {
			strokeWidth := 4.0
			strokeColor := &color.RGBA{0, 0, 0, 255}

			if style.StrokeWidth > 0 {
				strokeWidth = style.StrokeWidth
			}

			if style.StrokeColor != "" {
				strokeColor, _ = config.ParseColor(style.StrokeColor)
			}

			img.context.SetStrokeWidth(strokeWidth)
			img.context.SetStrokeColor(*strokeColor)

			if style.Dashed {
				img.context.SetDashes(0.0, style.StrokeWidth, style.StrokeWidth)
			}
		} else if img.Config.ShowAll() {
			img.context.SetStrokeWidth(4.0)
			img.context.SetStrokeColor(color.RGBA{160, 160, 160, 255})
		}

		img.context.SetFillColor(color.Transparent)
		img.context.DrawPath(0, 0, path)
		img.context.ResetStyle()
	}
}

func (img *Image) Save(filename string, resolution canvas.Resolution) {
	renderers.Write(filename, img.mapCanvas, resolution)
}

func (img *Image) getStyleFromTags(way *RichWay) (style *config.FeatureStyle) {
	for _, tag := range way.Way.Tags {
		if tag.Key == "website" || tag.Key == "name" {
			continue
		}

		// Query the configuration for any styles that apply to the given attribute.
		tempStyle, _ := img.Config.Query(tag.Key, tag.Value)

		if tempStyle != nil {
			style = tempStyle
			break
		}
	}

	return
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
