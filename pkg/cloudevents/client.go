package cloudevent

import (
	"context"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	uuid "github.com/satori/go.uuid"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/injection"
	"knative.dev/pkg/logging"
	v1 "knative.dev/serving/pkg/apis/serving/v1"

	cdevents "github.com/cdfoundation/sig-events/cde/sdk/go/pkg/cdf/events"
)

type KServiceEvent string

const (
	ServiceDeployed KServiceEvent = "deployed"
	ServiceUpgraded KServiceEvent = "upgraded"
	ServiceRemoved  KServiceEvent = "removed"
)

var (
	Map = KServiceToCDEventMap{
		ServiceDeployed: cdevents.ServiceDeployedEventV1,
		ServiceUpgraded: cdevents.ServiceUpgradedEventV1,
		ServiceRemoved:  cdevents.ServiceRemovedEventV1,
	}
)

type KServiceToCDEventMap map[KServiceEvent]cdevents.CDEventType

type CECKey struct{}

func init() {
	injection.Default.RegisterClient(withCloudEventClient)
}

func withCloudEventClient(ctx context.Context, cfg *rest.Config) context.Context {
	logger := logging.FromContext(ctx)

	protocol, err := cloudevents.NewHTTP()
	if err != nil {
		logger.Panicf("Error creating the cloudevents http protocol: %s", err)
	}

	cloudEventClient, err := cloudevents.NewClient(protocol, cloudevents.WithUUIDs(), cloudevents.WithTimeNow())
	if err != nil {
		logger.Panicf("Error creating the cloudevents client: %s", err)
	}

	return context.WithValue(ctx, CECKey{}, cloudEventClient)
}

func Get(ctx context.Context) cloudevents.Client {
	logger := logging.FromContext(ctx)

	untyped := ctx.Value(CECKey{})
	if untyped == nil {
		logger.Errorf(
			"Unable to fetch client from context.")
		return nil
	}

	client := untyped.(cloudevents.Client)
	return client
}

func ToContext(ctx context.Context, client cloudevents.Client) context.Context {
	return context.WithValue(ctx, CECKey{}, client)
}

func SendEvent(ctx context.Context, eventType KServiceEvent, obj *v1.Service) {
	logger := logging.FromContext(ctx)

	if eventType == ServiceDeployed {
		logger.Infof("SendEvent received %s event", ServiceDeployed)
	}

	cdEvent, ok := Map[eventType]
	if !ok {
		logger.Errorf("no known cloud event mapping found for event type %s", eventType)
		return
	}

	client := Get(ctx)

	switch eventType {
	case ServiceDeployed:
		event := createEvent(cdEvent.String(), obj)

		ctx := injectIntoContext(ctx, "http://localhost:8080")
		result := client.Send(ctx, event)
		if !cloudevents.IsACK(result) {
			logger.Warnf("Failed to send cloudevent: %s", result.Error())
		}

		if cloudevents.IsUndelivered(result) {
			logger.Errorf("failed sending cloud event, error: %s", result.Error())
		}

		logger.Infof("Sent event for %s type", ServiceDeployed)

	case ServiceUpgraded:
		event := createEvent(cdEvent.String(), obj)
		ctx := injectIntoContext(ctx, "http://localhost:8080")
		result := client.Send(ctx, event)
		if !cloudevents.IsACK(result) {
			logger.Warnf("Failed to send cloudevent: %s", result.Error())
		}

		if cloudevents.IsUndelivered(result) {
			logger.Errorf("cloud event undelivered, error: %s", result.Error())
		}

		logger.Infof("Sent event for %s type", ServiceUpgraded)

	default:
		logger.Warnf("unknown event type %s", eventType)
	}

}

func injectIntoContext(c context.Context, target string) context.Context {
	ctx := cloudevents.ContextWithRetriesExponentialBackoff(c, 10*time.Millisecond, 10)
	ctx = cloudevents.ContextWithTarget(ctx, target)

	return ctx
}

func createEvent(cdEventType string, obj *v1.Service) cloudevents.Event {
	event := cloudevents.NewEvent()
	event.SetSource(obj.GetNamespace() + "/" + obj.GetName())
	event.SetID(uuid.NewV4().String())
	event.SetType(cdEventType)
	event.SetTime(time.Now())

	return event
}
