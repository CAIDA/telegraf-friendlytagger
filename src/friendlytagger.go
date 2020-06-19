/*
Package friendlytagger implements a telegraf processing plugin that
adds human-readable labels to certain tags present in time series input.

Code tags that can be augmented with labels by this plugin:
  continent_code
  country_code
  region_code
  county_code
  asn

The code -> label mappings for all code tags (except the continent
codes which are statically defined in this module) are read from an
SQLite3 database. The database should have a separate table for each
code tag type, e.g. a table for country codes, a table for ASNs etc.

The mapping tables should have the following columns:
   code (text)
   label (text)
   apply_from (int)

"code" is the tag value that will already be present in the report
plugin output for the metric concerned. "label" is the human-readable
label that corresponds to the code. For instance, a country code of
"US" would have a label of "United States of America".

"apply_from" is a unix timestamp that indicates when a given label
should be considered as applying to the code. This allows for a code
to have different labels at different time periods, such as when a
region changes names or an ASN changes ownership (and therefore name).
When assigning a label to a code, this plugin will use the label with
the closest "apply_from" timestamp to the timestamp associated with
the telegraf metric.

If the underlying database is updated, you will need to restart
telegraf to load any new mappings into the plugin.

See the accompanying README for details on how to install this plugin.
*/
package friendlytagger

import (
    "fmt"
    "log"
    "database/sql"
    "github.com/influxdata/telegraf"
    "github.com/influxdata/telegraf/plugins/processors"
    _ "github.com/mattn/go-sqlite3"
)

var sampleConfig = `
  [[processors.friendlytagger]]
    databasename = "mysqlite.db"
    reloadfrequency = 3600

    countrylabeltable = "country_mappings"
    regionlabeltable = "region_mappings"
    countylabeltable = "county_mappings"
    asnlabeltable = "asn_mappings"

`

// continents are unlikely to change, so we can just define those here
var staticContinents = map[string]LabelSet {
    "??": {
        StartTimes: []int64{0},
        Labels: []string {"Unassigned"},
    },
    "AS": {
        StartTimes: []int64{0},
        Labels: []string {"Asia"},
    },
    "NA": {
        StartTimes: []int64{0},
        Labels: []string {"North America"},
    },
    "EU": {
        StartTimes: []int64{0},
        Labels: []string {"Europe"},
    },
    "OC": {
        StartTimes: []int64{0},
        Labels: []string {"Oceania"},
    },
    "SA": {
        StartTimes: []int64{0},
        Labels: []string {"South America"},
    },
    "AF": {
        StartTimes: []int64{0},
        Labels: []string {"Africa"},
    },
    "AN": {
        StartTimes: []int64{0},
        Labels: []string {"Antarctica"},
    },
}

// An instance of this plugin
type FriendlyTagger struct {
    // The path to the SQLite3 database where the code->label mappings are
    DatabaseName string

    // The name of the table where the country code mappings are
    CountryLabelTable string

    // The name of the table where the region code mappings are
    RegionLabelTable string

    // The name of the table where the county code mappings are
    CountyLabelTable string

    // The name of the table where the ASN name mappings are
    AsnLabelTable string

    // The set of code->label mappings, indexed by tag type (e.g.
    // country_code, asn, region_code, etc.)
    Replacements map[string]FriendlyTag

    LastReload int64

    ReloadFrequency int64
}

// A set of labels for a given code and the timestamps from which those
// labels apply. Each entry in the Labels array will have its corresponding
// timestamp at the same index in the StartTimes array.
//
// Example: ASN 681 changes its name from "University of Waikato, NZ" to
// "Quigley College, NZ" at timestamp 1592346088
//
// StartTimes: {0, 1592346088}
// Labels: {"University of Waikato, NZ", "Quigley College, NZ"}
type LabelSet struct {
    StartTimes  []int64
    Labels      []string
}

// A set of code->label mappings for a single tag type (e.g. country_code)
type FriendlyTag struct {
    // The name of the code field in the report plugin output
    CodedTag        string
    // The name of the label field to add to the report plugin output
    NewTag          string
    // The code->label mappings for this tag type
    ValueMappings   map[string]LabelSet
}

// SampleConfig returns some sample configuration for this plugin, required by
// the telegraf processor plugin interface.
func (tagger *FriendlyTagger) SampleConfig() string {
    return sampleConfig
}

// Description returns a brief description of this plugin, required by the
// telegraf processor plugin interface.
func (tagger *FriendlyTagger) Description() string {
    return "Add additional human-readable labels for certain tags within time series"
}


/*
LoadGenericLabels reads the code->label mappings for a given tag type from the
SQLite database and inserts them into memory for fast lookup when processing.
"table" is the name of the table to read the mappings from.
"replacecode" is the tag name for the field where the code will be found in
the report plugin output (e.g. 'country_code', 'asn').
"replacelabel" is the tag name for the field where the label will be added if
the code matches one of the codes found in the table (e.g. "country_label").
*/
func (tagger *FriendlyTagger) LoadGenericLabels(table string,
        replacecode string, replacelabel string) {

    db, err := sql.Open("sqlite3", tagger.DatabaseName)
    if err != nil {
        log.Fatal(err)
    }

    // ensure multiple labels for the same code are in timestamp order,
    // otherwise the reverse iteration that we do later on to find the
    // nearest preceding timestamp will not work
    query := fmt.Sprintf("SELECT code, label, apply_from FROM %s ORDER BY apply_from",
            table)

    rows, err := db.Query(query)

    if err != nil {
        fmt.Printf(query)
        log.Fatal(err)
    }

    var code string
    var label string
    var applyfrom int64
    var labelmap map[string]LabelSet

    labelmap = make(map[string]LabelSet)
    for rows.Next() {
        rows.Scan(&code, &label, &applyfrom)

        if lmap, exists := labelmap[code]; exists {
            lmap.StartTimes = append(lmap.StartTimes, applyfrom)
            lmap.Labels = append(lmap.Labels, label)
            // TODO: would it be more efficient to store lmap as a pointer,
            // so we don't need to re-assign each time we update it?
            labelmap[code] = lmap
        } else {
            lset := LabelSet{}
            lset.StartTimes = append(lset.StartTimes, applyfrom)
            lset.Labels = append(lset.Labels, label)
            labelmap[code] = lset
        }
    }

    tagger.Replacements[replacecode] = FriendlyTag {replacecode, replacelabel,
            labelmap}

}

// LoadCountryLabels populates the Replacements map with the country code
// to country label mappings from the database
func (tagger *FriendlyTagger) LoadCountryLabels() {

    tagger.LoadGenericLabels(tagger.CountryLabelTable, "country_code",
            "country_label")
}

// LoadCountyLabels populates the Replacements map with the county code
// to county label mappings from the database
func (tagger *FriendlyTagger) LoadCountyLabels() {
    tagger.LoadGenericLabels(tagger.CountyLabelTable, "county_code",
            "county_label")
}

// LoadAsnLabels populates the Replacements map with the ASN
// to ASN label mappings from the database
func (tagger *FriendlyTagger) LoadAsnLabels() {
    tagger.LoadGenericLabels(tagger.AsnLabelTable, "asn", "asn_label")
}

// LoadRegionLabels populates the Replacements map with the region code
// to region label mappings from the database
func (tagger *FriendlyTagger) LoadRegionLabels() {
    tagger.LoadGenericLabels(tagger.RegionLabelTable, "region_code",
            "region_label")
}

// Apply takes a metric received by telegraf and adds
// the appropriate human-friendly labels for any geo-tags and ASNs that
// might be present in the metric.
func (tagger *FriendlyTagger) Apply(in ...telegraf.Metric) []telegraf.Metric {

    if len(in) < 1 {
        return in
    }

    timestamp := in[0].Time().Unix()

    // If we need to read the labels from database yet, do that first.
    if timestamp - tagger.LastReload >= tagger.ReloadFrequency {
        tagger.Replacements["continent_code"] = FriendlyTag {"continent_code", "continent_label", staticContinents}
        tagger.LoadCountryLabels()
        tagger.LoadCountyLabels()
        tagger.LoadRegionLabels()
        tagger.LoadAsnLabels()
        tagger.LastReload = timestamp
    }

    for i := 0; i < len(in); i++ {
        in[i] = tagger.InsertFriendlyLabels(in[i])
    }
    return in
}

// InsertFriendlyLabels scans a telegraf metric for any tags that match
// names where we can potentially add human-readable alternative labels,
// then looks up the value for the tag in our replacements map. If it finds
// a suitable label, then that label is added to the metric as another tag.
func (tagger *FriendlyTagger) InsertFriendlyLabels(metric telegraf.Metric) telegraf.Metric {

    timestamp := metric.Time().Unix()
    var toaddTags []string
    var toaddLabels []string

    // loop over all tags in this metric, looking for anything that matches
    // a known "replaceable" tag (e.g. asn or country_code)
    for _, tag := range metric.TagList() {
        if ftag, ok := tagger.Replacements[tag.Key]; ok {
            // now, do we have an entry for this tag value (code)
            if replace, alsook := ftag.ValueMappings[tag.Value]; alsook {
                // now find the right label for the timestamp attached to
                // this metric -- we do a reverse order loop here because
                // the most common use case will be real-time data where the
                // most recent label will be the correct match
                for i := len(replace.StartTimes) - 1; i >= 0; i-- {
                    if replace.StartTimes[i] <= timestamp {
                        // add a new tag with the label as the value
                        toaddTags = append(toaddTags, ftag.NewTag)
                        toaddLabels = append(toaddLabels, replace.Labels[i])
                        break
                    }
                }
            }
        }
    }

    for i := 0; i < len(toaddTags); i++ {
        metric.AddTag(toaddTags[i], toaddLabels[i])
    }

    return metric
}

// init registers this plugin with telegraf and constructs the FriendlyTagger
// instance. This method will set default table names if they are not provided
// in the config file).
func init() {
    processors.Add("friendlytagger", func() telegraf.Processor {
        return &FriendlyTagger{ReloadFrequency: int64(120), Replacements: make(map[string]FriendlyTag), CountryLabelTable: "country_mappings", CountyLabelTable: "county_mappings", AsnLabelTable: "asn_mappings", RegionLabelTable: "region_mappings", LastReload: int64(0)}
    })
}

// vim: set sw=4 tabstop=4 softtabstop=4 expandtab :
