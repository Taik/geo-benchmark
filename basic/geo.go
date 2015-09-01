package geo

import "github.com/davidreynolds/gos2/s2"


// RectFromLatLong returns a rect instance that encapsulates all provided points.
func RectFromLatLong(points []s2.LatLng) s2.Rect {
	rect := s2.RectFromLatLng(points[0])
	for _, point := range points[1:] {
		rect = rect.AddPoint(point)
	}
	return rect
}

// CellUnionFromLatLong returns a normalized CellUnion.
func CellUnionFromLatLong(points []s2.LatLng) s2.CellUnion {
	cells := s2.CellUnion{}
	for _, point := range points {
		cells = append(cells, s2.CellIDFromLatLng(point))
	}
	cells.Normalize()
	return cells
}