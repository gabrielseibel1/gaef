#!/bin/sh

cd ..
docker compose up -d --build
sleep 1s
cd test
go test ./... --cover -v
cd ..
docker compose down --volumes