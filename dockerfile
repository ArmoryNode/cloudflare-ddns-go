# syntax=docker/dockerfile:1
FROM golang

ENV APITOKEN ""
ENV ZONEIDENTIFIER ""
ENV DNSRECORDIDENTIFIER ""
ENV UPDATEINTERVAL 1

COPY . "/app"

WORKDIR /app

RUN ["go", "build", "-o=cloudflare-ddns"]

ENTRYPOINT ./cloudflare-ddns -apiToken=${APITOKEN} -zoneIdentifier=${ZONEIDENTIFIER} -dnsRecordIdentifier=${DNSRECORDIDENTIFIER} -updateInterval=${UPDATEINTERVAL}