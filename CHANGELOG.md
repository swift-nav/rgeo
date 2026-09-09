# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
 - New `CountriesEEZ10` dataset, built from the MarineRegions EEZ land union
 (v4, 2024-10) rather than Natural Earth. Its polygons extend to the edge of
 each country's exclusive economic zone, roughly 200 nautical miles offshore,
 instead of stopping at the coastline, so coordinates at sea resolve to a
 country rather than returning `ErrLocationNotFound`. Do not use that error as
 an "is at sea" test with it.

   This is additive: `Countries10` is unchanged and still uses Natural Earth.
   Datasets are embedded individually so the linker strips unused ones, meaning
   `CountriesEEZ10` costs nothing unless you call it.

   Compared with `Countries10`, it fixes several coastal attribution problems:

   - France and Norway return `FR`/`FRA` and `NO`/`NOR` rather than the Natural
   Earth `-99` sentinel with an empty `CountryCode3`, and Taiwan returns `TW`
   rather than `CN-TW`.
   - Réunion, Guadeloupe, Martinique, French Guiana, Mayotte, Svalbard,
   Bonaire, Christmas Island, the Cocos Islands and Tokelau resolve as their
   sovereign state; several previously returned `ErrLocationNotFound`.

   It also drops some codes `Countries10` has, because MarineRegions has no EEZ
   record for them: Hong Kong (`HKG`), Macao (`MAC`) and Åland (`ALA`) resolve
   to `CHN`, `CHN` and `FIN`, and the Chagos Archipelago is attributed to
   Mauritius, so `IOT` resolves to `MUS`.

 - `scripts/eez/simplify.sh`, which rebuilds `CountriesEEZ10` reproducibly:
 checksummed source data, a pinned mapshaper version, and a validation step
 that fails the build on malformed or missing country codes.

## [1.3.0] - 2025-03-08

### Added
 - New `Build` method, see below

### Changed
 - Make `ReverseGeocode` thread safe, thanks to @mologie (#34)
	- This makes the first call to `ReverseGeocode` much slower. This can be
	mitigated by calling `Build` first.
 -  Optimizations: Only link used datasets, avoid copies, avoid pkg/errors
 dependency, thanks to @mologie (#26)
 - Update natural earth data to v5.1.2, thanks to @SaTae66
 - Bump dependency versions, thanks to @dependabot

## [1.2.0] - 2023-01-03

It's been a while since the last release, so all of the dependencies have been
updated, but this is mostly about the first user contribution.

### Changed
 - Moved to using Go embed for the data files, thanks to @benjojo (#18)
 - Updated to Go 1.19
 - Updated other dependencies

## [1.1.1] - 2020-03-06

I broke v1.1.0 when removing the large files from the git history so this will
replace it.

### Changed
 - Much smaller repo
 - Readme rewrite

## [1.1.0] - 2020-02-28

Mostly just bumping the version to see if pkg.go.dev will update now the
data files are much smaller.

### Changed
 - GeoJSON data is now compressed with gzip in go files, leading to much smaller
   file sizes without a noticeable performance hit when reading.

## [1.0.0] - 2020-02-27

This release is now broken due to some large files being removed from the git
history.

### Added
 - Province information to `Location`.
 - New dataset `Cities10` and added `City` field to `Location`.

### Changed
 - Fixed datagen, and moved template back into code.
 - `New` can now accept multiple datasets (which is only really useful when you
   want to use `Cities10` and get province/country information.
 - Switched to Apache License.
 - Better errors.
 - 100% test coverage.

## [0.0.5] - 2020-02-23

This release is now broken due to some large files being removed from the git
history.

### Added
 - JSON strings in `Location` for more idiomatic marshalling .
 - datagen can now read multiple inputs to one output.
 - 1:10m scale datasets for countries and provinces.
 - Datasets are returned from functions so the compiler doesn't include unused
   ones in builds.
 - Brought back the `New` function to initialise the data.

### Changes
 - Massive speed increase on queries.
 - Data initialisation is probably slower.

## [0.0.4] - 2020-02-20

### Changed
 - Minor Documentation changes, making a new release just to get rid of the
   giant `This file is generated` in the godoc.

## [0.0.3] - 2020-02-19

### Added
 - More robust Location printing.

### Changed
 - Data is now included as a struct in a go file instead of geojson.
 - Changed the algorithm to remove dependence on s2, so it doesn't need to
   convert between s2 and geom types. This is a lot faster than converting to s2
   types every time, but slower than pre-converted.

### Removed
 - Function New() and access to the rgeo data type, the data is already parsed
   so it doesn't need to be parsed when used.

## [0.0.2] - 2020-02-17

### Added

 - 2 letter country codes, continents, regions and subregions to output.
 - Type `Rgeo` and function `New` to parse the JSON and create the polygons
   ahead of time so it doesn't need to be done every time `ReverseGeocode` is
   run.

### Changed

 - Moved to using s2 Polygons instead of just s2 Loops.
 - Using github.com/go-test/deep for nicer printing in tests.

## [0.0.1] - 2020-02-15

### Added

Initial release
 - Exposes Function `ReverseGeocode` and type `Location`.
 - Just give `ReverseGeocode` a pair of coordinates and it will return a
   `Location` containing information about which country those coordinates are
   in.
