/*
Copyright 2020 Sam Smith

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this file except in compliance with the License.  You may obtain a copy of the
License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied.  See the License for the
specific language governing permissions and limitations under the License.
*/

package rgeo

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	"github.com/go-test/deep"
	"github.com/twpayne/go-geom/encoding/geojson"
)

// decodeDataset unpacks an embedded dataset so tests can assert on the raw
// GeoJSON properties rather than only on ReverseGeocode results.
func decodeDataset(t *testing.T, dataset func() []byte) *geojson.FeatureCollection {
	t.Helper()

	zr, err := gzip.NewReader(bytes.NewReader(dataset()))
	if err != nil {
		t.Fatalf("decompressing dataset: %v", err)
	}
	defer zr.Close()

	var fc geojson.FeatureCollection
	if err := json.NewDecoder(zr).Decode(&fc); err != nil {
		t.Fatalf("decoding dataset: %v", err)
	}

	return &fc
}

var testdata = []struct {
	name     string
	in       []float64
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
			Province:     "El Bayadh",
			ProvinceCode: "DZ-32",
		},
	},
	{
		name: "Madagascar",
		in:   []float64{47.523836, -18.905691},
		err:  nil,
		expected: Location{
			Country:      "Madagascar",
			CountryLong:  "Republic of Madagascar",
			CountryCode2: "MG",
			CountryCode3: "MDG",
			Continent:    "Africa",
			Region:       "Africa",
			SubRegion:    "Eastern Africa",
			Province:     "Analamanga",
			ProvinceCode: "MG-T",
			City:         "Antananarivo",
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
			Province:     "Midlands",
			ProvinceCode: "ZW-MI",
		},
	},
	{
		name:     "Ocean",
		in:       []float64{0, 0},
		err:      ErrLocationNotFound,
		expected: Location{},
	},
	{
		name:     "North Pole",
		in:       []float64{-135, 90},
		err:      ErrLocationNotFound,
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
			Province:     "Antarctica",
			ProvinceCode: "AQ-X01~",
		},
	},
	{
		name: "Alaska",
		in:   []float64{-149.901785, 61.199134},
		err:  nil,
		expected: Location{
			Country:      "United States of America",
			CountryLong:  "United States of America",
			CountryCode2: "US",
			CountryCode3: "USA",
			Continent:    "North America",
			Region:       "Americas",
			SubRegion:    "Northern America",
			Province:     "Alaska",
			ProvinceCode: "US-AK",
			City:         "Anchorage",
		},
	},
	{
		name: "UK",
		in:   []float64{0, 51.5045},
		err:  nil,
		expected: Location{
			Country:      "United Kingdom",
			CountryLong:  "United Kingdom of Great Britain and Northern Ireland",
			CountryCode2: "GB",
			CountryCode3: "GBR",
			Continent:    "Europe",
			Region:       "Europe",
			SubRegion:    "Northern Europe",
			Province:     "Tower Hamlets",
			ProvinceCode: "GB-TWH",
			City:         "London",
		},
	},
	{
		name: "Libya",
		in:   []float64{24.98, 25.86},
		err:  nil,
		expected: Location{
			Country:      "Libya",
			CountryLong:  "Libya",
			CountryCode2: "LY",
			CountryCode3: "LBY",
			Continent:    "Africa",
			Region:       "Africa",
			SubRegion:    "Northern Africa",
			Province:     "Al Kufrah",
			ProvinceCode: "LY-KF",
		},
	},
	{
		name: "Egypt",
		in:   []float64{25.005187, 25.855963},
		err:  nil,
		expected: Location{
			Country:      "Egypt",
			CountryLong:  "Arab Republic of Egypt",
			CountryCode2: "EG",
			CountryCode3: "EGY",
			Continent:    "Africa",
			Region:       "Africa",
			SubRegion:    "Northern Africa",
			Province:     "Al Wadi at Jadid",
			ProvinceCode: "EG-WAD",
		},
	},
	{
		name: "US Border",
		in:   []float64{-102.560616, 48.992073},
		err:  nil,
		expected: Location{
			Country:      "United States of America",
			CountryLong:  "United States of America",
			CountryCode2: "US",
			CountryCode3: "USA",
			Continent:    "North America",
			Region:       "Americas",
			SubRegion:    "Northern America",
			Province:     "North Dakota",
			ProvinceCode: "US-ND",
		},
	},
	{
		name: "Canada Border",
		in:   []float64{-102.560616, 49.02},
		err:  nil,
		expected: Location{
			Country:      "Canada",
			CountryLong:  "Canada",
			CountryCode2: "CA",
			CountryCode3: "CAN",
			Continent:    "North America",
			Region:       "Americas",
			SubRegion:    "Northern America",
			Province:     "Saskatchewan",
			ProvinceCode: "CA-SK",
		},
	},
}

func TestReverseGeocode(t *testing.T) {
	testgeo := `{
		"type":"FeatureCollection",
			"features":[
				{"type":"Feature",
				"properties":{"ISO_A3":"TST"},
				"geometry":{"type":"Polygon",
					"coordinates":[[[0,52],[1,52],[1,53],[0,53],[0,52]]]}}
			]
		}`

	var testdata = []struct {
		name     string
		in       []float64
		err      error
		expected Location
	}{
		{
			name:     "in",
			in:       []float64{0.5, 52.5},
			err:      nil,
			expected: Location{CountryCode3: "TST"},
		},
		{
			name:     "out",
			in:       []float64{0, 0},
			err:      ErrLocationNotFound,
			expected: Location{},
		},
	}

	r, err := New(func() []byte { return compressData(t, testgeo) })
	if err != nil {
		t.Error(err)
	}

	for _, test := range testdata {
		test := test

		t.Run(test.name, func(t *testing.T) {
			result, err := r.ReverseGeocode(test.in)
			if err != test.err {
				t.Errorf("expected error: %s\n got: %s\n", test.err, err)
			}
			if diff := deep.Equal(test.expected, result); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestReverseGeocode_Countries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test (countries) for short mode")
	}
	t.Parallel()

	for _, dataset := range []func() []byte{Countries110, Countries10} {
		r, err := New(dataset)
		if err != nil {
			t.Error(err)
		}

		for _, test := range testdata {
			test := test

			test.expected.Province = ""
			test.expected.ProvinceCode = ""
			test.expected.City = ""

			t.Run(test.name, func(t *testing.T) {
				t.Parallel()

				result, err := r.ReverseGeocode(test.in)
				if err != test.err {
					t.Errorf("expected error: %s\n got: %s\n", test.err, err)
				}
				if diff := deep.Equal(test.expected, result); diff != nil {
					t.Error(diff)
				}
			})
		}
	}
}

func TestReverseGeocode_Provinces(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integraion test (provinces) in short mode")
	}
	t.Parallel()

	r, err := New(Provinces10)
	if err != nil {
		t.Error(err)
	}

	for _, test := range testdata {
		test := test

		test.expected.City = ""

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := r.ReverseGeocode(test.in)
			if err != test.err {
				t.Errorf("expected error: %s\n got: %s\n", test.err, err)
			}
			if diff := deep.Equal(test.expected, result); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func TestReverseGeocode_Cities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test (cities) in short mode.")
	}
	t.Parallel()

	r, err := New(Provinces10, Cities10)
	if err != nil {
		t.Error(err)
	}

	for _, test := range testdata {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, err := r.ReverseGeocode(test.in)
			if err != test.err {
				t.Errorf("expected error: %s\n got: %s\n", test.err, err)
			}
			if diff := deep.Equal(test.expected, result); diff != nil {
				t.Error(diff)
			}
		})
	}
}

// CountriesEEZ10 is built from the MarineRegions EEZ land union by
// scripts/eez/simplify.sh, using Natural Earth only to fill attributes the
// EEZ data lacks. Natural Earth writes the sentinel "-99" instead of an ISO
// code for countries whose code it treats as disputed (France and Norway among
// them), and an earlier revision of that script let the sentinel reach the
// shipped dataset, so ReverseGeocode returned CountryCode2 "-99" and an empty
// CountryCode3 for all French and Norwegian waters. Guard the whole dataset
// rather than a sample: no feature may carry the sentinel in either code.
func TestCountriesEEZ10_NoSentinelCodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test (sentinel codes) in short mode.")
	}
	t.Parallel()

	fc := decodeDataset(t, CountriesEEZ10)

	for i, f := range fc.Features {
		for _, key := range []string{"ISO_A3", "ISO_A2"} {
			if got := getPropertyString(f.Properties, key); got == "-99" {
				t.Errorf("feature %d (%s): %s is the sentinel %q, want a real code",
					i, getPropertyString(f.Properties, "ADMIN"), key, got)
			}
		}
	}
}

// Every feature must carry the fields ReverseGeocode reads, otherwise callers
// get a partially-populated Location for points inside that polygon.
func TestCountriesEEZ10_CompleteProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test (property completeness) in short mode.")
	}
	t.Parallel()

	fc := decodeDataset(t, CountriesEEZ10)

	required := []string{"ISO_A3", "ADMIN", "ISO_A2", "CONTINENT", "REGION_UN", "SUBREGION"}

	for i, f := range fc.Features {
		for _, key := range required {
			if getPropertyString(f.Properties, key) == "" {
				t.Errorf("feature %d (%s): %s is empty",
					i, getPropertyString(f.Properties, "ADMIN"), key)
			}
		}
	}
}

// Dependencies must keep their own ISO code and continent rather than
// collapsing into their sovereign. The EEZ source models these as
// SOVEREIGN1=Denmark / TERRITORY1=Greenland, and an earlier revision of
// scripts/eez/simplify.sh let a Natural Earth join overwrite the territory's
// code with the sovereign's, so Greenland reverse-geocoded to DNK/Europe and
// French Polynesia to Europe with no country code at all.
func TestReverseGeocode_TerritoriesNotCollapsed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test (territories) in short mode.")
	}
	t.Parallel()

	r, err := New(CountriesEEZ10)
	if err != nil {
		t.Fatal(err)
	}

	// Note HKG and MAC are deliberately absent: the MarineRegions EEZ land
	// union has no record for either, both fall inside CHN.
	tests := []struct {
		name      string
		in        []float64
		code2     string
		code3     string
		continent string
	}{
		{"Paris", []float64{2.3522, 48.8566}, "FR", "FRA", "Europe"},
		{"Oslo", []float64{10.7522, 59.9139}, "NO", "NOR", "Europe"},
		{"Nuuk, Greenland", []float64{-51.7216, 64.1836}, "GL", "GRL", "North America"},
		{"Torshavn, Faroe Islands", []float64{-6.7717, 62.0079}, "FO", "FRO", "Europe"},
		{"Papeete, French Polynesia", []float64{-149.5665, -17.5516}, "PF", "PYF", "Oceania"},
		{"Noumea, New Caledonia", []float64{166.4572, -22.2758}, "NC", "NCL", "Oceania"},
		{"Stanley, Falkland Islands", []float64{-57.8560, -51.6977}, "FK", "FLK", "South America"},
		{"Hamilton, Bermuda", []float64{-64.7810, 32.2949}, "BM", "BMU", "North America"},
		{"Douglas, Isle of Man", []float64{-4.4814, 54.1509}, "IM", "IMN", "Europe"},
		{"Oranjestad, Aruba", []float64{-70.0270, 12.5240}, "AW", "ABW", "North America"},
		{"Berlin", []float64{13.4050, 52.5200}, "DE", "DEU", "Europe"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			loc, err := r.ReverseGeocode(test.in)
			if err != nil {
				t.Fatalf("ReverseGeocode(%v): %v", test.in, err)
			}
			if loc.CountryCode2 != test.code2 {
				t.Errorf("CountryCode2 = %q, want %q", loc.CountryCode2, test.code2)
			}
			if loc.CountryCode3 != test.code3 {
				t.Errorf("CountryCode3 = %q, want %q", loc.CountryCode3, test.code3)
			}
			if loc.Continent != test.continent {
				t.Errorf("Continent = %q, want %q", loc.Continent, test.continent)
			}
		})
	}
}

// Natural Earth has no feature of its own for territories it folds into the
// sovereign's multipolygon (Réunion, Guadeloupe, Bonaire, Christmas Island and
// others), so the crosswalk in scripts/eez/simplify.sh cannot describe them and
// they resolve as their sovereign. That is deliberate, but they must still
// resolve: an earlier revision dropped them outright and made inhabited land
// return ErrLocationNotFound, which is worse than the dataset they replaced.
func TestReverseGeocode_SovereignFallbackTerritories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test (sovereign fallback) in short mode.")
	}
	t.Parallel()

	r, err := New(CountriesEEZ10)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		in    []float64
		code3 string
	}{
		{"Cayenne, French Guiana", []float64{-52.3260, 4.9224}, "FRA"},
		{"Saint-Denis, Reunion", []float64{55.4504, -20.8823}, "FRA"},
		{"Fort-de-France, Martinique", []float64{-61.0742, 14.6161}, "FRA"},
		{"Pointe-a-Pitre, Guadeloupe", []float64{-61.5314, 16.2415}, "FRA"},
		{"Mamoudzou, Mayotte", []float64{45.2270, -12.7806}, "FRA"},
		{"Longyearbyen, Svalbard", []float64{15.6469, 78.2232}, "NOR"},
		{"Kralendijk, Bonaire", []float64{-68.2800, 12.1500}, "NLD"},
		{"Christmas Island", []float64{105.6904, -10.4475}, "AUS"},
		{"Cocos Islands", []float64{96.8710, -12.1642}, "AUS"},
		{"Tokelau", []float64{-171.8484, -9.2002}, "NZL"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			loc, err := r.ReverseGeocode(test.in)
			if err != nil {
				t.Fatalf("ReverseGeocode(%v): %v", test.in, err)
			}
			if loc.CountryCode3 != test.code3 {
				t.Errorf("CountryCode3 = %q, want %q", loc.CountryCode3, test.code3)
			}
		})
	}
}

// Natural Earth stores "CN-TW" in ISO_A2 for Taiwan, which is neither a valid
// alpha-2 nor the "-99" sentinel, so a sentinel-only check lets it through.
// Assert the shape of every code rather than just the absence of "-99".
func TestCountriesEEZ10_WellFormedCodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test (code shape) in short mode.")
	}
	t.Parallel()

	fc := decodeDataset(t, CountriesEEZ10)

	for i, f := range fc.Features {
		admin := getPropertyString(f.Properties, "ADMIN")
		if got := getPropertyString(f.Properties, "ISO_A3"); len(got) != 3 {
			t.Errorf("feature %d (%s): ISO_A3 = %q, want 3 characters", i, admin, got)
		}
		if got := getPropertyString(f.Properties, "ISO_A2"); len(got) != 2 {
			t.Errorf("feature %d (%s): ISO_A2 = %q, want 2 characters", i, admin, got)
		}
	}
}

func TestNew_BadData(t *testing.T) {
	testdata := []struct {
		name string
		in   func() []byte
		err  string
	}{
		{
			name: "Empty data",
			in:   func() []byte { return []byte(``) },
			err:  "no data in dataset 0",
		},
		{
			name: "Wrong type",
			in: func() []byte {
				return compressData(t,
					`{"type":"FeatureCollection","features":
							[{"type":"Feature","geometry":
								{"type":"Point","coordinates":[0,0]}}]}`,
				)
			},
			err: "bad polygon in geometry: needs Polygon or MultiPolygon",
		},
		{
			name: "Small polygon",
			in: func() []byte {
				return compressData(t,
					`{"type":"FeatureCollection","features":
							[{"type":"Feature","geometry":
								{"type":"Polygon",
								"coordinates":[[[1,2],[3,4],[1,2]]]}}]}`,
				)
			},
			err: "bad polygon in geometry: can't convert ring with less than 4 points",
		},
		{
			name: "No repeated end",
			in: func() []byte {
				return compressData(t,
					`{"type":"FeatureCollection","features":
							[{"type":"Feature","geometry":
								{"type":"Polygon",
								"coordinates":[[[1,2],[3,4],[5,6],[7,8]]]}}]}`,
				)
			},
			err: "bad polygon in geometry: " +
				"last coordinate not same as first for polygon: [1 2 3 4 5 6 7 8]",
		},
		{
			name: "Bad Multipolygon",
			in: func() []byte {
				return compressData(t,
					`{"type":"FeatureCollection","features":
							[{"type":"Feature","geometry":
								{"type":"MultiPolygon",
								"coordinates":[[[[1,2],[3,4],[5,6],[7,8]]]]}}]}`,
				)
			},
			err: "bad polygon in geometry: " +
				"last coordinate not same as first for polygon: [1 2 3 4 5 6 7 8]",
		},
		{
			name: "Bad compression",
			in:   func() []byte { return []byte(`dGhpcyBpcyBub3QgU29tcHJIc3NIZA==`) },
			err:  "decompression failed for dataset 0: gzip: invalid header",
		},
		{
			name: "Bad JSON",
			in:   func() []byte { return []byte(compressData(t, `this is not JSON`)) },
			err:  "invalid JSON in dataset 0: invalid character 'h' in literal true (expecting 'r')",
		},
	}
	for _, test := range testdata {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.in)
			if err != nil && err.Error() != test.err {
				t.Errorf("expected error: %s\n got: %s\n", test.err, err)
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
		{
			name: "London",
			in: Location{
				Country:      "United Kingdom",
				CountryLong:  "United Kingdom of Great Britain and Northern Ireland",
				CountryCode3: "GBR",
				Continent:    "Europe",
				City:         "London",
			},
			expected: "<Location> London, United Kingdom (GBR), Europe",
		},
		{
			name:     "Empty",
			in:       Location{},
			expected: "<Location> Empty Location",
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

func ExampleRgeo_ReverseGeocode() {
	r, err := New(Countries110)
	if err != nil {
		// Handle error
	}

	loc, err := r.ReverseGeocode([]float64{0, 52})
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

func ExampleRgeo_ReverseGeocode_city() {
	r, err := New(Provinces10, Cities10)
	if err != nil {
		// Handle error
	}

	loc, err := r.ReverseGeocode([]float64{141.35, 43.07})
	if err != nil {
		// Handle error
	}

	fmt.Println(loc)
	// Output: <Location> Sapporo, Hokkaidō, Japan (JPN), Asia
}

func BenchmarkReverseGeocode_110(b *testing.B) {
	r, err := New(Countries110)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}

func BenchmarkReverseGeocode_10(b *testing.B) {
	r, err := New(Countries10)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}

func BenchmarkReverseGeocode_Prov10(b *testing.B) {
	r, err := New(Provinces10)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}

func BenchmarkReverseGeocode_City10(b *testing.B) {
	r, err := New(Provinces10, Cities10)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = r.ReverseGeocode([]float64{
			(rand.Float64() * 360) - 180,
			(rand.Float64() * 180) - 90,
		})
	}
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := New(Countries110)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkNewCity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := New(Cities10)
		if err != nil {
			b.Error(err)
		}
	}
}

func compressData(t *testing.T, in string) []byte {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	if _, err := zw.Write([]byte(in)); err != nil {
		t.Error(err)
	}

	if err := zw.Close(); err != nil {
		t.Error(err)
	}

	return buf.Bytes()
}
