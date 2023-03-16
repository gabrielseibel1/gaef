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

docker compose up -d --build
sleep 1s
cd client || exit
go test ./... --cover -count=1
cd ..
docker compose down --volumes

