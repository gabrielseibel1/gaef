#!/bin/sh

# vet
go vet github.com/gabrielseibel1/gaef/auth/...
go vet github.com/gabrielseibel1/gaef/encounter-proposal/...
go vet github.com/gabrielseibel1/gaef/group/...
go vet github.com/gabrielseibel1/gaef/user/...

# unit tests
go test github.com/gabrielseibel1/gaef/auth/... --cover -count=1
go test github.com/gabrielseibel1/gaef/encounter-proposal/... --cover -count=1
go test github.com/gabrielseibel1/gaef/group/... --cover -count=1
go test github.com/gabrielseibel1/gaef/user/... --cover -count=1

# integration tests
docker compose up -d --build
sleep 1s
go test github.com/gabrielseibel1/gaef/client/... --cover -count=1
docker compose down --volumes

