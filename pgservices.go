package pgservices

import (
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"

	logging "github.com/op/go-logging"

	ini "gopkg.in/ini.v1"
)

const pgServiceName = "pgservices"

// SslModes makes it easier to check the sslmode option
var SslModes = []string{
	"disable",
	"allow",
	"prefer",
	"require",
	"verify-ca",
	"verify-full",
}

var log = logging.MustGetLogger(pgServiceName)
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// postgresService is an individual service definition's
// (in the context of pg_service.conf) worth of values
type postgresService struct {
	DBName   string
	Host     string
	HostAddr net.IP
	Port     int
	User     string
	Password string

	ConnectTimeout int
	ClientEncoding string

	Options string

	ApplicationName         string
	FallbackApplicationName string

	KeepAlives         bool
	KeepAlivesIdle     int
	KeepAlivesInterval int
	KeepAlivesCount    int

	TTY string // ignored by postgres, but it could legitimately appear in a pg_service.conf

	SSLMode        string
	RequireSSL     bool
	SSLCompression bool
	SSLCert        string
	SSLKey         string
	SSLRootCert    string
	SSLCrl         string

	RequirePeer string

	KrbSrvname string
	GSSLib     string
}

// PostgresServiceGroup is the object that holds
// all of the service definitions found in a
// pg_service.conf file. It's basically the parsed
// version of that file
//
// No public methods are exposed, so you can just
// get any of the details you'd like by accessing
// PostgresServiceGroup's map of postgresService objects,
// each of which contains all of the necessary pg
// connection data
type PostgresServiceGroup struct {
	Category map[string]postgresService
}

// String method prints a string representation of
// the object. I was using this for debugging, but it
// could potentially be used to output a valid pg_service.conf
// file
func (p postgresService) String() string {
	passwordStr := "<none>"
	if len(p.Password) > 0 {
		passwordStr = "<defined>"
	}

	keepAlivesInt := 0
	if p.KeepAlives {
		keepAlivesInt = 1
	}

	requireSSLInt := 0
	if p.RequireSSL {
		requireSSLInt = 1
	}

	sslCompressionInt := 0
	if p.SSLCompression {
		sslCompressionInt = 1
	}

	hostAddrStr := ""
	if p.HostAddr != nil {
		hostAddrStr = string(p.HostAddr)
	}

	return fmt.Sprintf(`dbname = %s
host = %s
hostaddr = %v
port = %d
user = %s
password = %s

connect_timeout = %d
client_encoding = %s

options = %s

application_name = %s
fallback_application_name = %s

keepalives = %d
keepalives_idle = %d
keepalives_interval = %d
keepalives_count = %d

sslmode = %s
requiressl = %d
sslcompression = %d
sslcert = %s
sslkey = %s
sslrootcert = %s
sslcrl = %s

requirepeer = %s

krbsrvname = %s
gsslib = %s
`,
		p.DBName,
		p.Host,
		hostAddrStr,
		p.Port,
		p.User,
		passwordStr,
		p.ConnectTimeout,
		p.ClientEncoding,
		p.Options,
		p.ApplicationName,
		p.FallbackApplicationName,
		keepAlivesInt,
		p.KeepAlivesIdle,
		p.KeepAlivesInterval,
		p.KeepAlivesCount,
		p.SSLMode,
		requireSSLInt,
		sslCompressionInt,
		p.SSLCert,
		p.SSLKey,
		p.SSLRootCert,
		p.SSLCrl,
		p.RequirePeer,
		p.KrbSrvname,
		p.GSSLib,
	)
}

// Set method lets you set values in a postgresService
// struct by using string key/value pairs. This is convenient
// when you're reading in a bunch of key/value pairs from
// a file and want to set those values in the associated object
// easily
//
// postgresService doesn't have an associated Get method right now
// since that's kind of redundant. I don't have the same need to
// quickly grab data from the postgresService object by using strings
// as keys. Despite the redundancy, it's somewhat tempting to add
// simply for symmetry's sake.
func (p *postgresService) Set(
	k string,
	v string,
) error {
	structPtr := reflect.ValueOf(p)
	// struct
	structElem := structPtr.Elem()

	field := structElem.FieldByName(k)
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("Invalid field %s for PostgresService", k)
	}

	if field.Kind() == reflect.Int {
		intVal, err := strconv.Atoi(v)
		if err != nil {
			return err
		}
		field.SetInt(int64(intVal))
	} else if field.Kind() == reflect.String {
		// There's probably a better way to be smarter?
		// For now, we're checking some specific input values
		// (eg. sslmode) to make sure it falls into an appropriate
		// list of options (require, disable, etc)
		if strings.ToLower(k) == "sslmode" {
			validString := false
			for i := range SslModes {
				if SslModes[i] == v {
					validString = true
					break
				}
			}

			if validString {
				field.SetString(v)
			}
		} else {
			field.SetString(v)
		}
	} else if field.Kind() == reflect.Bool {
		lowerVal := strings.ToLower(v)
		if lowerVal == "" {
			return fmt.Errorf("No value provided for bool type %s", k)
		} else if lowerVal == "false" || lowerVal == "f" || lowerVal == "0" {
			field.SetBool(false)
		} else {
			field.SetBool(true)
		}
	} else {
		return fmt.Errorf("Field type %v is unhandled.", field.Kind())
	}

	return nil
}

// addService adds a PostgresService object to the PostgresServiceGroup
// Category map, but with an extra check to make sure it isn't already defined.
func (p PostgresServiceGroup) addService(
	name string,
	pgService postgresService,
) error {
	if _, ok := p.Category[name]; ok {
		return fmt.Errorf("A postgres service named `%s' already exists", name)
	}

	p.Category[name] = pgService

	return nil
}

// New function for creating PostgresServiceGroup objects. This is
// just handy since we need to make sure we initialize the map
// contained in the struct during creation and this function does
// that for us.
func New(
	bufReader io.Reader,
) *PostgresServiceGroup {
	postgres := new(PostgresServiceGroup)
	postgres.Category = make(map[string]postgresService)

	return postgres
}

// ParsePgServices parses the pg_services.conf file contents by
// reading the supplied io.Reader object. Results are dropped into
// the PostgresServiceGroup object pointer and returned
func ParsePgServices(
	bufReader io.Reader,
) (*PostgresServiceGroup, error) {
	pgServices := New(bufReader)

	cfg, err := ini.Load(bufReader)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	for _, section := range cfg.Sections() {
		log.Debugf("Retrieved config[%s]: %v", section.Name(), section)

		pgServiceCategory := postgresService{}
		for key, value := range section.KeysHash() {
			structMember, err := pgServiceKeyToStructMember(key)
			if err != nil {
				return nil, err
			}

			err = pgServiceCategory.Set(structMember, value)
			if err != nil {
				return nil, err
			}
		}

		pgServices.addService(section.Name(), pgServiceCategory)
	}

	return pgServices, nil
}
