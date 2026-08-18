package dbutil

import (
	"fmt"
	"net/url"
	"strings"
)

func RawDSNToURL(dsn string) string {
	if strings.HasPrefix(dsn, "mysql://") {
		return dsn
	}
	atIdx := strings.Index(dsn, "@")
	if atIdx < 0 {
		return dsn
	}
	userinfo := dsn[:atIdx]
	rest := dsn[atIdx+1:]
	tcpPrefix := "tcp("
	if !strings.HasPrefix(rest, tcpPrefix) {
		return dsn
	}
	rest = strings.TrimPrefix(rest, tcpPrefix)
	closeIdx := strings.Index(rest, ")")
	if closeIdx < 0 {
		return dsn
	}
	hostport := rest[:closeIdx]
	dbpart := rest[closeIdx+1:]
	if strings.HasPrefix(dbpart, "/") {
		dbpart = dbpart[1:]
	}
	return fmt.Sprintf("mysql://%s@%s/%s", url.PathEscape(userinfo), hostport, dbpart)
}
