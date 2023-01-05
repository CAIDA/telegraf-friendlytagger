FROM golang:alpine AS build-env
ENV GO111MODULE=auto
RUN apk update && apk add --no-cache make git bash gcc sed build-base linux-headers

WORKDIR /telegraf
RUN mkdir -p /go/src/github.com/influxdata
RUN git clone https://github.com/influxdata/telegraf.git /go/src/github.com/influxdata/telegraf

COPY src /telegraf/src
COPY helperscripts /telegraf/helperscripts
RUN mkdir -p /go/src/github.com/influxdata/telegraf/plugins/processors/friendlytagger && cp src/friendlytagger.go /go/src/github.com/influxdata/telegraf/plugins/processors/friendlytagger/

RUN cp /telegraf/src/register_ft.go /go/src/github.com/influxdata/telegraf/plugins/processors/all/

RUN cd /go/src/github.com/influxdata/telegraf/ && go mod tidy && make && go install -ldflags "-w -s" ./cmd/telegraf
RUN cd helperscripts && go mod tidy && go build -o /go/bin/queryasnnames .

FROM alpine:latest
RUN apk update && apk add --no-cache bash sqlite

WORKDIR /app/
COPY --from=build-env /go/bin/telegraf /app/
COPY --from=build-env /go/bin/queryasnnames /app/
COPY docker/entrypoint.sh /entrypoint.sh
COPY docker/friendlytag.db /app/

EXPOSE 8125/udp 8092/udp 8094

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/app/telegraf"]

