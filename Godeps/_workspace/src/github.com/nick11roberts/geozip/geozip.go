//Package geozip implements the Geozip algorithm described here: http://geozipcode.blogspot.nl/2015/02/geozip.html
package geozip

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const maxPrecision int = 18

//Encode generates a geozip bucket id for a given latitude and longitude.
//Argument validate is true if a validation is to be performed.
//Precision is a number in the range of [0, 18] such that 0 gives lowest precision (000000000000000000) and 18 gives the most precise bucket id.
func Encode(latitude, longitude float64, validate bool, precision int) int64 {
	if validate && !Valid(latitude, longitude) {
		return 0
	}
	latitudeShifted := decimal.NewFromFloat(latitude).Add(decimal.NewFromFloat(90.0))
	longitudeShifted := decimal.NewFromFloat(longitude).Add(decimal.NewFromFloat(180.0))
	latString := latitudeShifted.String() + ".0"
	lonString := longitudeShifted.String() + ".0"
	latParts := strings.Split(latString, ".")
	lonParts := strings.Split(lonString, ".")
	latString = resizeCharacteristic(latParts[0]) + resizeMantissa(latParts[1])
	lonString = resizeCharacteristic(lonParts[0]) + resizeMantissa(lonParts[1])
	bucketString := zip(latString, lonString)
	bucket, err := strconv.ParseInt(bucketString, 10, 64)
	if err != nil {
		fmt.Errorf("Error parsing zipped string to int64")
	}
	for i := 0; i < maxPrecision-precision; i++ {
		bucket /= 10
	}
	for i := 0; i < maxPrecision-precision; i++ {
		bucket *= 10
	}
	return bucket
}

//Decode is the inverse operation of Encode.
//Decode returns latitude, longitude, and whether or not they are both represented precisely as float64 types.
func Decode(bucket int64) (float64, float64, bool) {
	var latitudeUnshifted, longitudeUnshifted decimal.Decimal
	var latitude, longitude float64
	var err error
	var exact bool
	bucketString := strconv.FormatInt(bucket, 10)
	for len(bucketString) < 18 {
		bucketString = "0" + bucketString
	}

	latString, lonString := unzip(bucketString)
	latString = latString[0:3] + "." + latString[3:]
	lonString = lonString[0:3] + "." + lonString[3:]

	latitudeUnshifted, err = decimal.NewFromString(latString)
	longitudeUnshifted, err = decimal.NewFromString(lonString)
	if err != nil {
		fmt.Errorf("Error creating decimal from string")
	}
	latitudeUnshifted = latitudeUnshifted.Sub(decimal.NewFromFloat(90.0))
	longitudeUnshifted = longitudeUnshifted.Sub(decimal.NewFromFloat(180.0))
	latitude, exact = latitudeUnshifted.Float64()
	longitude, exact = longitudeUnshifted.Float64()
	return latitude, longitude, exact
}

//Valid returns true if latitude is in the range of [-90, 90] and longtitude is in the range of [-180, 180].
func Valid(latitude, longitude float64) bool {
	if latitude < 90.0 && latitude > -90.0 && longitude < 180.0 && longitude > -180.0 {
		return true
	}
	return false

}

func zip(latDigits, lonDigits string) string {
	var bucketDigits string
	for i := 0; i < 9; i++ {
		bucketDigits += string(latDigits[i])
		bucketDigits += string(lonDigits[i])
	}
	return bucketDigits
}

func unzip(bucketDigits string) (string, string) {
	var latDigits, lonDigits string
	for i := 0; i < 18; i += 2 {
		latDigits += string(bucketDigits[i])
		lonDigits += string(bucketDigits[i+1])
	}
	return latDigits, lonDigits
}

func resizeCharacteristic(characteristic string) string {
	for len(characteristic) < 3 {
		characteristic = "0" + characteristic
	}
	return characteristic
}

func resizeMantissa(mantissa string) string {
	for len(mantissa) > 6 {
		mantissa = mantissa[0 : len(mantissa)-1]
	}
	for len(mantissa) < 6 {
		mantissa = mantissa + "0"
	}
	return mantissa
}
