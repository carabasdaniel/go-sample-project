package testharness

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
)

// CreateClient creates a new http client that can talk to the API
func (h *TestHarness) CreateClient() *http.Client {
	caCert, err := os.ReadFile(h.GoSampleProject.Configuration.API.Gateway.Certs.TLSCACertPath)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
				MinVersion: tls.VersionTLS12,
		}}}

	return client
}
