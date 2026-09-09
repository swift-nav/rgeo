#!/usr/bin/env bash
set -euo pipefail

# Build data/EEZ_land_union_v4_202410.gz, the dataset behind rgeo's
# CountriesEEZ10, as a schema-compatible alternative to data/Countries10.gz.
#
# Geometry: MarineRegions EEZ_land_union_v4_202410 (CC-BY 4.0), obtained from
#   https://www.marineregions.org/downloads.php ("EEZ land union v4 (2024-10)").
#   Only the .shp and .dbf are vendored here; mapshaper reads a shapefile
#   without its .shx, and the source is already EPSG:4326 so no .prj is needed.
#   The checksums below pin the exact release this pipeline was written against.
#
# Attribute crosswalk (ISO_A2, CONTINENT, REGION_UN, SUBREGION) and the
# Antarctica polygon (ATA): Natural Earth Countries10, read back out of
# rgeo's own data/Countries10.gz (which is kept in the repo for this reason).
#
# Requires: bash 4+, node/npx, jq 1.6+, gzip, shasum.
# Run it from anywhere; it operates relative to its own location.

cd "$(dirname "${BASH_SOURCE[0]}")"

# Pin mapshaper: simplification output is version-sensitive, so an unpinned
# `npx mapshaper` would silently produce different geometry over time.
MAPSHAPER_VERSION=0.6.102
mapshaper() { npx --yes "mapshaper@${MAPSHAPER_VERSION}" "$@"; }

for cmd in node npx jq gzip shasum; do
  command -v "$cmd" >/dev/null || { echo "missing required command: $cmd" >&2; exit 1; }
done

if ! shasum -a 256 --check --status <<'EOF'
75aefbf949733235a2eda9d3b4d223c0898945005b8ed10f83d44ba3a58f6cdf  EEZ_land_union_v4_202410.shp
13afc71275e244f62d390f3285f7f64adf318ede4058982f14c1f99c68ef2ee9  EEZ_land_union_v4_202410.dbf
EOF
then
  echo "source shapefile does not match the expected checksums" >&2
  exit 1
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# 1. Attribute crosswalk from Natural Earth, keyed by ISO_A3.
#
# Natural Earth writes "-99" rather than an empty string for countries whose
# ISO code it considers disputed -- France and Norway among them. Falling back
# to the _EH ("Europe holdings") columns keeps those countries in the crosswalk;
# without this they drop out entirely and every French and Norwegian polygon
# fails the completeness check in step 4.
# gzip -dc rather than gunzip -kc: -k is redundant with -c and is rejected
# outright by gzip < 1.6.
gzip -dc ../../data/Countries10.gz > "$work/countries10.geojson"

mapshaper "$work/countries10.geojson" \
  -each 'ISO_A3 = (ISO_A3 && ISO_A3 !== "-99") ? ISO_A3
                : (ISO_A3_EH && ISO_A3_EH !== "-99") ? ISO_A3_EH
                : ADM0_A3,
         ISO_A2 = (ISO_A2 && ISO_A2.length === 2) ? ISO_A2 : ISO_A2_EH' \
  -filter 'ISO_A3 && ISO_A3 !== "-99" && ISO_A3.length === 3 && ISO_A2 && ISO_A2.length === 2' \
  -filter-fields ISO_A3,ISO_A2,CONTINENT,REGION_UN,SUBREGION \
  -o format=csv "$work/iso_lookup.csv"

# 2. Antarctica polygon from Natural Earth (MarineRegions omits it).
mapshaper "$work/countries10.geojson" \
  -filter 'ISO_A3 === "ATA"' \
  -filter-fields ISO_A3,ADMIN,FORMAL_EN,ISO_A2,CONTINENT,REGION_UN,SUBREGION \
  -o "$work/antarctica.geojson"

# 3a. Simplify MarineRegions, rewrite attrs, drop null geoms, join crosswalk.
#
# ISO_A3 and ADMIN must describe the *same* entity. MarineRegions models a
# dependency as SOVEREIGN1=Denmark / TERRITORY1=Greenland, so taking the ISO
# from ISO_TER1 while taking the name from SOVEREIGN1 would label Greenland's
# waters "GRL / Denmark". Use the territory's name only when it actually has
# its own ISO code -- TERRITORY1 is also set for plain subdivisions of a
# sovereign (TERRITORY1=Alaska, ISO_TER1=USA), which must stay "United States".
# These names are only a fallback; step 4 prefers Natural Earth's naming.
mapshaper EEZ_land_union_v4_202410.shp \
  -simplify 10% keep-shapes \
  -clean \
  -filter '$.partCount > 0' \
  -each 'ISO_A3 = ISO_TER1 || ISO_SOV1,
         ADMIN = (ISO_TER1 && ISO_TER1 !== ISO_SOV1) ? TERRITORY1 : SOVEREIGN1,
         FORMAL_EN = SOVEREIGN1' \
  -filter 'ISO_A3 !== ""' \
  -join "$work/iso_lookup.csv" keys=ISO_A3,ISO_A3 \
        fields=ISO_A2,CONTINENT,REGION_UN,SUBREGION \
  -filter-fields ISO_A3,ADMIN,FORMAL_EN,ISO_A2,CONTINENT,REGION_UN,SUBREGION,ISO_SOV1 \
  -o "$work/marineregions.geojson"

# 3b. Merge MarineRegions + Antarctica into one layer.
mapshaper "$work/marineregions.geojson" "$work/antarctica.geojson" combine-files \
  -merge-layers force \
  -o precision=0.0001 "$work/merged.geojson"

# 4. Normalize and validate output schema, joining against Countries10 by ISO_A3.
#
# The two sources are authoritative for different things:
#   - Identity (ISO_A3, ISO_A2, CONTINENT, REGION_UN, SUBREGION) comes from
#     MarineRegions and the step 3a join. Natural Earth only fills blanks. It
#     must never overwrite these: NE describes a dependency's waters using the
#     sovereign's row, which is what collapsed Greenland into DNK/Europe.
#   - Names (ADMIN, FORMAL_EN) come from Natural Earth when the ISO code
#     matched, because rgeo has always returned NE's naming ("United States of
#     America", "Republic of Zimbabwe") and MarineRegions carries short forms.
#
# Natural Earth's "-99" sentinel is treated as absent throughout so it cannot
# reach the output; rgeo surfaces it verbatim as CountryCode2.
# Rows missing a required field are dropped, and Antarctica is re-added.
jq -cs '
  # Natural Earth uses "-99" for a disputed/absent code; treat it as missing.
  def nz($v): if ($v // "") == "" or ($v // "") == "-99" then null else $v end;

  # Reduce any property bag to exactly the seven fields rgeo reads.
  def normalize($p):
    {
      ISO_A3:    (nz($p.ISO_A3)    // ""),
      ADMIN:     (nz($p.ADMIN)     // ""),
      FORMAL_EN: (nz($p.FORMAL_EN) // ""),
      ISO_A2:    (nz($p.ISO_A2)    // ""),
      CONTINENT: (nz($p.CONTINENT) // ""),
      REGION_UN: (nz($p.REGION_UN) // ""),
      SUBREGION: (nz($p.SUBREGION) // "")
    };

  # A Natural Earth row carries its real codes in the _EH columns whenever the
  # plain ones are the "-99" sentinel (France, Norway) or a non-ISO string
  # ("CN-TW" for Taiwan). Resolve them the same way the step 1 crosswalk does,
  # otherwise falling back to a sovereign row yields blank codes.
  def ne_normalize($p):
    normalize($p) + {
      ISO_A3: (nz($p.ISO_A3) // nz($p.ISO_A3_EH) // nz($p.ADM0_A3) // ""),
      ISO_A2: (
        [$p.ISO_A2, $p.ISO_A2_EH]
        | map(select((. // "") | length == 2)) | first // ""
      )
    };

  # rgeo needs every field below to render a usable Location.
  def complete:
    .properties.ISO_A3 != ""
    and .properties.ADMIN != ""
    and .properties.ISO_A2 != ""
    and .properties.CONTINENT != ""
    and .properties.REGION_UN != ""
    and .properties.SUBREGION != "";

  .[0] as $eez
  | .[1] as $ne

  # Index Natural Earth once by resolved ISO_A3. Two rows can share a resolved
  # key (France and Clipperton Island both land on FRA), so sort the sovereign
  # own row last -- INDEX keeps the final entry for a duplicated key.
  | (
      $ne.features | map(.properties)
      | map(. + {_k: (nz(.ISO_A3) // nz(.ISO_A3_EH) // nz(.ADM0_A3))})
      | map(select(._k != null))
      | sort_by(.ADMIN == .SOVEREIGNT)
      | INDEX(._k)
    ) as $ne_by_iso

  | [
      $eez.features[]
      | select(.geometry != null)
      | (.properties.ISO_SOV1 // "") as $sov
      | normalize(.properties) as $p
      | ne_normalize($ne_by_iso[$p.ISO_A3] // {}) as $n

      # Natural Earth has no feature of its own for some territories that
      # MarineRegions models separately -- Reunion, Guadeloupe, Bonaire,
      # Christmas I. and friends are all inside the sovereign multipolygon --
      # so the step 3a join finds nothing and every region field is blank.
      # Dropping those polygons would make inhabited land unresolvable, which
      # is worse than master, so fall back to the sovereign row wholesale.
      # The territory keeps its own identity only when NE can actually
      # describe it; otherwise it resolves as its sovereign, exactly as master
      # did. This is deliberately narrower than the bug it replaces: a
      # territory NE *does* model (GRL, PYF, FRO...) never collapses.
      | (if $n.ISO_A3 != "" then $n else ne_normalize($ne_by_iso[$sov] // {}) end) as $src
      | (if $n.ISO_A3 != "" then $p.ISO_A3 else ($src.ISO_A3 // $p.ISO_A3) end) as $iso

      | .properties = {
          ISO_A3:    (if $iso != "" then $iso else $p.ISO_A3 end),
          ISO_A2:    (if $p.ISO_A2    != "" then $p.ISO_A2    else $src.ISO_A2    end),
          CONTINENT: (if $p.CONTINENT != "" then $p.CONTINENT else $src.CONTINENT end),
          REGION_UN: (if $p.REGION_UN != "" then $p.REGION_UN else $src.REGION_UN end),
          SUBREGION: (if $p.SUBREGION != "" then $p.SUBREGION else $src.SUBREGION end),
          ADMIN:     (if $src.ADMIN     != "" then $src.ADMIN     else $p.ADMIN     end),
          FORMAL_EN: (if $src.FORMAL_EN != "" then $src.FORMAL_EN else $p.FORMAL_EN end)
        }
      | select(complete)
    ]
  | (if (map(.properties.ISO_A3) | index("ATA")) == null
     then . + (
       $ne.features
       | map(select(.properties.ISO_A3 == "ATA"))
       | map(.properties = ne_normalize(.properties))
       | map(select(complete))
     )
     else . end) as $features
  | $eez | .features = $features
' "$work/merged.geojson" "$work/countries10.geojson" > "$work/out.geojson"

# 5. Fail loudly rather than shipping a dataset with known-bad attribution.
problems="$(jq -r '
  (.features | length) as $n
  | (.features | map(select(.properties.ISO_A3 == "-99" or .properties.ISO_A2 == "-99")) | length) as $sentinel
  | (.features | map(.properties.ISO_A3) | unique) as $codes
  | [
      (if $n < 300 then "only \($n) features survived" else empty end),
      (if $sentinel > 0 then "\($sentinel) features still carry the -99 sentinel" else empty end),
      (if ($codes | index("ATA")) == null then "Antarctica (ATA) is missing" else empty end),
      (
        # Sovereigns Natural Earth marks "-99", plus dependencies that must not
        # collapse into their sovereign. Note HKG/MAC are deliberately absent:
        # MarineRegions has no record for either, they fall inside CHN.
        ["FRA","NOR","USA","GBR","GRL","FRO","PYF","NCL","FLK","BMU","ATA"] - $codes
        | if length > 0 then "missing expected country codes: \(join(", "))" else empty end
      ),
      (
        # Codes must be well-formed: Natural Earth stores "CN-TW" in ISO_A2 for
        # Taiwan, which is neither a valid alpha-2 nor the "-99" sentinel.
        .features
        | map(select((.properties.ISO_A3 | length) != 3 or (.properties.ISO_A2 | length) != 2))
        | if length > 0
          then "\(length) features have malformed ISO codes, e.g. \(.[0].properties.ADMIN): \(.[0].properties.ISO_A3)/\(.[0].properties.ISO_A2)"
          else empty end
      )
    ]
  | .[]
' "$work/out.geojson")"

if [ -n "$problems" ]; then
  echo "output validation failed:" >&2
  echo "$problems" >&2
  exit 1
fi

# -n omits the timestamp and original filename from the gzip header, so a
# re-run of this script produces a byte-identical artifact.
gzip -9 -n -c "$work/out.geojson" > ../../data/EEZ_land_union_v4_202410.gz

echo "wrote data/EEZ_land_union_v4_202410.gz ($(jq '.features|length' "$work/out.geojson") features)"
