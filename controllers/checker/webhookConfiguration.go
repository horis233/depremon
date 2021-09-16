package checker

import (
	"context"
	"regexp"
	"time"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/horis233/k8s-deprecation-checker/controllers/handler"
)

func WebhookConfigurationChecks(client client.Client, config *rest.Config) error {
	dc := discovery.NewDiscoveryClientForConfigOrDie(config)
	_, apiLists, err := dc.ServerGroupsAndResources()
	if err != nil {
		return err
	}
	for _, apiList := range apiLists {
		if apiList.GroupVersion == "admissionregistration.k8s.io/v1beta1" {
			if err := wait.PollImmediateInfinite(time.Minute*3, func() (done bool, err error) {
				mutatingwebhookconfigurations := &admissionregistrationv1.MutatingWebhookConfigurationList{}
				if err := client.List(context.TODO(), mutatingwebhookconfigurations); err != nil {
					return false, err
				}
				for _, mu := range mutatingwebhookconfigurations.Items {
					if len(mu.ManagedFields) == 0 {
						continue
					}
					for _, template := range mu.ManagedFields {
						if template.APIVersion == "admissionregistration.k8s.io/v1beta1" {
							apiFromRequest := handler.DeprecatedObjectList{
								Group:   "admissionregistration.k8s.io",
								Version: "v1beta1",
								Kind:    mu.Kind,
								Objects: []handler.DeprecatedObject{
									{
										Name: mu.Name,
										RequesterList: []string{
											template.Manager,
										},
									},
								},
							}
							klog.Info(mu.Name)
							err = handler.UpdateConfigmap(context.TODO(), client, apiFromRequest)
							if err != nil {
								return false, err
							}
							break
						}
					}
				}
				validatingwebhookconfigurations := &admissionregistrationv1.ValidatingWebhookConfigurationList{}
				if err := client.List(context.TODO(), validatingwebhookconfigurations); err != nil {
					return false, err
				}
				for _, va := range validatingwebhookconfigurations.Items {
					if len(va.ManagedFields) == 0 {
						continue
					}
					for _, template := range va.ManagedFields {
						if template.APIVersion == "admissionregistration.k8s.io/v1beta1" {
							regocp, err := regexp.Compile(`^(.*)openshift\.io`)
							if err != nil {
								klog.Error(err)
							}
							regk8s, err := regexp.Compile(`^(.*)k8s\.io`)
							if err != nil {
								klog.Error(err)
							}
							if regocp.MatchString(va.Name) || regk8s.MatchString(va.Name) {
								break
							}
							klog.Info(va.Name)
							apiFromRequest := handler.DeprecatedObjectList{
								Group:   "admissionregistration.k8s.io",
								Version: "v1beta1",
								Kind:    va.Kind,
								Objects: []handler.DeprecatedObject{
									{
										Name: va.Name,
										RequesterList: []string{
											template.Manager,
										},
									},
								},
							}
							err = handler.UpdateConfigmap(context.TODO(), client, apiFromRequest)
							if err != nil {
								return false, err
							}
							break
						}
					}
				}
				return false, nil
			}); err == nil {
				break
			} else {
				return err
			}

		}
	}
	return nil
}
