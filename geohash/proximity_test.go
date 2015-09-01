package proximity_test

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/taik/geo-benchmark/geohash"
)

func newInMemoryDB() (*leveldb.DB, error) {
	storage := storage.NewMemStorage()
	return leveldb.Open(storage, &opt.Options{})
}

func newProximity(t testing.TB) *proximity.Proximity {
	db, err := newInMemoryDB()
	assert.NoError(t, err)
	return proximity.New(db)
}

func TestGeoKey(t *testing.T) {
	got := proximity.GeoKey(45.370000, -121.700000)
	assert.Equal(t, "c216nekg2kyz", got)
}

func TestGeoKeyPrecision(t *testing.T) {
	got := proximity.GeoKeyWithPrecision(45.37, -121.7, 12)
	assert.Equal(t, "c216nekg2kyz", got)
}

func TestAddLatLng(t *testing.T) {
	p := newProximity(t)

	err := p.AddLatLng(45.370000, -121.700000)
	assert.NoError(t, err)

	key := []byte("c216nekg2kyz")

	val, err := p.DB.Get(key, &opt.ReadOptions{})
	assert.NoError(t, err)

	assert.Equal(t, key, val)
}

func TestProximitySearch(t *testing.T) {
	p := newProximity(t)

	points := []struct {
		Lat float64
		Lng float64
	}{
		{40.752115, -73.977653},
		{40.7375526, -73.9780705},
		{40.721246, -73.998231},

		// Should not be included
		{40.7413556816633, -74.0509843826294},
	}

	for _, point := range points {
		p.AddLatLng(point.Lat, point.Lng)
	}

	got := p.RadiusSearch(40.7419302, -73.9854313, 5, 10)
	assert.Len(t, got, 3)
}

func generateLatLng(x0, y0 float64, radius int) (float64, float64) {
	radiusInDegrees := float64(radius) / 111000.0

	u := rand.Float64()
	v := rand.Float64()
	w := radiusInDegrees * math.Sqrt(u)
	t := 2 * math.Pi * v
	x := w * math.Cos(t)
	y := w * math.Sin(t)

	newX := x / math.Cos(y0)

	foundLng := newX + x0
	foundLat := y + y0

	return foundLat, foundLng
}

type point struct {
	Lat float64
	Lng float64
}

func benchmarkProximitySearch(b *testing.B, count int) {
	b.StopTimer()

	p := newProximity(b)
	var newLat, newLng float64

	for i := 0; i < count; i++ {
		newLat, newLng = generateLatLng(40.752115, -73.977653, 10)
		p.AddLatLng(newLat, newLng)
	}

	b.StartTimer()
	for n := 0; n < b.N; n++ {
		p.RadiusSearch(newLat, newLng, 8, 10)
	}
}

func BenchmarkProximitySearch_1K(b *testing.B) {
	benchmarkProximitySearch(b, 1000)
}

func BenchmarkProximitySearch_100K(b *testing.B) {
	benchmarkProximitySearch(b, 100000)
}

func BenchmarkProximitySearch_1M(b *testing.B) {
	benchmarkProximitySearch(b, 1000000)
}

func benchmarkRadiusMatch(b *testing.B, count int) {
	b.StopTimer()

	p := newProximity(b)

	var newLat, newLng float64
	for i := 0; i < count; i++ {
		newLat, newLng = generateLatLng(40.752115, -73.977653, 1)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p.RadiusMatch(newLat, newLng, 7)
	}
}

func BenchmarkRadiusMatch_1K(b *testing.B) {
	benchmarkRadiusMatch(b, 1000)
}

func BenchmarkRadiusMatch_100K(b *testing.B) {
	benchmarkRadiusMatch(b, 100000)
}

func BenchmarkRadiusMatch_1M(b *testing.B) {
	benchmarkRadiusMatch(b, 1000000)
}
