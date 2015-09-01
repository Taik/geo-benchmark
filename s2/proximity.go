package proximity

import (
	"github.com/davidreynolds/gos2/s2"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type Proximity struct {
	DB *leveldb.DB
}

func New(db *leveldb.DB) *Proximity {
	return &Proximity{DB: db}
}

func (p *Proximity) AddLatlng(point s2.LatLng) error {
	cell := s2.CellIDFromLatLng(point)
	token := []byte(cell.ToToken())
	return p.DB.Put(token, []byte{}, &opt.WriteOptions{})
}

func (p *Proximity) AddLatLngs(points []s2.LatLng) error {
	for _, ll := range points {
		err := p.AddLatlng(ll)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Proximity) Match(p0, p1 s2.LatLng) bool {
	boundRect := s2.RectFromPointPair(p0, p1)

	startToken := []byte(s2.CellIDFromLatLng(boundRect.Hi()).ToToken())
	endToken := []byte(s2.CellIDFromLatLng(boundRect.Lo()).ToToken())

	iter := p.DB.NewIterator(
		&util.Range{
			Start: startToken,
			Limit: endToken,
		},
		nil,
	)
	defer iter.Release()

	for iter.Next() {
		return true
	}
	return false
}

func (p *Proximity) Search(p0, p1 s2.LatLng) []s2.CellID {
	boundRect := s2.RectFromPointPair(p0, p1)

	startToken := []byte(s2.CellIDFromLatLng(boundRect.Hi()).ToToken())
	endToken := []byte(s2.CellIDFromLatLng(boundRect.Lo()).ToToken())

	iter := p.DB.NewIterator(
		&util.Range{
			Start: startToken,
			Limit: endToken,
		},
		nil,
	)
	defer iter.Release()

	results := []s2.CellID{}

	for iter.Next() {
		cell := s2.CellIDFromToken(string(iter.Key()))
		results = append(results, cell)
	}
	return results
}