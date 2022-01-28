package serving

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	cloudeventclient "github.com/salaboy/knative-serving-events-controller/pkg/cloudevents"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	v1 "knative.dev/serving/pkg/apis/serving/v1"

	clientset "knative.dev/serving/pkg/client/clientset/versioned"
)

const (
	LastActiveRevision = "experimental.serving.knative.dev/last-active-annotaion"

	Finalizer = "experimental.serving.knative.dev"
)

type Reconciler struct {
	tracker          tracker.Interface
	cloudEventClient cloudevents.Client

	config Config
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
	ctx = cloudeventclient.SetTarget(ctx, r.config.EventSink)
	logger.Infof("Reconciling %s", ksvc.Name)

	/*
		if !ksvc.ObjectMeta.DeletionTimestamp.IsZero() {
			logger.Info("service %s/%s deleted", ksvc.GetNamespace(), ksvc.GetName())

		} else {
			logger.Info("service deletion timestamp: %s", ksvc.ObjectMeta.DeletionTimestamp)
		}
	*/

	if !r.checkFinalizer(ctx, ksvc) {
		return r.setFinalizer(ctx, ksvc)
	}

	revision, ok := ksvc.Annotations[LastActiveRevision]
	if !ok {
		// logger.Infof("annotation does not exist, checking if latest revision is ready")
		logger.Info("possible creation event")
		err, ok := r.handleCreation(ctx, ksvc)
		if err != nil {
			return err
		}

		if ok {
			// trigger event for deployed
			logger.Infof("********* service deployed %s **********", ksvc.GetName())
			cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceDeployed, ksvc)

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
		cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceUpgraded, ksvc)

		return nil
	}

	if !ksvc.DeletionTimestamp.IsZero() {
		logger.Info("+++++++++++++++++++++ detected service deletion ++++++++++++++++++")
	}

	// TODO: implement spec upgrade detection
	/*
		if ksvc.ObjectMeta.Generation == 1 {
			// new object, log new event
			logger.Infof("service deployed %#v", ksvc)
			cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceDeployed, ksvc)
		} else if ksvc.ObjectMeta.Generation > 1 {
			logger.Infof("service upgraded %#v", ksvc)
			cloudeventclient.SendEvent(ctx, cloudeventclient.ServiceUpgraded, ksvc)
		}
	*/

	return nil
}

func (r *Reconciler) checkFinalizer(ctx context.Context, ksvc *v1.Service) bool {
	for _, f := range ksvc.Finalizers {
		if f == Finalizer {
			return true
		}
	}

	return false
}

func (r *Reconciler) setFinalizer(ctx context.Context, ksvc *v1.Service) error {
	logger := logging.FromContext(ctx)

	ksvc.Finalizers = append(ksvc.Finalizers, Finalizer)
	_, err := r.client.ServingV1().Services(ksvc.Namespace).Update(ctx, ksvc, metav1.UpdateOptions{})
	if err != nil {
		logger.Errorf("error adding finalizer, %s", err.Error())
		return err
	}

	return nil
}

func (r *Reconciler) handleCreation(ctx context.Context, ksvc *v1.Service) (error, bool) {
	logger := logging.FromContext(ctx)

	if ksvc.Status.LatestReadyRevisionName == "" {
		// not yet ready, this is a new deployment of this svc
		logger.Infof("ksvc deployed, waiting for revision")
		return nil, false
	}

	// set the revision
	logger.Infof("updating annotation with new revision")
	ksvc.Annotations[LastActiveRevision] = ksvc.Status.LatestReadyRevisionName
	_, err := r.client.ServingV1().Services(ksvc.GetNamespace()).Update(ctx, ksvc, metav1.UpdateOptions{})
	if err != nil {
		logger.Errorf("error updating revision annotation: %s", err.Error())

		return err, false
	}

	logger.Info("added revision metadata for the service, deployed initial version")
	return nil, true
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
