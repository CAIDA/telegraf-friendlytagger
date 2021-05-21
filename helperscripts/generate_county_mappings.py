#!/usr/bin/python3

import string, re
import wandio

regioncodes = "http://loki.caida.org:3282/gadm/polygons/gadm.counties.v2.0.processed.polygons.csv.gz"

try:
        with wandio.open(regioncodes, mode='rb') as fh:
                for l in fh:
                        if ',' not in l:
                                continue
                        s = l.strip().split(',')
                        if s[2] == "name":
                                continue
                        if s[2] == "\"\"" or s[2] == "\"?\"":
                                s[2] = "\"Unknown\""
                        print ("%s,%s,0,0" % (s[2], s[0]))
except IOError as err:
        raise err
