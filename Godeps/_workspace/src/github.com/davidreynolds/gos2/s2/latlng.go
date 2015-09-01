package s2

import (
	"fmt"
	"math"

	"github.com/davidreynolds/gos2/s1"
)

// LatLng represents a point on the unit sphere as a pair of angles.
type LatLng struct {
	Lat, Lng s1.Angle
}

// LatLngFromDegrees returns a LatLng for the coordinates given in degrees.
func LatLngFromDegrees(lat, lng float64) LatLng {
	return LatLng{s1.Angle(lat) * s1.Degree, s1.Angle(lng) * s1.Degree}
}

func LatLngFromRadians(lat_radians, lng_radians float64) LatLng {
	return LatLng{s1.Angle(lat_radians), s1.Angle(lng_radians)}
}

// IsValid returns true iff the LatLng is normalized, with Lat ∈ [-π/2,π/2] and Lng ∈ [-π,π].
func (ll LatLng) IsValid() bool {
	return math.Abs(ll.Lat.Radians()) <= math.Pi/2 && math.Abs(ll.Lng.Radians()) <= math.Pi
}

func (ll LatLng) String() string { return fmt.Sprintf("[%v, %v]", ll.Lat, ll.Lng) }

// Distance returns the angle between two LatLngs.
func (ll LatLng) Distance(ll2 LatLng) s1.Angle {
	// Haversine formula, as used in C++ S2LatLng::GetDistance.
	lat1, lat2 := ll.Lat.Radians(), ll2.Lat.Radians()
	lng1, lng2 := ll.Lng.Radians(), ll2.Lng.Radians()
	dlat := math.Sin(0.5 * (lat2 - lat1))
	dlng := math.Sin(0.5 * (lng2 - lng1))
	x := dlat*dlat + dlng*dlng*math.Cos(lat1)*math.Cos(lat2)
	return s1.Angle(2*math.Atan2(math.Sqrt(x), math.Sqrt(math.Max(0, 1-x)))) * s1.Radian
}

// NOTE(mikeperrow): The C++ implementation publicly exposes latitude/longitude
// functions. Let's see if that's really necessary before exposing the same functionality.

func latitude(p Point) s1.Angle {
	return s1.Angle(math.Atan2(p.Z, math.Sqrt(p.X*p.X+p.Y*p.Y))) * s1.Radian
}

func longitude(p Point) s1.Angle {
	return s1.Angle(math.Atan2(p.Y, p.X)) * s1.Radian
}

// PointFromLatLng returns an Point for the given LatLng.
func PointFromLatLng(ll LatLng) Point {
	phi := ll.Lat.Radians()
	theta := ll.Lng.Radians()
	cosphi := math.Cos(phi)
	return PointFromCoords(math.Cos(theta)*cosphi, math.Sin(theta)*cosphi, math.Sin(phi))
}

// LatLngFromPoint returns an LatLng for a given Point.
func LatLngFromPoint(p Point) LatLng {
	return LatLng{latitude(p), longitude(p)}
}

// BUG(dsymonds): The major differences from the C++ version are:
//   - normalization
