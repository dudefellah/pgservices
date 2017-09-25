package pgservices

import (
	"fmt"
	"strings"
)

// Just for convenience. There's no simple way to
// take a left-side value from the pg_service.conf
// file and convert it to a properly formatted member
// variable, (at least not one that also satisfies golint)
// but since there's only a small set of variables,
// we'll just directly convert them here
func pgServiceKeyToStructMember(
	key string,
) (string, error) {
	switch strings.ToLower(key) {
	case "host":
		return "Host", nil
	case "hostaddr":
		return "HostAddr", nil
	case "port":
		return "Port", nil
	case "dbname":
		return "DBName", nil
	case "user":
		return "User", nil
	case "password":
		return "Password", nil
	case "connect_timeout":
		return "ConnectTimeout", nil
	case "client_encoding":
		return "ClientEncoding", nil
	case "options":
		return "Options", nil
	case "application_name":
		return "ApplicationName", nil
	case "fallback_application_name":
		return "FallbackApplicationName", nil
	case "keepalives":
		return "KeepAlives", nil
	case "keepalives_idle":
		return "KeepAlivesIdle", nil
	case "keepalives_interval":
		return "KeepAlivesInterval", nil
	case "keepalives_count":
		return "KeepAlivesCount", nil
	case "tty":
		return "TTY", nil
	case "sslmode":
		return "SSLMode", nil
	case "sslcompression":
		return "SSLCompression", nil
	case "sslkey":
		return "SSLKey", nil
	case "sslrootcert":
		return "SSLRootCert", nil
	case "sslcrl":
		return "SSLCrl", nil
	case "requirepeer":
		return "RequirePeer", nil
	case "krbsrvname":
		return "KrbSrvname", nil
	case "gsslib":
		return "GSSLib", nil
	case "sslcert":
		return "SSLCert", nil
	}

	return "", fmt.Errorf("Invalid postgres service key value %s", key)
}
