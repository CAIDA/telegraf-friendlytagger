//go:build !custom || processors || processors.friendlytagger

package all

import _ "github.com/influxdata/telegraf/plugins/processors/friendlytagger" // register plugin
