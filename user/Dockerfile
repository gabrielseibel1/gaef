## Build
FROM golang:1.19.4 AS build

WORKDIR /build

COPY . .
RUN CGO_ENABLED=0 go build -o user/server github.com/gabrielseibel1/gaef/user

## Deploy
FROM scratch

WORKDIR /app

COPY --from=build /build/user/server .

EXPOSE 8080

ENTRYPOINT [ "/app/server" ]