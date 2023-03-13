#!/bin/sh

cd ..
docker compose up -d --build
sleep 1s
cd test
go run .
cd ..
docker compose down --volumes