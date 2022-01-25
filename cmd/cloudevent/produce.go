package main

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"knative.dev/sample-controller/pkg/server/models"
)

func main() {
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	// Create an Event.
	event := cloudevents.NewEvent()
	event.SetSource("example/uri")
	event.SetType(models.CreateKService.String())
	event.SetData(cloudevents.ApplicationJSON, map[string]string{
		"name":      "hello",
		"namespace": "default",
		"image":     "gcr.io/google-samples/hello-app:1.0",
	})

	// Set a target.
	ctx := cloudevents.ContextWithTarget(context.Background(), "http://localhost:10000/")

	// Send that Event.
	if result := c.Send(ctx, event); cloudevents.IsUndelivered(result) {
		log.Fatalf("failed to send, %v", result)
	}
}
