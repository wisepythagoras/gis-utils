package gis

import (
	"bytes"
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

// DrawShapePolygons draws polygons found in the land shapefile.
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

func (img *Image) DrawWays(ways []*RichWay) {
	for _, way := range ways {
		var style *config.FeatureStyle

		if img.Config != nil {
			style = img.getStyleFromTags(way)
		}

		if style == nil {
			continue
		}

		path := &canvas.Path{}

		img.context.SetFillColor(color.Transparent)
		img.context.SetStrokeColor(color.Transparent)

		for _, ring := range way.Points {
			for i, point := range ring {
				if i == 0 {
					path.MoveTo(point.X, point.Y)
				} else {
					path.LineTo(point.X, point.Y)
				}
			}
		}

		strokeWidth := 0.0
		strokeColor := &color.RGBA{0, 0, 0, 0}
		fillColor := &color.RGBA{0, 0, 0, 0}

		if style.StrokeWidth > 0 {
			strokeWidth = style.StrokeWidth
		}

		if style.StrokeColor != "" {
			strokeColor, _ = config.ParseColor(style.StrokeColor)
		}

		if style.FillColor != "" {
			fillColor, _ = config.ParseColor(style.FillColor)
		}

		if style.Dashed {
			img.context.SetDashes(0.0, style.StrokeWidth, style.StrokeWidth)
		}

		img.context.SetStrokeWidth(strokeWidth)
		img.context.SetStrokeColor(*strokeColor)
		img.context.SetFillColor(*fillColor)
		img.context.SetZIndex(style.ZIndex)
		img.context.DrawPath(0, 0, path)
		img.context.ResetStyle()
	}
}

func (img *Image) PNG(filename string, resolution canvas.Resolution) error {
	return renderers.Write(filename, img.mapCanvas, resolution)
}

func (img *Image) SVG(filename string) error {
	return img.mapCanvas.WriteFile(filename, renderers.SVG())
}

func (img *Image) getImageBytes(writer canvas.Writer) ([]byte, error) {
	var b bytes.Buffer

	if err := writer(&b, img.mapCanvas); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (img *Image) SVGBytes() ([]byte, error) {
	return img.getImageBytes(renderers.SVG())
}

func (img *Image) PNGBytes() ([]byte, error) {
	return img.getImageBytes(renderers.PNG())
}

func (img *Image) TIFFBytes() ([]byte, error) {
	return img.getImageBytes(renderers.TIFF())
}

func (img *Image) getStyleFromTags(way *RichWay) (style *config.FeatureStyle) {
	tagMap := make(map[string]string)

	for _, tag := range way.Way.Tags {
		tagMap[tag.Key] = tag.Value
	}

	// First we look through the list of way ids (if there are any) in the styles. If a style is
	// found, then we can check if it should be excluded.
	style, _ = img.Config.QueryId(int64(way.Way.ID))

	if style != nil && style.ShouldExclude(tagMap, way.Way.ID) {
		style = nil
		return
	} else if style != nil {
		// If it shouldn't be excluded, then return it.
		return
	}

	// Otherwise, if no style was found from the way id, then we should loop through all the tags
	// (attributes) and look for any style based on that.
	for _, tag := range way.Way.Tags {
		if tag.Key == "website" || tag.Key == "name" {
			continue
		}

		// Query the configuration for any styles that apply to the given attribute.
		tempStyle, _ := img.Config.Query(tag.Key, tag.Value)

		if tempStyle != nil {
			if tempStyle.ShouldExclude(tagMap, way.Way.ID) {
				continue
			}

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
