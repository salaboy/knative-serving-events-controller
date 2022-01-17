package serving

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	cloudeventclient "knative.dev/sample-controller/pkg/cloudevents"
	v1 "knative.dev/serving/pkg/apis/serving/v1"

	clientset "knative.dev/serving/pkg/client/clientset/versioned"
)

type Reconciler struct {
	tracker          tracker.Interface
	cloudEventClient cloudevents.Client

	client clientset.Interface
}

// ReconcileKind implements custom logic to reconcile v1.Service. Any changes
// to the objects .Status or .Finalizers will be propagated to the stored
// object. It is recommended that implementors do not call any update calls
// for the Kind inside of ReconcileKind, it is the responsibility of the calling
// controller to propagate those properties. The resource passed to ReconcileKind
// will always have an empty deletion timestamp.
func (r *Reconciler) ReconcileKind(ctx context.Context, ksvc *v1.Service) reconciler.Event {
	logger := logging.FromContext(ctx)
	ctx = cloudeventclient.ToContext(ctx, r.cloudEventClient)
	logger.Infof("Reconciling %s", ksvc.Name)

	/*
		if !ksvc.ObjectMeta.DeletionTimestamp.IsZero() {
			logger.Info("service %s/%s deleted", ksvc.GetNamespace(), ksvc.GetName())

		} else {
			logger.Info("service deletion timestamp: %s", ksvc.ObjectMeta.DeletionTimestamp)
		}
	*/
	if ksvc.ObjectMeta.Generation == 1 {
		// new object, log new event
		logger.Infof("service deployed %v", ksvc)
		cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceDeployed, ksvc)
	} else if ksvc.ObjectMeta.Generation > 1 {
		logger.Infof("service upgraded %#v", ksvc)
		cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceUpgraded, ksvc)
	}

	return nil
}
