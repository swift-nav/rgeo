#!/usr/bin/env bash
set -euo pipefail

# Build EEZ_land_union_v4_202410.gz as a drop-in replacement for
# rgeo's data/Countries10.gz.
#
# Geometry: MarineRegions EEZ_land_union_v4_202410 (CC-BY 4.0).
# Attribute crosswalk (ISO_A2, CONTINENT, REGION_UN, SUBREGION) and
# the Antarctica polygon (ATA): Natural Earth Countries10.

# 1. Attribute crosswalk from Natural Earth, keyed by ISO_A3.
gunzip -kc ../../data/Countries10.gz > countries10.geojson

npx mapshaper countries10.geojson \
  -filter 'ISO_A3 && ISO_A3 !== "-99"' \
  -filter-fields ISO_A3,ISO_A2,CONTINENT,REGION_UN,SUBREGION \
  -o format=csv iso_lookup.csv

# 2. Antarctica polygon from Natural Earth (MarineRegions omits it).
npx mapshaper countries10.geojson \
  -filter 'ISO_A3 === "ATA"' \
  -filter-fields ISO_A3,ADMIN,FORMAL_EN,ISO_A2,CONTINENT,REGION_UN,SUBREGION \
  -o antarctica.geojson

# 3a. Simplify MarineRegions, rewrite attrs, drop null geoms, join crosswalk.
npx mapshaper EEZ_land_union_v4_202410.shp \
  -simplify 10% keep-shapes \
  -clean \
  -filter '$.partCount > 0' \
  -each 'ISO_A3 = ISO_TER1 || ISO_SOV1,
         ADMIN = SOVEREIGN1 || TERRITORY1,
         FORMAL_EN = SOVEREIGN1' \
  -filter 'ISO_A3 !== ""' \
  -join iso_lookup.csv keys=ISO_A3,ISO_A3 \
        fields=ISO_A2,CONTINENT,REGION_UN,SUBREGION \
  -filter-fields ISO_A3,ADMIN,FORMAL_EN,ISO_A2,CONTINENT,REGION_UN,SUBREGION \
  -o marineregions.geojson

# 3b. Merge MarineRegions + Antarctica into one layer.
npx mapshaper marineregions.geojson antarctica.geojson combine-files \
  -merge-layers force \
  -o precision=0.0001 EEZ_land_union_v4_202410.geojson

# 4. Normalize and validate output schema.
# - Remove null-geometry features.
# - Backfill properties from Countries10 by ISO_A3, with ADMIN fallback.
# - Keep only rows with complete required fields.
# - Ensure Antarctica is present.
jq -cs '
  .[0] as $eez
  | .[1] as $ne
  | def ne_props($f):
      (
        $ne.features
        | map(
            select(
              ((.properties.ISO_A3 // "") != "" and .properties.ISO_A3 == ($f.properties.ISO_A3 // ""))
              or (.properties.ADMIN // "") == ($f.properties.ADMIN // "")
            )
          )
        | .[0].properties // {}
      );
  $eez
  | .features = (
      [
        .features[]
        | select(.geometry != null)
        | . as $f
        | .properties = (.properties + ne_props($f))
        | .properties = {
            ISO_A3: (.properties.ISO_A3 // ""),
            ADMIN: (.properties.ADMIN // ""),
            FORMAL_EN: (.properties.FORMAL_EN // ""),
            ISO_A2: (.properties.ISO_A2 // ""),
            CONTINENT: (.properties.CONTINENT // ""),
            REGION_UN: (.properties.REGION_UN // ""),
            SUBREGION: (.properties.SUBREGION // "")
          }
        | select(
            .properties.ISO_A3 != ""
            and .properties.ADMIN != ""
            and .properties.ISO_A2 != ""
            and .properties.CONTINENT != ""
            and .properties.REGION_UN != ""
            and .properties.SUBREGION != ""
          )
      ]
      | if (map(.properties.ISO_A3) | index("ATA")) == null
        then . + (
          $ne.features
          | map(select(.properties.ISO_A3 == "ATA"))
          | map(
              .properties = {
                ISO_A3: (.properties.ISO_A3 // ""),
                ADMIN: (.properties.ADMIN // ""),
                FORMAL_EN: (.properties.FORMAL_EN // ""),
                ISO_A2: (.properties.ISO_A2 // ""),
                CONTINENT: (.properties.CONTINENT // ""),
                REGION_UN: (.properties.REGION_UN // ""),
                SUBREGION: (.properties.SUBREGION // "")
              }
            | select(
                .properties.ISO_A3 != ""
                and .properties.ADMIN != ""
                and .properties.ISO_A2 != ""
                and .properties.CONTINENT != ""
                and .properties.REGION_UN != ""
                and .properties.SUBREGION != ""
              )
            )
        )
        else .
        end
    )
' EEZ_land_union_v4_202410.geojson countries10.geojson > EEZ_land_union_v4_202410.normalized.geojson
mv EEZ_land_union_v4_202410.normalized.geojson EEZ_land_union_v4_202410.geojson

gzip -9 -f EEZ_land_union_v4_202410.geojson
mv EEZ_land_union_v4_202410.geojson.gz ../../data/EEZ_land_union_v4_202410.gz
