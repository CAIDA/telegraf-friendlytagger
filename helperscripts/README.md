## Quick-start guide

    # remember to configure your swift credentials first!

    touch friendlytag.db
    ./load_mappings.py friendlytag.db ./countries.map country_mappings
    ./generate_region_mappings.py > regions.map
    ./load_mappings.py friendlytag.db ./regions.map region_mappings
    ./generate_county_mappings.py > counties.map
    ./load_mappings.py friendlytag.db ./counties.map county_mappings
    go run queryasnnames.go friendlytag.db


## Some basic notes on the helper scripts

### load_mappings.py

Given a ".map" file (either the existing `countries.map` or the output
generated by one of the `generate_X_mappings` scripts, this script
will insert all of the tag to label mappings into an SQLite3 database
where they can be accessed by the telegraf-friendlytagger module.

To run:
    ./load_mappings.py <database file> <.map file> <table name>

This will insert all mappings found in the `<.map file>` into the
table "`<table name>`" in the SQLite database stored in the file
`<database file>`.


### .map file format
Map files use a simple CSV format, with one line per label. Each line should
contain 4 comma-separated fields: the label, the code that matches the label,
the timestamp that the mapping applies from, and the timestamp that the
mapping ceases to apply (note this latter field is not used in practice).

See `countries.map` for an example.

### generate_region_mappings.py

This script will query the Natural Earth polygon list to derive a set of
polygon IDs representing regions in the Netacq geo-location data and the
corresponding name for each of those polygons. These are then written
to standard output as CSV conforming to the .map file format described
above.

This script uses a swift lookup to read the file with this information, so
you will need to configured appropriate swift credentials in your
environment for the script to work.

If the output is redirected to a .map file, that file can then be used with
`load_mapping.py` to insert those region names into the database that will be
used by the telegraf plugin.

### generate_county_mappings.py

This script will query the GADM polygon list to derive a set of
polygon IDs representing US counties in the Netacq geo-location data and the
corresponding name for each of those polygons. These are then written
to standard output as CSV conforming to the .map file format described
above.

This script uses a swift lookup to read the file with this information, so
you will need to configured appropriate swift credentials in your
environment for the script to work.

### queryasnnames.go

This go program uses the ASRank REST API to discover all of the ASN to
AS Name mappings and inserts any new entries into the SQLite3 database.
The mappings are automatically inserted into a table called `asn_mappings`,
although this could be made configurable in the future if need be.

You can run this program as follows (assuming go is installed):

    go run queryasnnames.go <database file>

The "Created" timestamp associated with each ASN returned by the REST API
is used as the "apply from" timestamp for the label mapping. You will
want to re-run this program on a regular basis to ensure any name changes
or ASN additions are incorporated into your database sooner rather than later.

ASNs that have no usable name in the ASRank dataset will be assigned the label
"Name Unknown" followed by the corresponding Country field.
