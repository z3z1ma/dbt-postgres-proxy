dbt-postgres-proxy
==================

A reverse proxy for postgres which compiles queries in flight using a dbt rpc server. This is both a PoC as well as a dev tool that allows rapid prototying dbt models directly from any tool which is capable of connecting to Postgres. Think dbeaver, datagrip, BI tools, SQLAlchemy, etc.

Failure to rewrite the query will raise a NOTICE but it will not cause an error. It is likely in invalidly compiled dbt query will not be accepted by the database

All other messages than Query are passed to the upstream unmodified. SSL connections are not supported.

## Prerequisites:
- Go 1.17 or greater

## Build from source:
```
CGO_ENABLED=0 go build -o dbt-pg-proxy main.go
``` 

## Usage:
```
# Run compiled binary
./dbt-pg-proxy --help
  -dbtHost string
        dbt rpc server host (default "127.0.0.1")
  -dbtPort int
        dbt rpc server port (default 8580)
  -listen string
        Listen address (default "0.0.0.0:6432")
  -upstream string
        Upstream postgres server (default "127.0.0.1:5432")


# Running via go
go run cmd/main.go -listen="0.0.0.0:5111"
```

## Attribution
This work is an extension of https://github.com/patientsknowbest/pg-rewrite-proxy which both this and the original codebase leverage https://github.com/jackc/pgproto3
