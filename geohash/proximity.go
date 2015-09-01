package proximity

import (
	"github.com/TomiHiltunen/geohash-golang"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// Proximity is a temporary struct to store data associated with a particular
// database search.
type Proximity struct {
	DB *leveldb.DB
}

// GeoKey returns the GeoHash value of the lat/lng pair.
func GeoKey(lat, lng float64) string {
	return geohash.Encode(lat, lng)
}

// GeoKeyWithPrecision returns the GeoHash value of the lat/long pair, up to a
// precision value.
//
// Precision digits and its corresponding accuracy (roughly in meters):
//	1	5,009.4km x 4,992.6km
//	2	1,252.3km x 624.1km
//	3	156.5km x 156km
//	4	39.1km x 19.5km
//	5	4.9km x 4.9km
//	6	1.2km x 609.4m
//	7	152.9m x 152.4m
//	8	38.2m x 19m
//	9	4.8m x 4.8m
//	10	1.2m x 59.5cm
//	11	14.9cm x 14.9cm
//	12	3.7cm x 1.9cm
func GeoKeyWithPrecision(lat, lng float64, precision int) string {
	return geohash.EncodeWithPrecision(lat, lng, precision)
}

func New(db *leveldb.DB) *Proximity {
	return &Proximity{DB: db}
}

func (p *Proximity) AddLatLng(lat, lng float64) error {
	key := GeoKey(lat, lng)
	return p.DB.Put([]byte(key), []byte(key), &opt.WriteOptions{})
}

// RadiusSearch returns a list of GeoHashes that intersects with the radius of the provided lat/long pair.
func (p *Proximity) RadiusSearch(lat, lng float64, precision int, maxMatches int) []string {
	key := GeoKeyWithPrecision(lat, lng, precision)
	edges := geohash.CalculateAllAdjacent(key)
	edges = append(edges, key)

	matches := make([]string, 0, maxMatches)

	for _, edge := range edges {
		itr := p.DB.NewIterator(util.BytesPrefix([]byte(edge)), nil)
		defer itr.Release()
		for itr.Next() {
			if len(matches) == maxMatches {
				return matches
			}
			matches = append(matches, string(itr.Value()))
		}
	}

	return matches
}

// RadiusMatch returns true if there are at least one point that falls within the radius of the provided
// lat/long pair.
func (p *Proximity) RadiusMatch(lat, lng float64, precision int) bool {
	key := GeoKeyWithPrecision(lat, lng, precision)
	edges := geohash.CalculateAllAdjacent(key)
	edges = append(edges, key)

	for _, edge := range edges {
		itr := p.DB.NewIterator(util.BytesPrefix([]byte(edge)), nil)
		defer itr.Release()

		for itr.Next() {
			return true
		}
	}
	return false
}

