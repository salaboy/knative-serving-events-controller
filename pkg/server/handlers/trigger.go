package handlers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"knative.dev/pkg/logging"
	"github.com/salaboy/knative-serving-events-controller/pkg/server/models"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func TriggerEvent(event cloudevents.Event) {
	ctx := context.Background()
	logger := logging.FromContext(ctx)
	fmt.Printf("%s\n", event)

	if event.Type() == models.CreateKService.String() {
		var ksvc models.KService
		err := event.DataAs(&ksvc)
		if err != nil {
			fmt.Printf("error receiving data as ksvc. error: %s\n", err.Error())
		}

		fmt.Println("*************************************")
		fmt.Println("event of type create service received")
		fmt.Println("*************************************")

		_, err = servingClient.ServingV1().Services(ksvc.Namespace).Create(context.TODO(), &servingv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: ksvc.Name,
			},
			Spec: servingv1.ServiceSpec{
				ConfigurationSpec: servingv1.ConfigurationSpec{
					Template: servingv1.RevisionTemplateSpec{
						Spec: servingv1.RevisionSpec{
							PodSpec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Image: ksvc.Image,
									},
								},
							},
						},
					},
				},
			},
		}, metav1.CreateOptions{})

		if err != nil {
			logger.Errorf("unable to create ksvc. error: %s", err.Error())
		}
	}
}
