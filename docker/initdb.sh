#!/bin/bash

## Note, you may need to configure your swift credentials for this script
## to work!

SCRIPTLOC=../helperscripts

touch friendlytag.db
${SCRIPTLOC}/load_mappings.py friendlytag.db ${SCRIPTLOC}/countries.map country_mappings
${SCRIPTLOC}/generate_region_mappings.py > regions.map
${SCRIPTLOC}/load_mappings.py friendlytag.db ./regions.map region_mappings
${SCRIPTLOC}/generate_county_mappings.py > counties.map
${SCRIPTLOC}/load_mappings.py friendlytag.db ./counties.map county_mappings
