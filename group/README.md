# gaef-group-service

Groups service for Groups and Encounters Finder.

## Running

Provide .env files in the services folders with the necessary environment variables (see each main.go).

Verify code quality:

```
go fmt ./...
go vet ./...
go test ./...
```

Run the server with a mongoDB using docker compose:

```
docker compose up -d --build
```

Shutdown the server:

```
docker compose down
```
