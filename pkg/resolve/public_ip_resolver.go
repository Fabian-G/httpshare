package resolve

import (
	"encoding/json"
	"net/http"
)

// PublicIPResolver resolves the public ip by querying another server.
// If an error occurs the Fallback resolver is used. Zero value is *not* usable.
type PublicIPResolver struct {
	Fallback IPResolver
}

// NewPublicIPResolver creaes a new PublicIpResolver. fallback must not be null.
func NewPublicIPResolver(fallback IPResolver) *PublicIPResolver {
	return &PublicIPResolver{fallback}
}

// Resolve gets your public ip from a remote server. Defaults to fallback resolver on errors
func (p *PublicIPResolver) Resolve() string {
	res, err := http.Get("https://api.meineip.eu/?format=json")
	if err != nil || res.StatusCode != http.StatusOK {
		return p.Fallback.Resolve()
	}
	ipResponse := &struct {
		IP      string `json:"ipaddress"`
		Version string `json:"ipversion"`
	}{}
	err = json.NewDecoder(res.Body).Decode(ipResponse)
	if err != nil {
		return p.Fallback.Resolve()
	}

	return ipResponse.IP
}
