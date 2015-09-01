package geo_test

import (
	"fmt"
	"testing"

	"github.com/TomiHiltunen/geohash-golang"
	"github.com/davidreynolds/gos2/s2"
	"github.com/nick11roberts/geozip"
	"github.com/stretchr/testify/assert"
	"github.com/taik/geo-test"
)

var topLeft = s2.LatLngFromDegrees(40.755534, -74.001743)
var bottomRight = s2.LatLngFromDegrees(40.732351, -73.984963)
var center = s2.LatLngFromDegrees(40.743700, -73.991615)

var points = []s2.LatLng{
	topLeft,
	bottomRight,
}

func TestRectFromLatLong(t *testing.T) {
	rect := geo.RectFromLatLong(points)

	assert.True(t, rect.IsValid())
	assert.True(t, rect.Contains(center))
}

func BenchmarkRectContains(b *testing.B) {
	rect := geo.RectFromLatLong(points)
	for n := 0; n < b.N; n++ {
		rect.Contains(center)
	}
}

func BenchmarkRectFromLatLong(b *testing.B) {
	for n := 0; n < b.N; n++ {
		geo.RectFromLatLong(points)
	}
}

func TestCellUnionFromLatLong(t *testing.T) {
	cells := geo.CellUnionFromLatLong(points)
	assert.True(t, cells.ContainsCellID(s2.CellIDFromLatLng(topLeft)))
	assert.True(t, cells.ContainsCellID(s2.CellIDFromLatLng(bottomRight)))
}

func TestCellUnionIntersection(t *testing.T) {
	cells := geo.CellUnionFromLatLong(points)
	cells.IntersectsCellID(s2.CellIDFromLatLng(center))
}

func BenchmarkGeohashFromLatLong(b *testing.B) {
	lat, long := 40.743700, -73.991615
	for n := 0; n < b.N; n++ {
		geohash.Encode(lat, long)
	}
}

func BenchmarkCellIDFromLatLong(b *testing.B) {
	var cell s2.CellID
	for i := 0; i < b.N; i++ {
		cell = s2.CellIDFromLatLng(center)
	}
	fmt.Printf("%v", cell.ToToken())
}

func BenchmarkGeohash(b *testing.B) {
	lat, lng := 40.743700, -73.991615
	for i := 1; i < b.N; i++ {
		geohash.Encode(lat, lng)
	}
}

func BenchmarkGeozip(b *testing.B) {
	lat, lng := 40.743700, -73.991615
	for i := 0; i < b.N; i++ {
		geozip.Encode(lat, lng, false, 18)
	}
}
