package resolve

import (
	"fmt"
	"strings"
)

// IPResolver resolves an IPAddresse
type IPResolver interface {
	Resolve() string
}

// FormatIPForURL basically puts brackets [] around the IP if it is IPv6
func FormatIPForURL(ip string) string {
	if strings.Contains(ip, ":") {
		return fmt.Sprintf("[%s]", ip)
	}
	return ip
}