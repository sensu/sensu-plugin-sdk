package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

// ResourceRequest specifies a request for a resource. Use NewResourceRequest
// to create a new ResourceRequest. As a special case, use NewEventRequest
// when working with events.
type ResourceRequest struct {
	corev2.TypeMeta
	corev2.ObjectMeta
	types.Resource
}

func (r ResourceRequest) String() string {
	return fmt.Sprintf("<%s.%s>(%s/%s)", r.APIVersion, r.Type, r.Namespace, r.Name)
}

// NewResourceRequest creates a new ResourceRequest. ResourceRequests are a
// specification of a Sensu type, and a resource's unique name, determined
// by the namespace and name.
//
// This generic method of resource lookup does not work for Events.
func NewResourceRequest(apiVersion, typeName, namespace, name string) (ResourceRequest, error) {
	resource, err := types.ResolveType(apiVersion, typeName)
	if err != nil {
		return ResourceRequest{}, err
	}
	resource.SetObjectMeta(corev2.ObjectMeta{
		Namespace: namespace,
		Name:      name,
	})
	return ResourceRequest{
		TypeMeta: corev2.TypeMeta{
			APIVersion: apiVersion,
			Type:       typeName,
		},
		ObjectMeta: corev2.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Resource: resource,
	}, nil
}

// NewEventRequest creates a request for an Event, based on its entity and
// check name.
func NewEventRequest(namespace, entity, check string) ResourceRequest {
	return ResourceRequest{
		TypeMeta: corev2.TypeMeta{
			APIVersion: "core/v2",
			Type:       "Event",
		},
		ObjectMeta: corev2.ObjectMeta{
			Namespace: namespace,
			Name:      path.Join(entity, check),
		},
		Resource: &corev2.Event{
			Check: &corev2.Check{
				ObjectMeta: corev2.ObjectMeta{
					Namespace: namespace,
					Name:      check,
				},
			},
			Entity: &corev2.Entity{
				ObjectMeta: corev2.ObjectMeta{
					Namespace: namespace,
					Name:      entity,
				},
			},
		},
	}
}

// HTTPError is an error type that holds the HTTP response code and body.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (h HTTPError) Error() string {
	return fmt.Sprintf("error %d: %s", h.StatusCode, h.Body)
}

// CoreClient is a simple HTTP client for accessing the core Sensu API.
type CoreClient struct {
	HTTPClient http.Client
	Config     CoreClientConfig
}

// CoreClientConfig contains the configuration information needed for a CoreClient.
type CoreClientConfig struct {
	// URL is the server URL.
	URL string

	// APIKey is the Sensu API key.
	APIKey string

	// CACert, if non-nil, will be used to configure TLS communication. This
	// is only needed when using a self-signed certificate.
	CACert *x509.Certificate
}

func newRequest(ctx context.Context, resource corev2.Resource, verb, server, apikey string) (*http.Request, error) {
	var (
		req      *http.Request
		err      error
		location = server + resource.URIPath()
	)

	switch verb {
	case http.MethodGet:
		req, err = http.NewRequestWithContext(ctx, verb, location, nil)
	case http.MethodPut, http.MethodPost:
		body, berr := json.Marshal(resource)
		if berr != nil {
			return nil, berr
		}
		req, err = http.NewRequestWithContext(ctx, verb, location, bytes.NewReader(body))
	case http.MethodDelete:
		req, err = http.NewRequestWithContext(ctx, verb, location, nil)
	default:
		return nil, fmt.Errorf("method not supported: %s", verb)
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", apikey))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// GetResource is a generic method for getting a Sensu resource. The client will
// consult the contents of the ResourceRequest in order to fill the request.
//
// Once the request is completed, the contents of the response will be read into
// the types.Resource passed in.
//
// If the server returns a reponse code greater than 400, an HTTPError will be
// returned with the status code and the first 64KB of the response.
//
// If the HTTP request is completed successfully, whether or not a 4xx error
// occurred, then a non-nil http.Response will be returned with its response
// body closed.
func (c *CoreClient) GetResource(ctx context.Context, r ResourceRequest, in types.Resource) (*http.Response, error) {
	req, err := newRequest(ctx, r.Resource, http.MethodGet, c.Config.URL, c.Config.APIKey)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := validateResponse(resp); err != nil {
		return resp, err
	}
	reader := io.LimitReader(resp.Body, 1<<24)
	return resp, json.NewDecoder(reader).Decode(&in)
}

// DeleteResource is a generic method for deleting a Sensu resource.
//
// If the server returns a reponse code greater than 400, an HTTPError will be
// returned with the status code and the first 64KB of the response.
//
// If the HTTP request is completed successfully, whether or not a 4xx error
// occurred, then a non-nil http.Response will be returned with its response
// body closed.
func (c *CoreClient) DeleteResource(ctx context.Context, r ResourceRequest) (*http.Response, error) {
	req, err := newRequest(ctx, r.Resource, http.MethodDelete, c.Config.URL, c.Config.APIKey)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, validateResponse(resp)
}

// PutResource is a generic method for creating a Sensu resource.
//
// If the server returns a reponse code greater than 400, an HTTPError will be
// returned with the status code and the first 64KB of the response.
//
// If the HTTP request is completed successfully, whether or not a 4xx error
// occurred, then a non-nil http.Response will be returned with its response
// body closed.
func (c *CoreClient) PutResource(ctx context.Context, r ResourceRequest) (*http.Response, error) {
	req, err := newRequest(ctx, r.Resource, http.MethodPut, c.Config.URL, c.Config.APIKey)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, validateResponse(resp)
}

// PutResource is a generic method for creating a Sensu resource. Make sure the
// resource supplied has an ObjectMeta configured.
//
// If the server returns a reponse code greater than 400, an HTTPError will be
// returned with the status code and the first 64KB of the response.
//
// If the HTTP request is completed successfully, whether or not a 4xx error
// occurred, then a non-nil http.Response will be returned with its response
// body closed.
func (c *CoreClient) PostResource(ctx context.Context, r ResourceRequest) (*http.Response, error) {
	req, err := newRequest(ctx, r.Resource, http.MethodPost, c.Config.URL, c.Config.APIKey)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, validateResponse(resp)
}

func validateResponse(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}
	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return err
	}
	return HTTPError{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}
}

// NewCoreClient creates a new core API client that uses the supplied CoreClientConfig.
// Once the client is created, the embedded http.Client can be manipulated as desired.
func NewCoreClient(config CoreClientConfig) *CoreClient {
	client := &CoreClient{
		Config: config,
	}
	// Set up CA cert if provided. By default, the Go HTTP client uses the
	// system cert pool.
	if config.CACert != nil {
		rootCAs, err := x509.SystemCertPool()
		if err != nil {
			log.Println(err)
			rootCAs = x509.NewCertPool()
		}

		rootCAs.AddCert(config.CACert)

		client.HTTPClient.Transport.(*http.Transport).TLSClientConfig = &tls.Config{
			RootCAs: rootCAs,
		}
	}
	return client
}
