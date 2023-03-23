## Build
FROM golang:1.19.4 AS build

# TODO use general relative path
WORKDIR /home/seibel/src/github.com/gabrielseibel1/gaef

RUN go install github.com/go-delve/delve/cmd/dlv@latest

COPY . .

RUN CGO_ENABLED=0 go build -gcflags="all=-N -l" -o user/server github.com/gabrielseibel1/gaef/user

EXPOSE 8080
EXPOSE 40000

# TODO use general relative path
WORKDIR /home/seibel/src/github.com/gabrielseibel1/gaef/user

ENTRYPOINT [ "dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/home/seibel/src/github.com/gabrielseibel1/gaef/user/server" ]