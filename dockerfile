# syntax=docker/dockerfile:1
FROM golang

ENV ApiToken ""
ENV ZoneIdentifier ""
ENV DnsRecordIdentifier ""
ENV UpdateInterval 1

COPY . "/app"

WORKDIR /app

RUN ["go", "build", "-o=cloudflare-ddns"]

ENTRYPOINT ./cloudflare-ddns -apiToken=${ApiToken} -zoneIdentifier=${ZoneIdentifier} -dnsRecordIdentifier=${DnsRecordIdentifier} -updateInterval=${UpdateInterval}