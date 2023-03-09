# gaef-user-service

Users service for Groups and Encounters Finder.
Authenticates, stores and retrieves users.

## Running

Provide a .env file in the repo root with the necessary environment variables (see [](main.go)).

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

