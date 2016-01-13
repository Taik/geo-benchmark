package proximity

import (
	"bytes"

	"github.com/boltdb/bolt"
	"github.com/timehop/gos2/s2"
)

type Proximity struct {
	DB *bolt.DB
}

func New(db *bolt.DB) (*Proximity, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("proximity"))
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Proximity{DB: db}, nil
}

func (p *Proximity) AddLatlng(point s2.LatLng) error {
	cell := s2.CellIDFromLatLng(point)
	token := []byte(cell.ToToken())

	return p.DB.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("proximity"))
		err := bucket.Put(token, []byte{})
		return err
	})
}

func (p *Proximity) AddLatLngs(points []s2.LatLng) error {
	var cell s2.CellID
	var token []byte

	return p.DB.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("proximity"))

		for _, ll := range points {
			cell = s2.CellIDFromLatLng(ll)
			token = []byte(cell.ToToken())

			err := bucket.Put(token, []byte{})
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *Proximity) Match(p0, p1 s2.LatLng) bool {
	boundRect := s2.RectFromLatLng(p0)
	boundRect = boundRect.AddPoint(p1)

	startToken := []byte(s2.CellIDFromLatLng(boundRect.Hi()).ToToken())
	endToken := []byte(s2.CellIDFromLatLng(boundRect.Lo()).ToToken())

	found := false
	p.DB.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("proximity")).Cursor()
		for k, _ := cursor.Seek(startToken); k != nil && bytes.Compare(k, endToken) <= 0; k, _ = cursor.Next() {
			found = true
			break
		}
		return nil
	})
	return found
}

func (p *Proximity) Search(p0, p1 s2.LatLng) []s2.CellID {
	boundRect := s2.RectFromLatLng(p0)
	boundRect = boundRect.AddPoint(p1)

	startToken := []byte(s2.CellIDFromLatLng(boundRect.Hi()).ToToken())
	endToken := []byte(s2.CellIDFromLatLng(boundRect.Lo()).ToToken())

	results := []s2.CellID{}
	p.DB.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket([]byte("proximity")).Cursor()
		for k, _ := cursor.Seek(startToken); k != nil && bytes.Compare(k, endToken) <= 0; k, _ = cursor.Next() {
			cell := s2.CellIDFromToken(string(k))
			results = append(results, cell)
		}
		return nil
	})
	return results
}
