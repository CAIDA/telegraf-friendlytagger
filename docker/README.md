### Quickstart

    cd ../
    docker build -t friendlytag -f docker/Dockerfile .

Update Kafka and Influx configuration in `docker/telegraf.conf` (see below
for more details).

    docker run --name <choose a name> -v $(pwd)/docker/telegraf.conf:/etc/telegraf/telegraf.conf:ro friendlytag telegraf


### Purpose

Runs telegraf with the CAIDA "friendly tagger" plugin within a Docker container.

The telegraf instance will perform the following tasks:

 * Read collected measurements from a kafka topic
 * Look for tags that refer to geolocations or ASNs (e.g. country codes,
   region IDs, AS numbers).
 * For each tag found, add an additional tag with a user-friendly label for
   the tagged identifier (e.g. a country code of 'AU' will be labelled
   'Australia'). ASNs will be augmented with the AS name (as per whois data).
 * The resulting augmented measurements will be written into a specified
   Influx time series database.

### Building the Docker image

Starting from the directory containing this README file:

    cd ../
    docker build -t friendlytag -f docker/Dockerfile .


### Running the container

First, update `docker/telegraf.conf` to point the telegraf instance at the
kafka broker that will be providing the input data (these options can be found
under `inputs.kafka_consumer`). You will want to update the `brokers` and
`topics` options. You will also need to replace every instance of `reptest`
in the templates list with the prefix that has been prepended to each result
emitted by the system that is generating the input data. If your input data
is coming from a corsaro report plugin, the prefix will be the value of
the `output_row_label` option set in the report plugin configuration.

Next, update `docker/telegraf.conf` to point the telegraf instance at the
Influx database that will be the recipient of your results. The influx
configuration can be found under `outputs.influxdb`. In this case, you
will want to check the `urls`, `database`, `username`, and `password` options
are correct.

From the directory where you built the docker image (i.e. `../` from the
directory containing this README), you can now run:

    docker run --name <choose a name> -v $(pwd)/docker/telegraf.conf:/etc/telegraf/telegraf.conf:ro friendlytag telegraf

This will run your container in the foreground and allow you to stop the
container by sending it a SIGINT via Ctrl-C. The container will begin by
querying the ASRank API for all of the ASN label mappings; each discovered ASN
and its label will be dumped to standard output during this process. Once this
is complete, telegraf will begin consuming from your kafka topic.

Once you're happy with how the container is running, you'll probably want to
run it in a `screen` or as a daemon process (possibly redirecting all the
logging noise somewhere else).


### Updating the ASN labels on a running container

To trigger a fresh query of the ASRank API to get updated ASN label mappings,
simply run the following on the container host:

    docker exec <container name> /app/queryasnnames /app/friendlytag.db


### Updating the geolocation labels

Occasionally, geopolitical events may cause the set of country, region or
county labels to change and therefore you may wish to update the database on
your container to reflect this.

This is a slightly more involved process than updating the ASN labels, but
hopefully it should seldom be necessary. You will need to re-create the
internal database used to map tags to labels, then rebuild the docker image
and restart your container. The container should quickly catch up with any
incoming data that arrives while the container is re-built.

If a new country has come into existence, you will need to manually add the
ISO-2 code and country name into the `helperscripts/countries.map` file
*before* running the steps below. New countries come about so infrequently
that it wasn't worth the effort of scripting the mapping generation, but
this could probably be done easily enough if it proves to be an issue. If you
are adding a new country, consider committing that change back to the
telegraf-friendlytagger git repository so that other users can benefit from
the addition as well.

New counties or regions will not require any manual file edits, as these will
be scraped from the latest geolocation data available -- just follow
the steps below directly.

Steps to follow:

    <configure your swift environment variables for geo-data file access>

    docker stop <container name>
    cd docker/
    rm friendlytag.db
    ./initdb.sh
    cd ..
    docker build -t friendlytag -f docker/Dockerfile .
    docker run --name <container name> -v $(pwd)/docker/telegraf.conf:/etc/telegraf/telegraf.conf:ro friendlytag telegraf



