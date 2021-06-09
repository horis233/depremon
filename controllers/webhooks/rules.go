package webhooks

import admissionregistrationv1 "k8s.io/api/admissionregistration/v1"

// The `RuleWithOperations` and `Rule` types redefine the original ones from
// k8s.io/api/admissionregistration/v1 in order to allow to define methods
// to build the rule as a fluent interface.

type RuleWithOperations struct {
	Operations []admissionregistrationv1.OperationType
	Rule
}

type Rule struct {
	APIGroups   []string
	APIVersions []string
	Resources   []string
	Scope       admissionregistrationv1.ScopeType
}

func NewRule() RuleWithOperations {
	return RuleWithOperations{}
}

func (rule RuleWithOperations) OneResource(apiGroup, apiVersion, resource string) RuleWithOperations {
	rule.APIGroups = []string{apiGroup}
	rule.APIVersions = []string{apiVersion}
	rule.Resources = []string{resource}

	return rule
}

func (rule RuleWithOperations) NamespacedScope() RuleWithOperations {
	rule.Scope = admissionregistrationv1.NamespacedScope
	return rule
}

func (rule RuleWithOperations) ClusterScope() RuleWithOperations {
	rule.Scope = admissionregistrationv1.ClusterScope
	return rule
}

func (rule RuleWithOperations) AllScope() RuleWithOperations {
	rule.Scope = admissionregistrationv1.AllScopes
	return rule
}

func (rule RuleWithOperations) ForCreate() RuleWithOperations {
	rule.Operations = append(rule.Operations, admissionregistrationv1.Create)
	return rule
}

func (rule RuleWithOperations) ForUpdate() RuleWithOperations {
	rule.Operations = append(rule.Operations, admissionregistrationv1.Update)
	return rule
}

func (rule RuleWithOperations) ForDelete() RuleWithOperations {
	rule.Operations = append(rule.Operations, admissionregistrationv1.Delete)
	return rule
}

func (rule RuleWithOperations) ForAll() RuleWithOperations {
	rule.Operations = append(rule.Operations, admissionregistrationv1.OperationAll)
	return rule
}
