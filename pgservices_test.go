package pgservices

import (
	"bytes"
	"io/ioutil"
	"testing"
)

// TestNoFile checks some bad buffers to make sure nothing
// dumb happens.
// New dumb ideas should be added as we go

func TestInvalidBuffer(t *testing.T) {
	badBuffers := [...]string{
		`butts`,
		`[cat`,
		`[category]
blah`,
		`[category]
blah=wat`,
	}

	for _, buffer := range badBuffers {
		byteReadCloser := ioutil.NopCloser(bytes.NewReader([]byte(buffer)))
		_, err := ParsePgServices(byteReadCloser)

		if err == nil {
			t.Errorf("Parser didn't complain and it should've!")
		}
	}
}

func TestValidBuffer(t *testing.T) {
	goodBuffers := [...]string{
		`[service_one]
host = localhost
port = 5432
dbname = test_db
sslmode = disable`,
		`[service_one]
host = localhost
port = 5432
dbname = test_db
sslmode = required
sslcrl = /etc/ssl/crl/root.crl
sslrootcert = /etc/postgres/ssl/rootcert.crt
sslcert = /etc/ssl/cert/postgres.crt
keepalives = 1
[service_two]
host = db.example.com
port = 5432
dbname = test_db
user = dbuser
password = abc123
`,
	}

	for _, buffer := range goodBuffers {
		byteReadCloser := ioutil.NopCloser(bytes.NewReader([]byte(buffer)))
		_, err := ParsePgServices(byteReadCloser)

		if err != nil {
			t.Errorf("Error parsing buffer: %v", err)
		}
	}
}
