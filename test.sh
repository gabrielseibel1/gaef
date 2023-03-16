#!/bin/sh

cd auth || exit
go test ./... --cover -count=1
cd ..

cd encounter-proposal || exit
go test ./... --cover -count=1
cd ..

cd group || exit
go test ./... --cover -count=1
cd ..

cd user || exit
go test ./... --cover -count=1
cd ..

cd client || exit
docker compose up -d --build
sleep 1s
go test ./... --cover -count=1
docker compose down --volumes
cd ..

