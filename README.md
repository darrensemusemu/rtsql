# rtsql

Real-time SQL updates

> Note: still WIP (proof of concept)

## Quick Start

Setup a test postgres docker db

```bash
docker-compose -f quickstart/postgres/compose.yaml up -d
```

Run go program

```bash
go run ./...
```

Insert data into users table and . Optional just run `quickstart/postgres/insert-users.sql`.

```sql
insert into "user"("email")	values('john'),
```

Watch console program will real time inserted values triggered by a DB insert.

## SQL Support

- [ ] PostgreSQL
- [ ] MySql
- [ ] SQLite
- [ ] SQL Server
