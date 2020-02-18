package rgeo

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/go-test/deep"
	geom "github.com/twpayne/go-geom"
)

func TestReverseGeocode(t *testing.T) {
	tests := []struct {
		name     string
		in       geom.Coord
		err      error
		expected Location
	}{
		{
			name: "Algeria",
			in:   []float64{1.880273, 31.787305},
			err:  nil,
			expected: Location{
				Country:      "Algeria",
				CountryLong:  "People's Democratic Republic of Algeria",
				CountryCode2: "DZ",
				CountryCode3: "DZA",
				Continent:    "Africa",
				Region:       "Africa",
				SubRegion:    "Northern Africa",
			},
		},
		{
			name: "Madagascar",
			in:   []float64{47.478275, -17.530126},
			err:  nil,
			expected: Location{
				Country:      "Madagascar",
				CountryLong:  "Republic of Madagascar",
				CountryCode2: "MG",
				CountryCode3: "MDG",
				Continent:    "Africa",
				Region:       "Africa",
				SubRegion:    "Eastern Africa",
			},
		},
		{
			name: "Zimbabwe",
			in:   []float64{29.832875, -19.948725},
			err:  nil,
			expected: Location{
				Country:      "Zimbabwe",
				CountryLong:  "Republic of Zimbabwe",
				CountryCode2: "ZW",
				CountryCode3: "ZWE",
				Continent:    "Africa",
				Region:       "Africa",
				SubRegion:    "Eastern Africa",
			},
		},
		{
			name:     "Ocean",
			in:       []float64{0, 0},
			err:      errCountryNotFound,
			expected: Location{},
		},
		{
			name:     "North Pole",
			in:       []float64{-135, 90},
			err:      errCountryNotFound,
			expected: Location{},
		},
		{
			name: "South Pole",
			in:   []float64{44.99, -89.99},
			err:  nil,
			expected: Location{
				Country:      "Antarctica",
				CountryLong:  "",
				CountryCode2: "AQ",
				CountryCode3: "ATA",
				Continent:    "Antarctica",
				Region:       "Antarctica",
				SubRegion:    "Antarctica",
			},
		},
		{
			name: "Alaska",
			in:   []float64{-150.542, 66.3},
			err:  nil,
			expected: Location{
				Country:      "United States of America",
				CountryLong:  "United States of America",
				CountryCode2: "US",
				CountryCode3: "USA",
				Continent:    "North America",
				Region:       "Americas",
				SubRegion:    "Northern America",
			},
		},
		{
			name: "UK",
			in:   []float64{0, 52},
			err:  nil,
			expected: Location{
				Country:      "United Kingdom",
				CountryLong:  "United Kingdom of Great Britain and Northern Ireland",
				CountryCode2: "GB",
				CountryCode3: "GBR",
				Continent:    "Europe",
				Region:       "Europe",
				SubRegion:    "Northern Europe",
			},
		},
	}

	/*
		rgeo, err := New()
		if err != nil {
			t.Error(err)
		}
	*/

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result, err := ReverseGeocode(test.in)
			if err != test.err {
				t.Errorf("expected error: %s\n got: %s\n", test.err, err)
			}
			if diff := deep.Equal(test.expected, result); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		in       Location
		expected string
	}{
		{
			name: "Algeria",
			in: Location{
				Country:      "Algeria",
				CountryCode3: "DZA",
				Continent:    "Africa",
			},
			expected: "<Location> Algeria (DZA), Africa",
		},
		{
			name: "Zimbabwe",
			in: Location{
				CountryLong:  "Republic of Zimbabwe",
				CountryCode2: "ZW",
				Region:       "Africa",
			},
			expected: "<Location> Republic of Zimbabwe (ZW), Africa",
		},
		{
			name: "Northern America",
			in: Location{
				SubRegion: "Northern America",
			},
			expected: "<Location> Northern America",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			result := test.in.String()
			if diff := deep.Equal(test.expected, result); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func ExampleReverseGeocode() {
	loc, err := ReverseGeocode([]float64{0, 52})
	if err != nil {
		// Handle error
	}

	fmt.Printf("%s\n", loc.Country)
	fmt.Printf("%s\n", loc.CountryLong)
	fmt.Printf("%s\n", loc.CountryCode2)
	fmt.Printf("%s\n", loc.CountryCode3)
	fmt.Printf("%s\n", loc.Continent)
	fmt.Printf("%s\n", loc.Region)
	fmt.Printf("%s\n", loc.SubRegion)

	// Output: United Kingdom
	// United Kingdom of Great Britain and Northern Ireland
	// GB
	// GBR
	// Europe
	// Europe
	// Northern Europe
}

func ExampleReverseGeocode_with_rgeo() {
	r, err := New()
	if err != nil {
		// Handle error
	}

	for i := -33; i <= 31; i += 5 {
		loc, err := r.ReverseGeocode([]float64{24, float64(i)})
		if err != nil {
			// Handle error
		}

		fmt.Printf("%s, ", loc.CountryCode2)
	}

	fmt.Printf("\n")

	// Output: ZA, ZA, BW, NA, ZM, CD, CD, CD, CF, SD, SD, LY, LY,
}

func BenchmarkReverseGeocode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}

func BenchmarkReverseGeocode_with_rgeo(b *testing.B) {
	r, _ := New()
	for i := 0; i < b.N; i++ {
		_, _ = r.ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}
