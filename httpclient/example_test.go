package httpclient_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/httpclient"
)

var server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		event := corev2.FixtureEvent("server", "network")
		_ = json.NewEncoder(w).Encode(event)
	}
}))

func ExampleCoreClient_GetResource() {
	config := httpclient.CoreClientConfig{
		URL:    server.URL,
		APIKey: "use transport layer security",
		CACert: server.Certificate(),
	}
	cl := httpclient.NewCoreClient(config)
	req := httpclient.NewEventRequest("default", "server", "network")
	event := new(corev2.Event)
	resp, err := cl.GetResource(context.Background(), req, event)
	if err != nil {
		panic(err)
	}
	fmt.Println("HTTP Response", resp.Status)
	fmt.Println("Event at", time.Unix(event.Timestamp, 0))
}

func ExampleCoreClient_PutResource() {
	config := httpclient.CoreClientConfig{
		URL:    server.URL,
		APIKey: "use transport layer security",
		CACert: server.Certificate(),
	}
	cl := httpclient.NewCoreClient(config)
	check := corev2.FixtureCheckConfig("fake")
	req := httpclient.ResourceRequest{Resource: check}
	resp, err := cl.PutResource(context.Background(), req)
	if err != nil {
		panic(err)
	}
	fmt.Println("HTTP Response", resp.Status)
}

func ExampleCoreClient_PostResource() {
	config := httpclient.CoreClientConfig{
		URL:    server.URL,
		APIKey: "use transport layer security",
		CACert: server.Certificate(),
	}
	cl := httpclient.NewCoreClient(config)
	check := corev2.FixtureCheckConfig("fake")
	req := httpclient.ResourceRequest{Resource: check}
	resp, err := cl.PostResource(context.Background(), req)
	if err != nil {
		panic(err)
	}
	fmt.Println("HTTP Response", resp.Status)
}

func ExampleCoreClient_DeleteResource() {
	config := httpclient.CoreClientConfig{
		URL:    server.URL,
		APIKey: "use transport layer security",
		CACert: server.Certificate(),
	}
	cl := httpclient.NewCoreClient(config)
	check := corev2.FixtureCheckConfig("fake")
	req := httpclient.ResourceRequest{Resource: check}
	resp, err := cl.DeleteResource(context.Background(), req)
	if err != nil {
		panic(err)
	}
	fmt.Println("HTTP Response", resp.Status)
}

func ExampleNewResourceRequest() {
	req, err := httpclient.NewResourceRequest("core/v2", "CheckConfig", "default", "disk")
	if err != nil {
		panic(err)
	}
	fmt.Println(req.Resource.URIPath())
	// Output: /api/core/v2/namespaces/default/checks/disk
}
