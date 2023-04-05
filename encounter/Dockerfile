## Build
FROM golang:1.19.4 AS build

WORKDIR /build

COPY . .
RUN CGO_ENABLED=0 go build -o encounter/webserver github.com/gabrielseibel1/gaef/encounter

## Deploy
FROM scratch

WORKDIR /app

COPY --from=build /build/encounter/.env .
COPY --from=build /build/encounter/webserver .

EXPOSE 8080

ENTRYPOINT [ "/app/webserver","--production=true" ]