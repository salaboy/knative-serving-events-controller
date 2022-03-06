package serving

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventclient "github.com/salaboy/knative-serving-events-controller/pkg/cloudevents"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	v1 "knative.dev/serving/pkg/apis/serving/v1"

	clientset "knative.dev/serving/pkg/client/clientset/versioned"
)

const (
	LastActiveRevision = "experimental.serving.knative.dev/last-active-annotation"

	// Finalizer = "experimental.serving.knative.dev"
)

type Reconciler struct {
	tracker          tracker.Interface
	cloudEventClient cloudevents.Client

	config Config
	client clientset.Interface
}

// FinalizeKind implements custom logic to finalize v1.Service. Any changes
// to the objects .Status or .Finalizers will be ignored. Returning a nil or
// Normal type reconciler.Event will allow the finalizer to be deleted on
// the resource. The resource passed to FinalizeKind will always have a set
// deletion timestamp.
func (r *Reconciler) FinalizeKind(ctx context.Context, o *v1.Service) reconciler.Event {
	logger := logging.FromContext(ctx)

	ctx = cloudeventclient.ToContext(ctx, r.cloudEventClient)
	ctx = cloudeventclient.SetTarget(ctx, r.config.EventSink)
	logger.Info("Received finalizer event for ", o.GetNamespace(), o.GetName())

	err := cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceRemoved, o)
	if err != nil {
		return err
	}

	return nil
}

// ObserveDeletion implements custom logic to observe deletion of the respective resource
// with the given key.
func (r *Reconciler) ObserveDeletion(ctx context.Context, key types.NamespacedName) error {
	logger := logging.FromContext(ctx)

	logger.Info("Received deletion event for ", key)

	return nil
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
	ctx = cloudeventclient.SetTarget(ctx, r.config.EventSink)
	logger.Infof("Reconciling %s", ksvc.Name)

	revision, ok := ksvc.Annotations[LastActiveRevision]
	if !ok {
		// logger.Infof("annotation does not exist, checking if latest revision is ready")
		logger.Info("possible creation event")
		ok, err := r.handleCreation(ctx, ksvc)
		if err != nil {
			return err
		}

		if ok {
			// trigger event for deployed
			logger.Infof("********* service deployed %s **********", ksvc.GetName())
			err := cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceDeployed, ksvc)
			if err != nil {
				return err
			}
		}

		return nil
	}

	if revision != ksvc.Status.LatestReadyRevisionName {
		// check if the annotation revision different than latest ready revision
		logger.Infof("service revision different %s", revision)
		logger.Infof("revision has been upgraded, marking the same")

		err := r.handleUpgrade(ctx, ksvc)
		if err != nil {
			return err
		}

		// trigger event for revision upgrade
		logger.Infof("************** revision upgraded *************")
		err = cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceUpgraded, ksvc)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func (r *Reconciler) handleCreation(ctx context.Context, ksvc *v1.Service) (bool, error) {
	logger := logging.FromContext(ctx)

	if ksvc.Status.LatestReadyRevisionName == "" {
		// not yet ready, this is a new deployment of this svc
		logger.Infof("ksvc deployed, waiting for revision")
		return false, nil
	}

	// set the revision
	logger.Infof("updating annotation with new revision")
	ksvc.Annotations[LastActiveRevision] = ksvc.Status.LatestReadyRevisionName
	_, err := r.client.ServingV1().Services(ksvc.GetNamespace()).Update(ctx, ksvc, metav1.UpdateOptions{})
	if err != nil {
		logger.Errorf("error updating revision annotation: %s", err.Error())

		return false, err
	}

	logger.Info("added revision metadata for the service, deployed initial version")
	return true, nil
}

func (r *Reconciler) handleUpgrade(ctx context.Context, ksvc *v1.Service) error {
	logger := logging.FromContext(ctx)

	ksvc.Annotations[LastActiveRevision] = ksvc.Status.LatestReadyRevisionName
	_, err := r.client.ServingV1().Services(ksvc.GetNamespace()).Update(ctx, ksvc, metav1.UpdateOptions{})
	if err != nil {
		logger.Errorf("error updating revision annotation: %s", err.Error())

		return err
	}

	logger.Info("updated annotation revision")

	return nil
}
