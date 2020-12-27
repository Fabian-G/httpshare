package resolve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type myIP struct {
	IP      string `json:"ipaddress"`
	Version string `json:"ipversion"`
}

// GetMyIP gets your public ip from a remote server. Defaults to 127.0.0.1 on errors
func GetMyIP() string {
	res, err := http.Get("https://api.meineip.eu/?format=json")
	if err != nil || res.StatusCode != http.StatusOK {
		return "127.0.0.1"
	}
	ipResponse := &myIP{}
	err = json.NewDecoder(res.Body).Decode(ipResponse)
	if err != nil {
		return "127.0.0.1"
	}

	return ipResponse.IP
}

// FormatIPForURL basically puts brackets [] around the IP if it is IPv6
func FormatIPForURL(ip string) string {
	if strings.Contains(ip, ":") {
		return fmt.Sprintf("[%s]", ip)
	}
	return ip
}