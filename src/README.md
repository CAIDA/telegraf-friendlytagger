This telegraf plugin can be used to add user-friendly "label" tags to
measurements which contain ASNs, ISO-2 country codes or Netacq-Edge
region/county identifiers.

Examples:
   * a `country_code` tag of `AU` will result in an additional `country_label`
     tag with the value "Australia".
   * a `region_code` tag of `4416` will result in an additional `region_label`
     tag with the value "California".
   * a `county_code` tag of `2735` will result in an additional `county_label`
     tag with the value "San Diego".
   * a `asn` tag of `10000` will result in an additional `asn_label` tag with
     the value "NCM, JP".

### Installation

Make sure ${GOROOT} is set -- default is usually `/usr/local/go`

    go get github.com/influxdata/telegraf
    go get github.com/mattn/go-sqlite3

    mkdir -p ${GOROOT}/src/github.com/influxdata/telegraf/plugins/processors/friendlytagger
    cp friendlytagger.go ${GOROOT}/src/github.com/influxdata/telegraf/plugins/processors/friendlytagger/
    sed -i '8i\        _ "github.com/influxdata/telegraf/plugins/processors/friendlytagger"' ${GOROOT}/src/github.com/influxdata/telegraf/plugins/processors/all/all.go

    cd ${GOROOT}/src/github.com/influxdata/telegraf/ && make && go install -ldflags "-w -s" ./cmd/telegraf


### Configuration

To enable this plugin, add the following to the "Processor Plugins" section of
your telegraf configuration file:

    [[processors.friendlytagger]]
        databasename = "/path/to/your/labeldatabase"

Your label database will be an SQLite3 database containing all of the tag to
label mappings that the plugin can apply. See `../helperscripts` for more
information on how to create and populate this database file.

This plugin also supports several optional configuration options (in
addition to the required `databasename` option):

**reloadfrequency**: the frequency at which the plugin should re-read the
label mappings in the database (in seconds). Defaults to 120 (2 minutes).
Longer intervals will mean it will take longer for any database updates to
be applied to data processed by this plugin.

**countrylabeltable**: the name of the table in the database where country
code to label mappings are stored. Defaults to `country_mappings`.

**countylabeltable**: the name of the table in the database where county
number to label mappings are stored. Defaults to `county_mappings`.

**regionlabeltable**: the name of the table in the database where region
number to label mappings are stored. Defaults to `region_mappings`.

**asnlabeltable**: the name of the table in the database where ASN to label
mappings are stored. Defaults to `asn_mappings`.

