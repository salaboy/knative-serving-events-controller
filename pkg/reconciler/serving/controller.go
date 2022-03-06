package serving

import (
	"context"
	"time"

	"github.com/kelseyhightower/envconfig"
	"k8s.io/client-go/tools/cache"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/tracker"

	cloudeventclient "github.com/salaboy/knative-serving-events-controller/pkg/cloudevents"
	"github.com/salaboy/knative-serving-events-controller/pkg/server/handlers"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	revisioninformer "knative.dev/serving/pkg/client/injection/informers/serving/v1/revision"
	serviceinformer "knative.dev/serving/pkg/client/injection/informers/serving/v1/service"

	servicereconciler "knative.dev/serving/pkg/client/injection/reconciler/serving/v1/service"

	servingclient "knative.dev/serving/pkg/client/injection/client"
)

type Config struct {
	EventSink string `default:"http://localhost:8080"`
}

func NewController(ctx context.Context, cmw configmap.Watcher) *controller.Impl {
	logger := logging.FromContext(ctx)

	var cfg Config

	err := envconfig.Process("", &cfg)
	if err != nil {
		panic(err)
	}

	servingInformer := serviceinformer.Get(ctx)
	revisionInformer := revisioninformer.Get(ctx)

	r := &Reconciler{
		client:           servingclient.Get(ctx),
		cloudEventClient: cloudeventclient.Get(ctx),
		config:           cfg,
	}

	go handlers.StartReceiver(ctx)

	impl := servicereconciler.NewImpl(ctx, r, func(impl *controller.Impl) controller.Options {
		return controller.Options{
			FinalizerName: "experimental.serving.knative.dev",
		}
	})

	logger.Info("Setting up event handlers")
	servingInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	grcb := func(obj interface{}) {
		impl.GlobalResync(servingInformer.Informer())
	}

	handleControllerOf := cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterController(&servingv1.Service{}),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	}

	revisionInformer.Informer().AddEventHandler(handleControllerOf)
	servingInformer.Informer().AddEventHandler(controller.HandleAll(grcb))

	r.tracker = tracker.New(impl.EnqueueKey, 30*time.Minute)

	return impl
}
