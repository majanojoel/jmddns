package externip

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

var (
	ErrUnexpectedResponse = errors.New("external ip response was unexpected")
	ErrIPNotV4            = errors.New("external ip was not an IPv4 address")
)

type (
	HTTPExternalIPProvider struct{}
)

const (
	externalIPProviderURL = "https://ifconfig.me/ip"
)

func NewHTTPExternalIPProvider() (*HTTPExternalIPProvider, error) {
	return &HTTPExternalIPProvider{}, nil
}

func (p *HTTPExternalIPProvider) GetExternalIP() (string, error) {
	request, err := http.NewRequest(http.MethodGet, externalIPProviderURL, nil)
	if err != nil {
		return "", fmt.Errorf("http.NewRequest: %w", err)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("(*http.Client).Do: %w", err)
	}
	// We should get back only an IP string here.
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("io.ReadAll: %w", err)
	}
	parsedIP := net.ParseIP(string(bytes))
	if parsedIP.To4() == nil {
		return "", ErrIPNotV4
	}
	return parsedIP.String(), nil
}
