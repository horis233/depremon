/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	operatorv1alpha1 "github.com/horis233/k8s-deprecation-checker/api/v1alpha1"
	"github.com/horis233/k8s-deprecation-checker/controllers/handler"
	"github.com/horis233/k8s-deprecation-checker/controllers/utils"
	"github.com/horis233/k8s-deprecation-checker/controllers/webhooks"
)

// DeprapiscanReconciler reconciles a Deprapiscan object
type DeprapiscanReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Manager *manager.Manager
}

//+kubebuilder:rbac:groups=operator.horis233.com,resources=deprapiscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.horis233.com,resources=deprapiscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.horis233.com,resources=deprapiscans/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps;services;secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations;validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Deprapiscan object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *DeprapiscanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	instance := &operatorv1alpha1.Deprapiscan{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	namespace, err := utils.GetOperatorNamespace()
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.setupWebhooks(namespace); err != nil {
		klog.Error(err, "Error setting up webhook server")
	}

	// Reconcile the webhooks
	if err := webhooks.Config.Reconcile(ctx, r.Client, instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeprapiscanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.Deprapiscan{}).
		Complete(r)
}

func (r *DeprapiscanReconciler) setupWebhooks(namespace string) error {

	klog.Info("Creating deprcated api checker webhook configuration")
	webhooks.Config.AddWebhook(webhooks.CSWebhook{
		Name:        "ibm-deprcated-api-record",
		WebhookName: "deprecateapi.operator.horis233.com",
		Rules: []webhooks.RuleWithOperations{
			webhooks.NewRule().
				OneResource("networking.k8s.io", "v1beta1", "ingresses").
				ForCreate().
				NamespacedScope(),
			webhooks.NewRule().
				OneResource("networking.k8s.io", "v1beta1", "ingressclasses").
				ForCreate().
				NamespacedScope(),
			webhooks.NewRule().
				OneResource("apiextensions.k8s.io", "v1beta1", "customresourcedefinitions").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("admissionregistration.k8s.io", "v1beta1", "mutatingwebhookconfigurations").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("admissionregistration.k8s.io", "v1beta1", "validatingwebhookconfigurations").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("apiregistration.k8s.io", "v1beta1", "apiservices").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("coordination.k8s.io", "v1beta1", "leases").
				ForCreate().
				NamespacedScope(),
			webhooks.NewRule().
				OneResource("rbac.authorization.k8s.io", "v1beta1", "roles").
				ForCreate().
				NamespacedScope(),
			webhooks.NewRule().
				OneResource("rbac.authorization.k8s.io", "v1beta1", "rolebindings").
				ForCreate().
				NamespacedScope(),
			webhooks.NewRule().
				OneResource("rbac.authorization.k8s.io", "v1beta1", "clusterroles").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("rbac.authorization.k8s.io", "v1beta1", "clusterrolebindings").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("scheduling.k8s.io", "v1beta1", "priorityclasses").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("storage.k8s.io", "v1beta1", "csidrivers").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("storage.k8s.io", "v1beta1", "csinodes").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("storage.k8s.io", "v1beta1", "storageclasses").
				ForCreate().
				ClusterScope(),
			webhooks.NewRule().
				OneResource("storage.k8s.io", "v1beta1", "volumeattachments").
				ForCreate().
				ClusterScope(),
		},
		Register: webhooks.AdmissionWebhookRegister{
			Type: webhooks.ValidatingType,
			Path: "/deprecate-api-check",
			Hook: &admission.Webhook{
				Handler: &handler.Recorder{
					Client: r.Client,
				},
			},
		},
	})

	klog.Info("setting up webhook server")
	if err := webhooks.Config.SetupServer(*r.Manager, namespace); err != nil {
		return err
	}
	return nil
}
