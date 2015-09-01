package main

import (
	"fmt"

	"github.com/nick11roberts/geozip"
)

func main() {
	lat, lon := 45.321, 164.4533
	bucket := geozip.Encode(lat, lon, true, 18)
	fmt.Println(bucket)
}
