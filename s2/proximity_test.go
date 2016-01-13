package proximity_test

import (
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"sort"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	proximity "github.com/taik/geo-benchmark/s2"
	"github.com/timehop/gos2/s2"
)

// tempfile returns a temporary file path.
// One of boltdb's test helper functions.
func tempfile() string {
	f, err := ioutil.TempFile("", "proximity-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}

// mustOpenDB returns a new, open DB at a temporary location.
func mustOpenDB() *bolt.DB {
	db, err := bolt.Open(tempfile(), 0666, nil)
	if err != nil {
		panic(err)
	}
	return db
}

func newProximity() (*proximity.Proximity, error) {
	db := mustOpenDB()
	proximity, err := proximity.New(db)
	if err != nil {
		return nil, err
	}
	return proximity, nil
}

func TestAddLatLng(t *testing.T) {
	points := []s2.LatLng{
		s2.LatLngFromDegrees(40.724153, -73.992610), // nw
		s2.LatLngFromDegrees(40.720533, -73.994144), // sw
		s2.LatLngFromDegrees(40.722190, -73.986226), // ne
	}

	p, err := newProximity()
	require.NoError(t, err)

	for _, ll := range points {
		assert.NoError(t, p.AddLatlng(ll))

		key := []byte(s2.CellIDFromLatLng(ll).ToToken())

		p.DB.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("proximity"))
			v := b.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, []byte{}, v)
			return nil
		})
	}
}

func TestProximitySearch(t *testing.T) {
	points := []s2.LatLng{
		s2.LatLngFromDegrees(40.724153, -73.992610), // nw
		s2.LatLngFromDegrees(40.720533, -73.994144), // sw
		s2.LatLngFromDegrees(40.722190, -73.986226), // ne
		// Should not be included
		s2.LatLngFromDegrees(40.717009, -73.983234), // SE
	}

	expectedCellIDs := []s2.CellID{
		s2.CellIDFromLatLng(s2.LatLngFromDegrees(40.724153, -73.992610)), // nw
		s2.CellIDFromLatLng(s2.LatLngFromDegrees(40.720533, -73.994144)), // sw
		s2.CellIDFromLatLng(s2.LatLngFromDegrees(40.722190, -73.986226)), // ne
	}

	p, err := newProximity()
	require.NoError(t, err)

	p.AddLatLngs(points)

	p0 := s2.LatLngFromDegrees(40.719657, -73.996632) // SW
	p1 := s2.LatLngFromDegrees(40.723353, -73.984014) // NE

	gotCellIDs := p.Search(p0, p1)

	expected := make([]int, 0, len(expectedCellIDs))
	for _, cell := range expectedCellIDs {
		expected = append(expected, int(cell))
	}

	got := make([]int, 0, len(gotCellIDs))
	for _, cell := range gotCellIDs {
		got = append(got, int(cell))
	}

	sort.Ints(expected)
	sort.Ints(got)

	assert.Equal(t, expected, got, "Should contain the first three points")
}

func TestProximityMatch(t *testing.T) {
	points := []s2.LatLng{
		s2.LatLngFromDegrees(40.724153, -73.992610), // nw
		s2.LatLngFromDegrees(40.720533, -73.994144), // sw
		s2.LatLngFromDegrees(40.722190, -73.986226), // ne
		// Should not be included
		s2.LatLngFromDegrees(40.717009, -73.983234), // SE
	}

	p, err := newProximity()
	assert.NoError(t, err)

	p.AddLatLngs(points)

	p0 := s2.LatLngFromDegrees(40.719657, -73.996632) // SW
	p1 := s2.LatLngFromDegrees(40.723353, -73.984014) // NE

	assert.True(t, p.Match(p0, p1))
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

	p, _ := newProximity()

	centerLat, centerLng := 40.724153, -73.992610

	for i := 0; i < elemCount; i++ {
		ll := generateLatLng(centerLat, centerLng, 10)
		p.AddLatlng(ll)
	}

	p0 := s2.LatLngFromDegrees(40.719657, -73.996632) // SW
	p1 := s2.LatLngFromDegrees(40.723353, -73.984014) // NE

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		p.Search(p0, p1)
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

func benchmarkProximityMatch(b *testing.B, elemCount int) {
	b.StopTimer()

	p, _ := newProximity()

	centerLat, centerLng := 40.724153, -73.992610

	for i := 0; i < elemCount; i++ {
		ll := generateLatLng(centerLat, centerLng, 10)
		p.AddLatlng(ll)
	}

	p0 := s2.LatLngFromDegrees(40.719657, -73.996632) // SW
	p1 := s2.LatLngFromDegrees(40.723353, -73.984014) // NE

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		p.Match(p0, p1)
	}
}

func BenchmarkProximityMatch_1K(b *testing.B) {
	benchmarkProximityMatch(b, 1000)
}

func BenchmarkProximityMatch_100K(b *testing.B) {
	benchmarkProximityMatch(b, 100000)
}

func BenchmarkProximityMatch_1M(b *testing.B) {
	benchmarkProximityMatch(b, 1000000)
}
