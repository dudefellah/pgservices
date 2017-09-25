# pgservices

A little Go module to parse and return details from a Postgres Services 
(pg_service.conf) file.

I like using the definitions in the pg_service.conf file to connect to 
Postgres, but didn't find any Go packages that quite fit for me. As a 
result, I wrote this little module to help me load pg_service.conf contents
and build a struct out of the contents. This should make it easy to then
generate your own connection pg string and connect to the database.

This package makes use of the Go [INI lib](https://gopkg.in/ini.v1) to do all of the
heavy lifting, but does a small amount of additional checking before loading
the pg_service.conf contents.

# Synopsis

You should really only need to use the ParsePgServices function to make
use of this package.

```go
package main

import (
    "fmt"
    "os"

    "github.com/dudefellah/pgservices"
)

func main() {
// Load pg services
    fileReader, err := os.Open("/etc/postgresql-common/pg_service.conf")
    if err != nil {
        fmt.Printf("Error opening file %v\n", err)
        os.Exit(1)
    }

    defer fileReader.Close()

    pgServices, err := pgservices.ParsePgServices(fileReader)
    if err != nil {
        fmt.Printf("Crap! %v\n", err)
        os.Exit(1)
    }

    connectionString := myCustomFunctionThatGeneratesAConnectionStringFromTheParsedPgService(pgServices)
    ...
}

```