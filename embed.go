package rgeo

// embedding files individually here to allow the linker to strip out unused ones
import _ "embed"

//go:embed data/Cities10.gz
var cities10 []byte

func Cities10() []byte {
	return cities10
}

//go:embed data/Countries10.gz
var countries10 []byte

func Countries10() []byte {
	return countries10
}

//go:embed data/EEZ_land_union_v4_202410.gz
var countriesEEZ10 []byte

// CountriesEEZ10 is an alternative to Countries10 built from the MarineRegions
// EEZ land union rather than Natural Earth. Its polygons extend to the edge of
// each country's exclusive economic zone, roughly 200 nautical miles offshore,
// rather than stopping at the coastline, so it resolves coordinates at sea that
// the other datasets report as ErrLocationNotFound. Do not use that error as an
// "is at sea" test with it; use Countries10 or Countries110 for that.
//
// It also attributes some coastal territories differently. Dependencies with
// their own ISO code keep it (Greenland resolves to GRL, not DNK), but
// territories the crosswalk cannot describe resolve as their sovereign state,
// and Hong Kong, Macao, Åland and the British Indian Ocean Territory are absent
// because MarineRegions has no EEZ record for them.
//
// The data is CC-BY 4.0; see data/EEZ_land_union_v4_202410.txt for the required
// attribution. It is rebuilt by ./scripts/eez/simplify.sh.
func CountriesEEZ10() []byte {
	return countriesEEZ10
}

//go:embed data/Countries110.gz
var countries110 []byte

func Countries110() []byte {
	return countries110
}

//go:embed data/Provinces10.gz
var provinces10 []byte

func Provinces10() []byte {
	return provinces10
}
