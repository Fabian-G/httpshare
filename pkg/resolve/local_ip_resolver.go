package resolve

import "net"

const (
	localhostAddresse = "127.0.0.1"
)

// LocalIPResolver resolves to a local ip addresse. Defaults to 127.0.0.1.
// Zero value is usable.
type LocalIPResolver struct {

}

// Resolve resolves to the first local non loopback addresse it finds. 
// Defaults to 127.0.0.1
func (l *LocalIPResolver) Resolve() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return localhostAddresse
	}
	for _, iface := range ifaces {
		if !isRelevantInterface(iface) {
			continue 
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.String() == "<nil>" {
				continue
			}
			return ip.String()
		}
	}
	return localhostAddresse
}

func isRelevantInterface(iface net.Interface) bool {
	return (iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0)
}