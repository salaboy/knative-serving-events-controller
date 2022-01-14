package serving

import (
	"context"
	"time"

	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"

	cloudeventclient "knative.dev/sample-controller/pkg/cloudevents"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	revisioninformer "knative.dev/serving/pkg/client/injection/informers/serving/v1/revision"
	serviceinformer "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"

	servicereconciler "knative.dev/serving/pkg/client/injection/reconciler/serving/v1/service"

	servingclient "knative.dev/serving/pkg/client/injection/client"
)

func NewController() func(context.Context, configmap.Watcher) *controller.Impl {
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		logger := logging.FromContext(ctx)

		servingInformer := serviceinformer.Get(ctx)
		revisionInformer := revisioninformer.Get(ctx)

		r := &Reconciler{
			client:           servingclient.Get(ctx),
			cloudEventClient: cloudeventclient.Get(ctx),
		}

		impl := servicereconciler.NewImpl(ctx, r, func(impl *controller.Impl) controller.Options {
			return controller.Options{}
		})

		logger.Info("Setting up event handlers")
		servingInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

		handleControllerOf := cache.FilteringResourceEventHandler{
			FilterFunc: controller.FilterController(&servingv1.Service{}),
			Handler:    controller.HandleAll(impl.EnqueueControllerOf),
		}

		revisionInformer.Informer().AddEventHandler(handleControllerOf)

		r.tracker = tracker.New(impl.EnqueueKey, 30*time.Minute)

		return impl
	}
}
