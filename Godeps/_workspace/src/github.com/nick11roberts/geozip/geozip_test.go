package geozip

import "testing"

var encodeTests = []struct {
	latitude  float64
	longitude float64
	validate  bool
	precision int
	bucket    int64
}{
	{-34.783467, 128.294109, true, 18, 35058221964513039},
	{-34.783467, 128.2941, true, 18, 35058221964513030},
	{0.0, 0.0, true, 18, 19800000000000000},
	{0.0, 0.0, true, 0, 00000000000000000},
	{-34.783467, 128.294109, true, 17, 35058221964513030},
	{-34.783467, 128.294109, true, 16, 35058221964513000},
	{-34.783467, 128.294109, true, 15, 35058221964513000},
	{-34.783467, 128.294109, true, 14, 35058221964510000},
	{-304.783467, 528.294109, true, 14, 0},
}

var decodeTests = []struct {
	latitude  float64
	longitude float64
	exact     bool
	bucket    int64
}{
	{-34.783467, 128.294109, true, 35058221964513039},
	{-34.783467, 128.2941, true, 35058221964513030},
}

func TestEncode(t *testing.T) {
	for _, tableEntry := range encodeTests {
		bucket := Encode(tableEntry.latitude, tableEntry.longitude, tableEntry.validate, tableEntry.precision)
		if bucket != tableEntry.bucket {
			t.Errorf("Encode(%v, %v, %v, %v) = %v, want %v", tableEntry.latitude, tableEntry.longitude, tableEntry.validate, tableEntry.precision, bucket, tableEntry.bucket)
		}
	}
}

func TestDecode(t *testing.T) {
	for _, tableEntry := range decodeTests {
		lat, lon, ex := Decode(tableEntry.bucket)
		if (lat != tableEntry.latitude || lon != tableEntry.longitude) && ex {
			t.Errorf("Decode(%v) = %v, %v, want %v, %v", tableEntry.bucket, lat, lon, tableEntry.latitude, tableEntry.longitude)
		}
	}
}

func BenchmarkGeozip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(-34.783467, 128.294109, true, 18)
	}
}
