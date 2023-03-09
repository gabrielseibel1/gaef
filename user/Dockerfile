## Build
FROM golang:1.19.4 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 go build -o server

## Deploy
FROM scratch

WORKDIR /app

COPY --from=build /app/.env .
COPY --from=build /app/server .

EXPOSE 8080

ENTRYPOINT [ "/app/server","--production=true" ]