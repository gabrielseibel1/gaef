#!/bin/sh

# integration tests
docker compose -f compose/docker-compose.yml up -d --build
sleep 10s
go test github.com/gabrielseibel1/gaef/client/... --cover -count=1
docker compose -f compose/docker-compose.yml down --volumes
