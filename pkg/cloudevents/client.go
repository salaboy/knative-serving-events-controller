package cloudevent

import (
	"context"
	"fmt"
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

type Key int

const (
	EventSink Key = iota
)

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

func SetTarget(ctx context.Context, target string) context.Context {
	return context.WithValue(ctx, EventSink, target)
}

func SendEvent(ctx context.Context, eventType KServiceEvent, obj *v1.Service) error {
	logger := logging.FromContext(ctx)

	if eventType == ServiceDeployed {
		logger.Infof("SendEvent received %s event", ServiceDeployed)
	}

	cdEvent, ok := Map[eventType]
	if !ok {
		err := fmt.Errorf("no known cloud event mapping found for event type %s", eventType)
		logger.Error(err)
		return err
	}

	client := Get(ctx)

	event := createEvent(cdEvent.String(), obj)

	target := ctx.Value(EventSink).(string)

	ctx = injectIntoContext(ctx, target)
	result := client.Send(ctx, event)
	logger.Info("result", result)
	if !cloudevents.IsACK(result) {
		err := fmt.Errorf("Failed to send cloudevent, error: %s", result.Error())
		logger.Error(err)

		return err
	}

	if cloudevents.IsUndelivered(result) {
		err := fmt.Errorf("cloud event cannot be delivered, error: %s", result.Error())
		logger.Error(err)

		return err
	}

	return nil
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
