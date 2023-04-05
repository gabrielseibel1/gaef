## Build
FROM golang:1.19.4 AS build

WORKDIR /build

COPY . .
RUN CGO_ENABLED=0 go build -o group/server github.com/gabrielseibel1/gaef/group

## Deploy
FROM scratch

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/group/server .

EXPOSE 8080

ENTRYPOINT [ "/app/server" ]