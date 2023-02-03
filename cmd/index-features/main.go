package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jonas-p/go-shp"
	"github.com/tidwall/buntdb"
	"github.com/wisepythagoras/gis-utils/gis"
	"github.com/wroge/wgs84"
)

func indexShapefile(tx *buntdb.Tx, shapefile *gis.Shapefile) error {
	return shapefile.Iter(func(i int, p *shp.Polygon) error {
		points := ""

		for j, point := range p.Points {
			lon, lat, _ := wgs84.WebMercator().To(wgs84.LonLat())(point.X, point.Y, 0)
			separator := ","

			if j == 0 {
				separator = ""
			}

			points = fmt.Sprintf("%s%s[%f %f]", points, separator, lon, lat)
			fmt.Println(i, j)
		}

		key := fmt.Sprintf("land:%d:feature", i)

		_, _, err := tx.Set(key, points, nil)

		if err != nil {
			return err
		}

		fmt.Printf("Saved feature %d with %d points.\n", i, len(p.Points))

		return nil
	})
}

func indexPbf(tx *buntdb.Tx, pbf *gis.PBF) error {
	for _, way := range pbf.Ways() {
		points := ""

		for _, ring := range way.Points {
			for j, point := range ring {
				separator := ","

				if j == 0 {
					separator = ""
				}

				points = fmt.Sprintf("%s%s[%f %f]", points, separator, point.Lon, point.Lat)
			}
		}

		key := fmt.Sprintf("land:%d:feature", way.Way.ID)

		_, _, err := tx.Set(key, points, nil)

		if err != nil {
			return err
		}

		fmt.Printf("Saved feature %d with %d points.\n", way.Way.ID, len(way.Points))
	}

	return nil
}

func main() {
	shapefilePtr := flag.String("shapefile", "", "The path to the land shapefile")
	pbfPtr := flag.String("pbf", "", "The path to the land PBF")
	flag.Parse()

	if len(*shapefilePtr) == 0 && len(*pbfPtr) == 0 {
		fmt.Println("A path to a shapefile or PBF is required (use -shapefile land.shp or -pbf area.pbf).")
		os.Exit(1)
	}

	var shapefile *gis.Shapefile
	var pbf *gis.PBF

	if len(*shapefilePtr) > 0 {
		shapefile = &gis.Shapefile{Filename: *shapefilePtr}

		if err := shapefile.Load(); err != nil {
			panic(err)
		}
	} else {
		pbf = &gis.PBF{}
		pbf.Init()

		f, err := os.Open(*pbfPtr)

		if err != nil {
			panic(err)
		}

		defer f.Close()

		if err := pbf.Load(f); err != nil {
			panic(err)
		}
	}

	target := *shapefilePtr

	if target == "" {
		target = *pbfPtr
	}

	dbPath := target
	_, filename := path.Split(target)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))
	dbPath = fmt.Sprintf("%s.db", filename)

	db, err := buntdb.Open(dbPath)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	db.CreateSpatialIndex("land", "land:*:feature", buntdb.IndexRect)

	err = db.Update(func(tx *buntdb.Tx) error {
		fmt.Println("Saving indecies.")

		if shapefile != nil {
			err = indexShapefile(tx, shapefile)
		} else if pbf != nil {
			err = indexPbf(tx, pbf)
		}

		fmt.Println("Interrated over all features.")

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		panic(err)
	}
}
