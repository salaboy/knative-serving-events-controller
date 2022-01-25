package handlers

import (
	"context"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol/http"

	servingclientset "knative.dev/serving/pkg/client/clientset/versioned"
	servingclient "knative.dev/serving/pkg/client/injection/client"
)

var (
	servingClient servingclientset.Interface
)

func StartReceiver(ctx context.Context) {
	servingClient = servingclient.Get(ctx)
	c, err := cloudevents.NewClientHTTP(http.WithPort(10000))
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Fatal(c.StartReceiver(context.Background(), TriggerEvent))
}
