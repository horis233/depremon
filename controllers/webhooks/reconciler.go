package webhooks

import (
	"context"
	"fmt"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/horis233/k8s-deprecation-checker/controllers/utils"
)

// WebhookReconciler knows how to reconcile webhook configuration CRs
type WebhookReconciler interface {
	SetName(name string)
	SetWebhookName(webhookName string)
	SetRule(rules []RuleWithOperations)
	SetNsSelector(selector v1.LabelSelector)
	Reconcile(ctx context.Context, client k8sclient.Client, caBundle []byte) error
}

type CompositeWebhookReconciler struct {
	Reconcilers []WebhookReconciler
}

func (reconciler *CompositeWebhookReconciler) SetName(name string) {
	for _, innerReconciler := range reconciler.Reconcilers {
		innerReconciler.SetName(name)
	}
}

func (reconciler *CompositeWebhookReconciler) SetWebhookName(webhookName string) {
	for _, innerReconciler := range reconciler.Reconcilers {
		innerReconciler.SetWebhookName(webhookName)
	}
}

func (reconciler *CompositeWebhookReconciler) SetRule(rules []RuleWithOperations) {
	for _, innerReconciler := range reconciler.Reconcilers {
		innerReconciler.SetRule(rules)
	}
}

func (reconciler *CompositeWebhookReconciler) SetNsSelector(selector v1.LabelSelector) {
	for _, innerReconciler := range reconciler.Reconcilers {
		innerReconciler.SetNsSelector(selector)
	}
}

func (reconciler *CompositeWebhookReconciler) Reconcile(ctx context.Context, client k8sclient.Client, caBundle []byte) error {
	for _, innerReconciler := range reconciler.Reconcilers {
		if err := innerReconciler.Reconcile(ctx, client, caBundle); err != nil {
			return err
		}
	}

	return nil
}

type ValidatingWebhookReconciler struct {
	Path              string
	name              string
	webhookName       string
	rules             []RuleWithOperations
	NameSpaceSelector v1.LabelSelector
}

type MutatingWebhookReconciler struct {
	Path              string
	name              string
	webhookName       string
	rules             []RuleWithOperations
	NameSpaceSelector v1.LabelSelector
}

//Reconcile MutatingWebhookConfiguration
func (reconciler *MutatingWebhookReconciler) Reconcile(ctx context.Context, client k8sclient.Client, caBundle []byte) error {
	var (
		sideEffects    = admissionregistrationv1.SideEffectClassNone
		port           = int32(servicePort)
		matchPolicy    = admissionregistrationv1.Exact
		ignorePolicy   = admissionregistrationv1.Ignore
		timeoutSeconds = int32(10)
	)

	namespace, err := utils.GetOperatorNamespace()
	if err != nil {
		return err
	}

	cr := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s", reconciler.name),
		},
	}

	klog.Infof("Creating/Updating MutatingWebhook %s", fmt.Sprintf("%s", reconciler.name))
	var webhokRules []admissionregistrationv1.RuleWithOperations
	for _, rule := range reconciler.rules {
		scope := admissionregistrationv1.AllScopes
		if rule.Scope != "" {
			scope = rule.Scope
		}
		webhokRules = append(webhokRules, admissionregistrationv1.RuleWithOperations{
			Operations: rule.Operations,
			Rule: admissionregistrationv1.Rule{
				APIGroups:   rule.APIGroups,
				APIVersions: rule.APIVersions,
				Resources:   rule.Resources,
				Scope:       &scope,
			}})
	}
	_, err = controllerutil.CreateOrUpdate(ctx, client, cr, func() error {
		cr.Webhooks = []admissionregistrationv1.MutatingWebhook{
			{
				Name:        fmt.Sprintf("%s", reconciler.webhookName),
				SideEffects: &sideEffects,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caBundle,
					Service: &admissionregistrationv1.ServiceReference{
						Namespace: namespace,
						Name:      operatorPodServiceName,
						Path:      &reconciler.Path,
						Port:      &port,
					},
				},
				Rules:                   webhokRules,
				MatchPolicy:             &matchPolicy,
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				FailurePolicy:           &ignorePolicy,
				TimeoutSeconds:          &timeoutSeconds,
			},
		}
		for index := range cr.Webhooks {
			cr.Webhooks[index].NamespaceSelector = &reconciler.NameSpaceSelector
		}
		return nil
	})
	if err != nil {
		klog.Error(err)
	}
	return err
}

//Reconcile ValidatingWebhookConfiguration
func (reconciler *ValidatingWebhookReconciler) Reconcile(ctx context.Context, client k8sclient.Client, caBundle []byte) error {
	var (
		sideEffects    = admissionregistrationv1.SideEffectClassNone
		port           = int32(servicePort)
		matchPolicy    = admissionregistrationv1.Exact
		failurePolicy  = admissionregistrationv1.Fail
		timeoutSeconds = int32(10)
	)

	namespace, err := utils.GetOperatorNamespace()
	if err != nil {
		return err
	}

	cr := &admissionregistrationv1.ValidatingWebhookConfiguration{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s", reconciler.name),
		},
	}

	var webhokRules []admissionregistrationv1.RuleWithOperations
	for _, rule := range reconciler.rules {
		scope := admissionregistrationv1.AllScopes
		if rule.Scope != "" {
			scope = rule.Scope
		}
		webhokRules = append(webhokRules, admissionregistrationv1.RuleWithOperations{
			Operations: rule.Operations,
			Rule: admissionregistrationv1.Rule{
				APIGroups:   rule.APIGroups,
				APIVersions: rule.APIVersions,
				Resources:   rule.Resources,
				Scope:       &scope,
			}})
	}

	klog.Infof("Creating/Updating ValidatingWebhook %s", fmt.Sprintf("%s", reconciler.name))
	_, err = controllerutil.CreateOrUpdate(ctx, client, cr, func() error {
		cr.Webhooks = []admissionregistrationv1.ValidatingWebhook{
			{
				Name:        fmt.Sprintf("%s", reconciler.webhookName),
				SideEffects: &sideEffects,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caBundle,
					Service: &admissionregistrationv1.ServiceReference{
						Namespace: namespace,
						Name:      operatorPodServiceName,
						Path:      &reconciler.Path,
						Port:      &port,
					},
				},
				Rules:                   webhokRules,
				MatchPolicy:             &matchPolicy,
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				FailurePolicy:           &failurePolicy,
				TimeoutSeconds:          &timeoutSeconds,
			},
		}
		for index := range cr.Webhooks {
			cr.Webhooks[index].NamespaceSelector = &reconciler.NameSpaceSelector
		}
		return nil
	})
	if err != nil {
		klog.Error(err)
	}
	return err
}

func (reconciler *ValidatingWebhookReconciler) SetName(name string) {
	reconciler.name = name
}

func (reconciler *MutatingWebhookReconciler) SetName(name string) {
	reconciler.name = name
}

func (reconciler *ValidatingWebhookReconciler) SetWebhookName(webhookName string) {
	reconciler.webhookName = webhookName
}

func (reconciler *MutatingWebhookReconciler) SetWebhookName(webhookName string) {
	reconciler.webhookName = webhookName
}

func (reconciler *ValidatingWebhookReconciler) SetRule(rules []RuleWithOperations) {
	reconciler.rules = rules
}

func (reconciler *MutatingWebhookReconciler) SetRule(rules []RuleWithOperations) {
	reconciler.rules = rules
}

func (reconciler *MutatingWebhookReconciler) SetNsSelector(selector v1.LabelSelector) {
	reconciler.NameSpaceSelector = selector
}

func (reconciler *ValidatingWebhookReconciler) SetNsSelector(selector v1.LabelSelector) {
	reconciler.NameSpaceSelector = selector
}
