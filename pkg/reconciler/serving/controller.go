package serving

import (
	"context"
	"time"

	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"

	serviceinformer "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"
	servicereconciler "knative.dev/serving/pkg/client/injection/reconciler/serving/v1/service"
)

func NewController() func(context.Context, configmap.Watcher) *controller.Impl {
	return func(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
		logger := logging.FromContext(ctx)

		servingInformer := serviceinformer.Get(ctx)
		r := &Reconciler{}

		impl := servicereconciler.NewImpl(ctx, r, func(impl *controller.Impl) controller.Options {
			return controller.Options{}
		})

		logger.Info("Setting up event handlers")
		servingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    impl.Enqueue,
			UpdateFunc: controller.PassNew(impl.Enqueue),
			DeleteFunc: impl.Enqueue,
		})

		r.tracker = tracker.New(impl.EnqueueKey, 30*time.Minute)

		return impl
	}
}
