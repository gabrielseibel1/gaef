FROM golang:1.19.4 AS build

WORKDIR /home/seibel/src/github.com/gabrielseibel1/gaef/group

RUN go install github.com/go-delve/delve/cmd/dlv@latest

COPY . .
RUN go mod download && go mod verify

RUN CGO_ENABLED=0 go build -gcflags="all=-N -l" -o server

EXPOSE 8080
EXPOSE 40000

ENTRYPOINT [ "dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/home/seibel/src/github.com/gabrielseibel1/gaef/group/server" ]