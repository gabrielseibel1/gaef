# gaef

Groups And Encounters Finder

## Running

```shell
docker compose -f compose/docker-compose.yml up -d --build
docker compose -f compose/docker-compose.yml down
```

## Debugging

```shell
docker compose -f compose/docker-compose-debug-<service>.yml up -d --build
docker compose -f compose/docker-compose-debug-<service>.yml down
```

## Testing

### Unit

```shell
test/unit_test.sh
```

### Integration

```shell
test/integration_test.sh
```