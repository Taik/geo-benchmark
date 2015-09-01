package proximity_test

import (
	"math"
	"math/rand"
	"testing"

	"sort"

	"github.com/davidreynolds/gos2/s2"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/storage"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func newInMemoryDB() (*leveldb.DB, error) {
	storage := storage.NewMemStorage()
	return leveldb.Open(storage, &opt.Options{})
}

func TestProximitySearch(t *testing.T) {
	points := []s2.LatLng{
		s2.LatLngFromDegrees(40.724153, -73.992610), // NW
		s2.LatLngFromDegrees(40.720533, -73.994144), // SW
		s2.LatLngFromDegrees(40.722190, -73.986226), // NE
		// Should not be included
		s2.LatLngFromDegrees(40.717009, -73.983234), // SE
	}

	tokens := make([]string, 0, len(points))

	db, err := newInMemoryDB()
	assert.NoError(t, err)

	for _, ll := range points {
		cell := s2.CellIDFromLatLng(ll)

		token := []byte(cell.ToToken())
		tokens = append(tokens, string(token))
		db.Put(token, token, &opt.WriteOptions{})
	}

	bb := s2.RectFromPointPair(
		s2.LatLngFromDegrees(40.719657, -73.996632), // SW
		s2.LatLngFromDegrees(40.723353, -73.984014), // NE
	)

	start := s2.CellIDFromLatLng(bb.Hi()).ToToken()
	end := s2.CellIDFromLatLng(bb.Lo()).ToToken()

	iter := db.NewIterator(&util.Range{Start: []byte(start), Limit: []byte(end)}, nil)

	got := make([]string, 0, len(points))
	for iter.Next() {
		token := string(iter.Value())
		got = append(got, token)
	}

	iter.Release()

	sort.Strings(tokens[0:3])
	sort.Strings(got)
	assert.Equal(t, tokens[0:3], got)
	assert.NotContains(t, tokens[3], got)
}

func generateLatLng(x0, y0 float64, radius int) s2.LatLng {
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

	return s2.LatLngFromDegrees(foundLat, foundLng)
}

func benchmarkProximitySearch(b *testing.B, elemCount int) {
	b.StopTimer()

	db, _ := newInMemoryDB()

	centerLat, centerLng := 40.724153, -73.992610

	for i := 0; i < elemCount; i++ {
		ll := generateLatLng(centerLat, centerLng, 10)
		cell := s2.CellIDFromLatLng(ll)
		token := []byte(cell.ToToken())
		db.Put(token, token, &opt.WriteOptions{})
	}

	bb := s2.RectFromPointPair(
		s2.LatLngFromDegrees(40.719657, -73.996632), // SW
		s2.LatLngFromDegrees(40.723353, -73.984014), // NE
	)

	start := s2.CellIDFromLatLng(bb.Hi()).ToToken()
	end := s2.CellIDFromLatLng(bb.Lo()).ToToken()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		iter := db.NewIterator(&util.Range{Start: []byte(start), Limit: []byte(end)}, nil)

		for iter.Next() {
			iter.Value()
		}

		iter.Release()
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
