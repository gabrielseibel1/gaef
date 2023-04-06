#!/bin/sh

# integration tests
docker compose up -d --build
sleep 1s
go test github.com/gabrielseibel1/gaef/client/... --cover -count=1
docker compose down --volumes
