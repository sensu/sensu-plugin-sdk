package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-plugin-sdk/httpclient"
)

func main() {
	config := httpclient.CoreClientConfig{
		URL:    "http://localhost:8080",
		APIKey: "af8a7a28-5030-4c52-9f15-1deab3defff7",
	}
	client := httpclient.NewCoreClient(config)
	event := new(corev2.Event)
	req := httpclient.NewEventRequest("default", "localhost.localdomain", "keepalive")
	_, err := client.GetResource(context.Background(), req, event)
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(event)
}
