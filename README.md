dbt-postgres-proxy
==================

![proxy-example](/static/dbt_postgres_proxy_demo.gif)

A reverse proxy for postgres which compiles queries in flight using a dbt rpc server. This is both a PoC as well as a dev tool that allows rapid prototying dbt models directly from any tool which is capable of connecting to Postgres. Think dbeaver, datagrip, BI tools, SQLAlchemy, etc.

Failure to rewrite the query will raise a NOTICE but it will not cause an error.

Queries which detect Jinja syntax are compiled and the result is written to stdout for inspection during the development cycle.

All other messages than Query are passed to the upstream unmodified. SSL connections are not supported. If you planned on using this transiently in production, would recommend using it on a bastion server with database traffic whitelisted for only that servers IP. Otherwise locally, it's fair game.

Requires a **running dbt-rpc server to interface with**. The default arguments for the CLI flags point to the default host and port of the rpc server. Forward looking, the interface may be rewritten to interface with dbt-server when it is released or for more permissive licensing, a custom server solution. As is, this is a fantastic tool for development flows that can be utilized anywhere a postgres connection would (Tableau, SQLAlchemy, JDBC, etc) for interactive testing and dev. 


## Prerequisites:
- Go 1.17 or greater

## Build from source:
```
CGO_ENABLED=0 go build -o dbt-pg-proxy main.go
``` 

## Usage:
```
# See help when running command for available flags
./dbt-pg-proxy --help
  -dbtHost string
        dbt rpc server host (default "127.0.0.1")
  -dbtPort int
        dbt rpc server port (default 8580)
  -listen string
        Listen address (default "0.0.0.0:6432")
  -upstream string
        Upstream postgres server (default "127.0.0.1:5432")

# Example Usage
dbt-rpc --host 127.0.0.1 --port 8580
./dbt-pg-proxy-linux-amd64 \
    -dbtHost="127.0.0.1" \
    -dbtPort="8580" \
    -listen="0.0.0.0:6432" \
    -upstream="my_db.a8dfs982hd.us-west-2.rds.amazonaws.com:5432"

# Running via go
go run cmd/main.go \
    -listen="0.0.0.0:6432" \
    -upstream="127.0.0.1:5432"
```

## Attribution
This work is an extension of https://github.com/patientsknowbest/pg-rewrite-proxy which both this and the original codebase leverage https://github.com/jackc/pgproto3
