package handlers

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"

	servingclientset "knative.dev/serving/pkg/client/clientset/versioned"
	servingclient "knative.dev/serving/pkg/client/injection/client"
)

var (
	servingClient servingclientset.Interface
)

func StartReceiver(ctx context.Context) {
	servingClient = servingclient.Get(ctx)
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), TriggerEvent))
}
