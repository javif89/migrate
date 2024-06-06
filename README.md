# Migrate

[Laravel](https://laravel.com/docs/11.x/migrations) inspired migration style, with the Go flexibility.

# Usage

## CLI tool

```bash
migrate create create_my_table # Will create a migration named [year]_[month]_[day]_hms_create_my_table.sql in your migration path
migrate # Run the migrations
```

## In code

```go
package main

import "github.com/javif89/migrate"

func main() {
    m := migrate.New("./database/migrations", database.DriverMysql, database.Config{
        Host: "localhost",
        Port: "3306",
        Username: "dbuser",
        Password: "password",
        Database: "testdb",
    })

    m.Migrate() // Run unexecuted migrations
    m.Rollback() // Rollback the last batch of migrations
    m.Fresh() // Wipe DB and run migrations
}
```

# Configuration

Configuration is pretty straight forward. You just need a .env file with a few variables.

You can run `migrate init` to create the .env file in the current directory. Then just fill in the
values for your dabase.

```dotenv
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=user
DB_PASSWORD=password
DB_DATABASE=mydb
DB_DRIVER=mysql
MIGRATIONS_PATH=./database/migrations
```

# Writing migrations

When you run `migrate create [migration name]` it will create a `.sql` file in your `MIGRATIONS_PATH` folder
that looks like this:

```sql
-- UP --

-- DOWN --
```

Just write the up part of your migration under `-- UP --` and the down portion under `-- DOWN --`