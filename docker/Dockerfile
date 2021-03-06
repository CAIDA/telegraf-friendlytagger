FROM golang:alpine AS build-env
RUN apk update && apk add --no-cache make git bash gcc sed build-base linux-headers

WORKDIR /telegraf

RUN go get github.com/influxdata/telegraf
RUN go get github.com/mattn/go-sqlite3
COPY src /telegraf/src
COPY helperscripts /telegraf/helperscripts
RUN mkdir -p /go/src/github.com/influxdata/telegraf/plugins/processors/friendlytagger && cp src/friendlytagger.go /go/src/github.com/influxdata/telegraf/plugins/processors/friendlytagger/
RUN sed -i '8i\        _ "github.com/influxdata/telegraf/plugins/processors/friendlytagger"' /go/src/github.com/influxdata/telegraf/plugins/processors/all/all.go

RUN cd /go/src/github.com/influxdata/telegraf/ && make && go install -ldflags "-w -s" ./cmd/telegraf
RUN cd helperscripts && go build -o /go/bin/queryasnnames .

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

