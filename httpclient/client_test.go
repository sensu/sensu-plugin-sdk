package httpclient_test

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/httpclient"
)

func TestClientGet(t *testing.T) {
	config := httpclient.CoreClientConfig{
		URL:    server.URL,
		APIKey: "use transport layer security",
		CACert: server.Certificate(),
	}
	cl := httpclient.NewCoreClient(config)
	req := httpclient.NewEventRequest("default", "server", "network")
	event := new(corev2.Event)
	_, err := cl.GetResource(context.Background(), req, event)
	if err != nil {
		t.Fatal(err)
	}
	if event.Timestamp == 0 {
		t.Fatal("0 timestamp")
	}
}

func TestClientPut(t *testing.T) {
	config := httpclient.CoreClientConfig{
		URL:    server.URL,
		APIKey: "use transport layer security",
		CACert: server.Certificate(),
	}
	cl := httpclient.NewCoreClient(config)
	check := corev2.FixtureCheckConfig("fake")
	req := httpclient.ResourceRequest{Resource: check}
	_, err := cl.PutResource(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientPost(t *testing.T) {
	config := httpclient.CoreClientConfig{
		URL:    server.URL,
		APIKey: "use transport layer security",
		CACert: server.Certificate(),
	}
	cl := httpclient.NewCoreClient(config)
	check := corev2.FixtureCheckConfig("fake")
	req := httpclient.ResourceRequest{Resource: check}
	_, err := cl.PostResource(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
}

func TestClientDelete(t *testing.T) {
	config := httpclient.CoreClientConfig{
		URL:    server.URL,
		APIKey: "use transport layer security",
		CACert: server.Certificate(),
	}
	cl := httpclient.NewCoreClient(config)
	check := corev2.FixtureCheckConfig("fake")
	req := httpclient.ResourceRequest{Resource: check}
	_, err := cl.DeleteResource(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
}
