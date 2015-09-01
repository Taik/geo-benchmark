package s2

import (
	"math"
	"testing"

	"github.com/davidreynolds/gos2/s1"
)

const (
	epsilon = 1e-14
	tinyRad = 1e-10
)

var (
	empty      = EmptyCap()
	full       = FullCap()
	defaultCap = EmptyCap()

	xAxisPt = PointFromCoords(1, 0, 0)
	yAxisPt = PointFromCoords(0, 1, 0)

	xAxis = CapFromPoint(xAxisPt)
	yAxis = CapFromPoint(yAxisPt)
	xComp = xAxis.Complement()

	hemi    = CapFromCenterHeight(Point{PointFromCoords(1, 0, 1).Normalize()}, 1)
	concave = CapFromCenterAngle(PointFromLatLng(LatLngFromDegrees(80, 10)),
		s1.Angle(150.0)*s1.Degree)
	tiny = CapFromCenterAngle(Point{PointFromCoords(1, 2, 3).Normalize()},
		s1.Angle(tinyRad))
)

func TestCapBasicEmptyFullValid(t *testing.T) {
	tests := []struct {
		got                Cap
		empty, full, valid bool
	}{
		{Cap{}, false, false, false},

		{empty, true, false, true},
		{empty.Complement(), false, true, true},
		{full, false, true, true},
		{full.Complement(), true, false, true},
		{defaultCap, true, false, true},

		{xComp, false, true, true},
		{xComp.Complement(), true, false, true},

		{tiny, false, false, true},
		{concave, false, false, true},
		{hemi, false, false, true},
		{tiny, false, false, true},
	}
	for _, test := range tests {
		if e := test.got.IsEmpty(); e != test.empty {
			t.Errorf("%v.IsEmpty() = %t; want %t", test.got, e, test.empty)
		}
		if f := test.got.IsFull(); f != test.full {
			t.Errorf("%v.IsFull() = %t; want %t", test.got, f, test.full)
		}
		if v := test.got.IsValid(); v != test.valid {
			t.Errorf("%v.IsValid() = %t; want %t", test.got, v, test.valid)
		}
	}
}

func TestCapCenterHeightRadius(t *testing.T) {
	if !xAxis.ApproxEqual(xAxis.Complement().Complement()) {
		t.Errorf("the double complement should equal itself, %v == %v",
			xAxis, xAxis.Complement().Complement())
	}

	if full.height != fullHeight {
		t.Error("full Caps should be full height")
	}
	if full.Radius().Degrees() != 180.0 {
		t.Error("radius of x-axis cap should be 180 degrees")
	}

	if empty.center != defaultCap.center {
		t.Error("empty Caps should be have the same center as the default")
	}
	if empty.height != defaultCap.height {
		t.Error("empty Caps should be have the same height as the default")
	}

	if yAxis.height != zeroHeight {
		t.Error("y-axis cap should not be empty height")
	}

	if xAxis.height != zeroHeight {
		t.Error("x-axis cap should not be empty height")
	}
	if xAxis.Radius().Radians() != zeroHeight {
		t.Errorf("radius of x-axis cap got %f want %f", xAxis.Radius().Radians(), emptyHeight)
	}

	hc := Point{hemi.center.Mul(-1.0)}
	if hc != hemi.Complement().center {
		t.Error("hemi center and its complement should have the same center")
	}
	if hemi.height != 1.0 {
		t.Error("hemi cap should be 1.0 in height")
	}
}

func TestCapContains(t *testing.T) {
	tests := []struct {
		c1, c2 Cap
		want   bool
	}{
		{empty, empty, true},
		{full, empty, true},
		{full, full, true},
		{empty, xAxis, false},
		{full, xAxis, true},
		{xAxis, full, false},
		{xAxis, xAxis, true},
		{xAxis, empty, true},
		{hemi, tiny, true},
		{hemi, CapFromCenterAngle(xAxisPt, s1.Angle(math.Pi/4-epsilon)), true},
		{hemi, CapFromCenterAngle(xAxisPt, s1.Angle(math.Pi/4+epsilon)), false},
		{concave, hemi, true},
		{concave, CapFromCenterHeight(Point{concave.center.Mul(-1.0)}, 0.1), false},
	}
	for _, test := range tests {
		if got := test.c1.Contains(test.c2); got != test.want {
			t.Errorf("%v.Contains(%v) = %t; want %t", test.c1, test.c2, got, test.want)
		}
	}
}

func TestCapContainsPoint(t *testing.T) {
	tangent := tiny.center.Cross(PointFromCoords(3, 2, 1).Vector).Normalize()
	tests := []struct {
		c    Cap
		p    Point
		want bool
	}{
		{xAxis, xAxisPt, true},
		{xAxis, PointFromCoords(1, 1e-20, 0), false},
		{yAxis, xAxis.center, false},
		{xComp, xAxis.center, true},
		{xComp.Complement(), xAxis.center, false},
		{tiny, Point{tiny.center.Add(tangent.Mul(tinyRad * 0.99))}, true},
		{tiny, Point{tiny.center.Add(tangent.Mul(tinyRad * 1.01))}, false},
		{hemi, Point{PointFromCoords(1, 0, -(1 - epsilon)).Normalize()}, true},
		{hemi, xAxisPt, true},
		{hemi.Complement(), xAxisPt, false},
		{concave, PointFromLatLng(LatLngFromDegrees(-70*(1-epsilon), 10)), true},
		{concave, PointFromLatLng(LatLngFromDegrees(-70*(1+epsilon), 10)), false},
		{concave, PointFromLatLng(LatLngFromDegrees(-50*(1-epsilon), -170)), true},
		{concave, PointFromLatLng(LatLngFromDegrees(-50*(1+epsilon), -170)), false},
	}
	for _, test := range tests {
		if got := test.c.ContainsPoint(test.p); got != test.want {
			t.Errorf("%v.ContainsPoint(%v) = %t, want %t", test.c, test.p, got, test.want)
		}
	}
}

func TestCapInteriorIntersects(t *testing.T) {
	tests := []struct {
		c1, c2 Cap
		want   bool
	}{
		{empty, empty, false},
		{empty, xAxis, false},
		{full, empty, false},
		{full, full, true},
		{full, xAxis, true},
		{xAxis, full, false},
		{xAxis, xAxis, false},
		{xAxis, empty, false},
		{concave, hemi.Complement(), true},
	}
	for _, test := range tests {
		if got := test.c1.InteriorIntersects(test.c2); got != test.want {
			t.Errorf("%v.InteriorIntersects(%v); got %t want %t", test.c1, test.c2, got, test.want)
		}
	}
}

func TestCapInteriorContains(t *testing.T) {
	if hemi.InteriorContainsPoint(Point{PointFromCoords(1, 0, -(1 + epsilon)).Normalize()}) {
		t.Errorf("hemi (%v) should not contain point just past half way(%v)", hemi,
			Point{PointFromCoords(1, 0, -(1 + epsilon)).Normalize()})
	}
}

func TestCapExpanded(t *testing.T) {
	cap50 := CapFromCenterAngle(xAxisPt, 50.0*s1.Degree)
	cap51 := CapFromCenterAngle(xAxisPt, 51.0*s1.Degree)

	if !empty.Expanded(s1.Angle(fullHeight)).IsEmpty() {
		t.Error("Expanding empty cap should return an empty cap")
	}
	if !full.Expanded(s1.Angle(fullHeight)).IsFull() {
		t.Error("Expanding a full cap should return an full cap")
	}

	if !cap50.Expanded(0).ApproxEqual(cap50) {
		t.Error("Expanding a cap by 0° should be equal to the original")
	}
	if !cap50.Expanded(1 * s1.Degree).ApproxEqual(cap51) {
		t.Error("Expanding 50° by 1° should equal the 51° cap")
	}

	if cap50.Expanded(129.99 * s1.Degree).IsFull() {
		t.Error("Expanding 50° by 129.99° should not give a full cap")
	}
	if !cap50.Expanded(130.01 * s1.Degree).IsFull() {
		t.Error("Expanding 50° by 130.01° should give a full cap")
	}
}

func TestRadiusToHeight(t *testing.T) {
	tests := []struct {
		got  s1.Angle
		want float64
	}{
		// Above/below boundary checks.
		{s1.Angle(-0.5), emptyHeight},
		{s1.Angle(0), 0},
		{s1.Angle(math.Pi), fullHeight},
		{s1.Angle(2 * math.Pi), fullHeight},
		// Degree tests.
		{-7.0 * s1.Degree, emptyHeight},
		{-0.0 * s1.Degree, 0},
		{0.0 * s1.Degree, 0},
		{12.0 * s1.Degree, 0.02185239926619},
		{30.0 * s1.Degree, 0.13397459621556},
		{45.0 * s1.Degree, 0.29289321881345},
		{90.0 * s1.Degree, 1.0},
		{179.99 * s1.Degree, 1.99999998476912},
		{180.0 * s1.Degree, fullHeight},
		{270.0 * s1.Degree, fullHeight},
		// Radians tests.
		{-1.0 * s1.Radian, emptyHeight},
		{-0.0 * s1.Radian, 0},
		{0.0 * s1.Radian, 0},
		{1.0 * s1.Radian, 0.45969769413186},
		{math.Pi / 2.0 * s1.Radian, 1.0},
		{2.0 * s1.Radian, 1.41614683654714},
		{3.0 * s1.Radian, 1.98999249660044},
		{math.Pi * s1.Radian, fullHeight},
		{4.0 * s1.Radian, fullHeight},
	}
	for _, test := range tests {
		// float64Eq comes from s2latlng_test.go
		if got := radiusToHeight(test.got); !float64Eq(got, test.want) {
			t.Errorf("radiusToHeight(%v) = %v; want %v", test.got, got, test.want)
		}
	}
}
