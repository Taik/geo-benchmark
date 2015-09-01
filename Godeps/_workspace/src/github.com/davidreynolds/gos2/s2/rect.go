package s2

import (
	"fmt"
	"math"

	"github.com/davidreynolds/gos2/r1"
	"github.com/davidreynolds/gos2/s1"
)

// Rect represents a closed latitude-longitude rectangle.
type Rect struct {
	Lat r1.Interval
	Lng s1.Interval
}

var (
	validRectLatRange = r1.Interval{-math.Pi / 2, math.Pi / 2}
	validRectLngRange = s1.FullInterval()
)

// FullRect returns the full rectangle.
func FullRect() Rect { return Rect{validRectLatRange, validRectLngRange} }

func EmptyRect() Rect { return Rect{r1.EmptyInterval(), s1.EmptyInterval()} }

// RectFromLatLng constructs a rectangle containing a single point p.
func RectFromLatLng(p LatLng) Rect {
	return Rect{
		Lat: r1.Interval{p.Lat.Radians(), p.Lat.Radians()},
		Lng: s1.Interval{p.Lng.Radians(), p.Lng.Radians()},
	}
}

func RectFromPointPair(p1, p2 LatLng) Rect {
	return Rect{
		Lat: r1.IntervalFromPointPair(p1.Lat.Radians(), p2.Lat.Radians()),
		Lng: s1.IntervalFromPointPair(p1.Lng.Radians(), p2.Lng.Radians()),
	}
}

// RectFromCenterSize constructs a rectangle with the given size and center.
// center needs to be normalized, but size does not. The latitude
// interval of the result is clamped to [-90,90] degrees, and the longitude
// interval of the result is FullRect() if and only if the longitude size is
// 360 degrees or more.
//
// Examples of clamping (in degrees):
//   center=(80,170),  size=(40,60)   -> lat=[60,90],   lng=[140,-160]
//   center=(10,40),   size=(210,400) -> lat=[-90,90],  lng=[-180,180]
//   center=(-90,180), size=(20,50)   -> lat=[-90,-80], lng=[155,-155]
func RectFromCenterSize(center, size LatLng) Rect {
	half := LatLng{size.Lat / 2, size.Lng / 2}
	return RectFromLatLng(center).expanded(half)
}

// IsValid returns true iff the rectangle is valid.
// This requires Lat ⊆ [-π/2,π/2] and Lng ⊆ [-π,π], and Lat = ∅ iff Lng = ∅
func (r Rect) IsValid() bool {
	return math.Abs(r.Lat.Lo) <= math.Pi/2 &&
		math.Abs(r.Lat.Hi) <= math.Pi/2 &&
		r.Lng.IsValid() &&
		r.Lat.IsEmpty() == r.Lng.IsEmpty()
}

func (r Rect) CapBound() Cap {
	// We consider two possible bounding caps, one whose axis passes
	// through the center of the lat-lng rectangle and one whose axis
	// is the north or south pole. We return the smaller of the two caps.
	if r.IsEmpty() {
		return EmptyCap()
	}
	var poleZ, poleAngle float64
	//	var poleAngle s1.Angle
	if r.Lat.Lo+r.Lat.Hi < 0 {
		// South pole axis yields smaller cap.
		poleZ = -1
		poleAngle = math.Pi/2 + r.Lat.Hi
	} else {
		poleZ = 1
		poleAngle = math.Pi/2 - r.Lat.Lo
	}
	poleCap := CapFromCenterAngle(PointFromCoords(0, 0, poleZ), s1.Angle(poleAngle))

	// For bounding rectangles that span 180 degrees or less in longitude,
	// the maximum cap size is achieved at one of the rectangle vertices.
	// For rectangles that are larger than 180 degrees, we punt and always
	// return a bounding cap centered at one of the two poles.
	lngSpan := r.Lng.Hi - r.Lng.Lo
	if math.Remainder(lngSpan, 2*math.Pi) >= 0 {
		if lngSpan < 2*math.Pi {
			midCap := CapFromCenterAngle(PointFromLatLng(r.Center()), s1.Angle(0))
			for k := 0; k < 4; k++ {
				midCap.AddPoint(PointFromLatLng(r.Vertex(k)))
			}
			if midCap.height < poleCap.height {
				return midCap
			}
		}
	}
	return poleCap
}

func (r Rect) Vertex(k int) LatLng {
	// Twiddle bits to return the points in CCW order (SW, SE, NE, NW).
	return LatLngFromRadians(r.Lat.Bound(k>>1), r.Lng.Bound((k>>1)^(k&1)))
}

// IsEmpty reports whether the rectangle is empty.
func (r Rect) IsEmpty() bool { return r.Lat.IsEmpty() }

// IsFull reports whether the rectangle is full.
func (r Rect) IsFull() bool { return r.Lat.Equal(validRectLatRange) && r.Lng.IsFull() }

// IsPoint reports whether the rectangle is a single point.
func (r Rect) IsPoint() bool { return r.Lat.Lo == r.Lat.Hi && r.Lng.Lo == r.Lng.Hi }

// Equal makes sure two rectangles are equal.
func (r Rect) Equal(other Rect) bool { return r.Lat.Equal(other.Lat) && r.Lng.Equal(other.Lng) }

// Lo returns one corner of the rectangle.
func (r Rect) Lo() LatLng {
	return LatLng{s1.Angle(r.Lat.Lo) * s1.Radian, s1.Angle(r.Lng.Lo) * s1.Radian}
}

// Hi returns the other corner of the rectangle.
func (r Rect) Hi() LatLng {
	return LatLng{s1.Angle(r.Lat.Hi) * s1.Radian, s1.Angle(r.Lng.Hi) * s1.Radian}
}

// Center returns the center of the rectangle.
func (r Rect) Center() LatLng {
	return LatLng{s1.Angle(r.Lat.Center()) * s1.Radian, s1.Angle(r.Lng.Center()) * s1.Radian}
}

// Size returns the size of the Rect.
func (r Rect) Size() LatLng {
	return LatLng{s1.Angle(r.Lat.Length()) * s1.Radian, s1.Angle(r.Lng.Length()) * s1.Radian}
}

// Area returns the surface area of the Rect.
func (r Rect) Area() float64 {
	if r.IsEmpty() {
		return 0
	}
	capDiff := math.Abs(math.Sin(r.Lat.Hi) - math.Sin(r.Lat.Lo))
	return r.Lng.Length() * capDiff
}

// AddPoint increases the size of the rectangle to include the given point.
func (r Rect) AddPoint(ll LatLng) Rect {
	if !ll.IsValid() {
		return r
	}
	return Rect{
		Lat: r.Lat.AddPoint(ll.Lat.Radians()),
		Lng: r.Lng.AddPoint(ll.Lng.Radians()),
	}
}

func (r Rect) Contains(ll LatLng) bool {
	return r.Lat.Contains(ll.Lat.Radians()) && r.Lng.Contains(ll.Lng.Radians())
}

func (r Rect) ContainsPoint(p Point) bool {
	return r.Contains(LatLngFromPoint(p))
}

func (r Rect) ContainsRect(other Rect) bool {
	return r.Lat.ContainsInterval(other.Lat) && r.Lng.ContainsInterval(other.Lng)
}

func (r Rect) ContainsCell(cell Cell) bool {
	return r.ContainsRect(cell.RectBound())
}

func (r Rect) MayIntersect(cell Cell) bool {
	return r.Intersects(cell.RectBound())
}

func (r Rect) Intersects(other Rect) bool {
	return r.Lat.Intersects(other.Lat) && r.Lng.Intersects(other.Lng)
}

func (r Rect) Union(other Rect) Rect {
	return Rect{
		Lat: r.Lat.Union(other.Lat),
		Lng: r.Lng.Union(other.Lng),
	}
}

// expanded returns a rectangle that contains all points whose latitude distance from
// this rectangle is at most margin.Lat, and whose longitude distance from
// this rectangle is at most margin.Lng. In particular, latitudes are
// clamped while longitudes are wrapped. Any expansion of an empty rectangle
// remains empty. Both components of margin must be non-negative.
//
// Note that if an expanded rectangle contains a pole, it may not contain
// all possible lat/lng representations of that pole, e.g., both points [π/2,0]
// and [π/2,1] represent the same pole, but they might not be contained by the
// same Rect.
//
// If you are trying to grow a rectangle by a certain distance on the
// sphere (e.g. 5km), refer to the ConvolveWithCap() C++ method implementation
// instead.
func (r Rect) expanded(margin LatLng) Rect {
	return Rect{
		Lat: r.Lat.Expanded(margin.Lat.Radians()).Intersection(validRectLatRange),
		Lng: r.Lng.Expanded(margin.Lng.Radians()),
	}
}

func (r Rect) String() string { return fmt.Sprintf("[Lo%v, Hi%v]", r.Lo(), r.Hi()) }

// BUG(dsymonds): The major differences from the C++ version are:
//   - almost everything
