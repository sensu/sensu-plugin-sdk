package sensu

import (
	"crypto/x509"
	"io/ioutil"
)

// SecurityConfig holds configuration for securely communicating with a Sensu
// backend.
type SecurityConfig struct {
	// CACertificate provide a means to use a self-signed certificate with
	// an HTTP client.
	CACertificate string

	// InsecureSkipVerify skips hostname verification for certificates. It is
	// not recommended to use this outside of testing.
	InsecureSkipVerify bool
}

// SensuSecurityOptions adds the following flags to a plugin:
//   --sensu-ca-cert
//   --sensu-insecure-skip-verify
func SensuSecurityOptions(config *SecurityConfig) []ConfigOption {
	return []ConfigOption{
		&PluginConfigOption[string]{
			Value:    &config.CACertificate,
			Path:     "sensu-ca-cert",
			Env:      "SENSU_CA_CERT",
			Argument: "sensu-ca-cert",
			Usage:    "--sensu-ca-cert /etc/ssl/self-signed-ca.crt",
		},
		&PluginConfigOption[bool]{
			Value:    &config.InsecureSkipVerify,
			Path:     "sensu-insecure-skip-verify",
			Env:      "SENSU_INSECURE_SKIP_VERIFY",
			Argument: "sensu-insecure-skip-verify",
			Usage:    "--sensu-insecure-skip-verify (disables TLS hostname verification)",
		},
	}
}

// GetCACertificate gets the CA certificate associated with the path stored at
// CACertificate. It returns an error if the file is not found, or if the
// certificate is not a valid x509 certificate. The certificate can be provided
// to the CoreClient in the httpclient package.
func (s *SecurityConfig) GetCACertificate() (*x509.Certificate, error) {
	b, err := ioutil.ReadFile(s.CACertificate)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(b)
}
