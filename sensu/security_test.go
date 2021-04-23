package sensu_test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sensu/sensu-plugin-sdk/sensu"
)

func TestSecurityConfig(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer server.Close()
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(tf.Name())
	}()
	if _, err := tf.Write(server.Certificate().Raw); err != nil {
		t.Fatal(err)
	}
	if err := tf.Close(); err != nil {
		t.Fatal(err)
	}
	cfg := sensu.SecurityConfig{
		CACertificate: tf.Name(),
	}
	cert, err := cfg.GetCACertificate()
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{}
	rootCAs := x509.NewCertPool()
	rootCAs.AddCert(cert)
	client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: rootCAs,
		},
	}
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Do(req); err != nil {
		t.Fatal(err)
	}
}
