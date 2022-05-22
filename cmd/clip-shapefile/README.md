# clip-shapefile

This is a utility that creates a new shapefile with all the polygon features that are contained within a specific bounding box. I built this so that I can clip parts of the [land polygons](https://osmdata.openstreetmap.de/data/land-polygons.html) shapefiles, which were created by the OSM team.

You can find a bounding box through [here](http://bboxfinder.com) (but the coordinate pairs should be inverted).

## Example Usage

``` sh
./clip-shapefile -shapefile /path/to/land_polygons.shp -bbox "-12.557373,30.168876,-19.616089,26.775039"
```

The output will be:

```
566 features found within the bounding box.
The clipped shapefile was saved as land_polygons_-12.557373,30.168876,-19.616089,26.775039.shp
```
