package serving

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
)

type Reconciler struct {
	tracker          tracker.Interface
	cloudEventClient *cloudevents.Client
}

// ReconcileKind implements custom logic to reconcile v1.Service. Any changes
// to the objects .Status or .Finalizers will be propagated to the stored
// object. It is recommended that implementors do not call any update calls
// for the Kind inside of ReconcileKind, it is the responsibility of the calling
// controller to propagate those properties. The resource passed to ReconcileKind
// will always have an empty deletion timestamp.
func (r *Reconciler) ReconcileKind(ctx context.Context, ksvc *v1.Service) reconciler.Event {
	logger := logging.FromContext(ctx)

	logger.Infof("Reconciling %s", ksvc.Name)

	return nil
}
