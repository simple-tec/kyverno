package openapi

import (
	"encoding/json"
	"testing"

	v1 "github.com/kyverno/kyverno/pkg/api/kyverno/v1"
)

func Test_ValidateMutationPolicy(t *testing.T) {

	tcs := []struct {
		description string
		policy      []byte
		errMessage  string
	}{
		{
			description: "Policy with mutating imagePullPolicy Overlay",
			policy:      []byte(`{"apiVersion":"kyverno.io/v1","kind":"ClusterPolicy","metadata":{"name":"set-image-pull-policy-2"},"spec":{"rules":[{"name":"set-image-pull-policy-2","match":{"resources":{"kinds":["Pod"]}},"mutate":{"overlay":{"spec":{"containers":[{"(image)":"*","imagePullPolicy":"Always"}]}}}}]}}`),
		},
		{
			description: "Policy with mutating imagePullPolicy Overlay, field does not exist",
			policy:      []byte(`{"apiVersion":"kyverno.io/v1","kind":"ClusterPolicy","metadata":{"name":"set-image-pull-policy-2"},"spec":{"rules":[{"name":"set-image-pull-policy-2","match":{"resources":{"kinds":["Pod"]}},"mutate":{"overlay":{"spec":{"containers":[{"(image)":"*","nonExistantField":"Always"}]}}}}]}}`),
			errMessage:  `ValidationError(io.k8s.api.core.v1.Pod.spec.containers[0]): unknown field "nonExistantField" in io.k8s.api.core.v1.Container`,
		},
		{
			description: "Policy with mutating imagePullPolicy Overlay, type of value is different (does not throw error since all numbers are also strings according to swagger)",
			policy:      []byte(`{"apiVersion":"kyverno.io/v1","kind":"ClusterPolicy","metadata":{"name":"set-image-pull-policy-2"},"spec":{"rules":[{"name":"set-image-pull-policy-2","match":{"resources":{"kinds":["Pod"]}},"mutate":{"overlay":{"spec":{"containers":[{"(image)":"*","imagePullPolicy":80}]}}}}]}}`),
		},
		{
			description: "Policy with patches",
			policy:      []byte(`{"apiVersion":"kyverno.io/v1","kind":"ClusterPolicy","metadata":{"name":"policy-endpoints"},"spec":{"rules":[{"name":"pEP","match":{"resources":{"kinds":["Endpoints"],"selector":{"matchLabels":{"label":"test"}}}},"mutate":{"patches":[{"path":"/subsets/0/ports/0/port","op":"replace","value":9663},{"path":"/metadata/labels/isMutated","op":"add","value":"true"}]}}]}}`),
		},
		{
			description: "Policy with patches, value converted from number to string",
			policy:      []byte(`{"apiVersion":"kyverno.io/v1","kind":"ClusterPolicy","metadata":{"name":"policy-endpoints"},"spec":{"rules":[{"name":"pEP","match":{"resources":{"kinds":["Endpoints"],"selector":{"matchLabels":{"label":"test"}}}},"mutate":{"patches":[{"path":"/subsets/0/ports/0/port","op":"replace","value":"9663"},{"path":"/metadata/labels/isMutated","op":"add","value":"true"}]}}]}}`),
			errMessage:  `ValidationError(io.k8s.api.core.v1.Endpoints.subsets[0].ports[0].port): invalid type for io.k8s.api.core.v1.EndpointPort.port: got "string", expected "integer"`,
		},
		{
			description: "Policy where boolean is been converted to number",
			policy:      []byte(`{"apiVersion":"kyverno.io/v1","kind":"ClusterPolicy","metadata":{"name":"mutate-pod-disable-automoutingapicred"},"spec":{"rules":[{"name":"pod-disable-automoutingapicred","match":{"resources":{"kinds":["Pod"]}},"mutate":{"overlay":{"spec":{"(serviceAccountName)":"*","automountServiceAccountToken":80}}}}]}}`),
			errMessage:  `ValidationError(io.k8s.api.core.v1.Pod.spec.automountServiceAccountToken): invalid type for io.k8s.api.core.v1.PodSpec.automountServiceAccountToken: got "integer", expected "boolean"`,
		},
		{
			description: "Dealing with nested variables",
			policy:      []byte(`{"apiVersion":"kyverno.io/v1","kind":"ClusterPolicy","metadata":{"name":"add-ns-access-controls","annotations":{"policies.kyverno.io/category":"Workload Isolation","policies.kyverno.io/description":"Create roles and role bindings for a new namespace"}},"spec":{"background":false,"rules":[{"name":"add-sa-annotation","match":{"resources":{"kinds":["Namespace"]}},"mutate":{"overlay":{"metadata":{"annotations":{"nirmata.io/ns-creator":"{{serviceAccountName-{{something}}}}"}}}}},{"name":"generate-owner-role","match":{"resources":{"kinds":["Namespace"]}},"preconditions":[{"key":"{{request.userInfo.username}}","operator":"NotEqual","value":""},{"key":"{{serviceAccountName}}","operator":"NotEqual","value":""},{"key":"{{serviceAccountNamespace}}","operator":"NotEqual","value":""}],"generate":{"kind":"ClusterRole","name":"ns-owner-{{request.object.metadata.name{{something}}}}-{{request.userInfo.username}}","data":{"metadata":{"annotations":{"nirmata.io/ns-creator":"{{serviceAccountName}}"}},"rules":[{"apiGroups":[""],"resources":["namespaces"],"verbs":["delete"],"resourceNames":["{{request.object.metadata.name}}"]}]}}},{"name":"generate-owner-role-binding","match":{"resources":{"kinds":["Namespace"]}},"preconditions":[{"key":"{{request.userInfo.username}}","operator":"NotEqual","value":""},{"key":"{{serviceAccountName}}","operator":"NotEqual","value":""},{"key":"{{serviceAccountNamespace}}","operator":"NotEqual","value":""}],"generate":{"kind":"ClusterRoleBinding","name":"ns-owner-{{request.object.metadata.name}}-{{request.userInfo.username}}-binding","data":{"metadata":{"annotations":{"nirmata.io/ns-creator":"{{serviceAccountName}}"}},"roleRef":{"apiGroup":"rbac.authorization.k8s.io","kind":"ClusterRole","name":"ns-owner-{{request.object.metadata.name}}-{{request.userInfo.username}}"},"subjects":[{"kind":"ServiceAccount","name":"{{serviceAccountName}}","namespace":"{{serviceAccountNamespace}}"}]}}},{"name":"generate-admin-role-binding","match":{"resources":{"kinds":["Namespace"]}},"preconditions":[{"key":"{{request.userInfo.username}}","operator":"NotEqual","value":""},{"key":"{{serviceAccountName}}","operator":"NotEqual","value":""},{"key":"{{serviceAccountNamespace}}","operator":"NotEqual","value":""}],"generate":{"kind":"RoleBinding","name":"ns-admin-{{request.object.metadata.name}}-{{request.userInfo.username}}-binding","namespace":"{{request.object.metadata.name}}","data":{"metadata":{"annotations":{"nirmata.io/ns-creator":"{{serviceAccountName}}"}},"roleRef":{"apiGroup":"rbac.authorization.k8s.io","kind":"ClusterRole","name":"admin"},"subjects":[{"kind":"ServiceAccount","name":"{{serviceAccountName}}","namespace":"{{serviceAccountNamespace}}"}]}}}]}}`),
		},
	}

	o, _ := NewOpenAPIController()

	for i, tc := range tcs {
		policy := v1.ClusterPolicy{}
		_ = json.Unmarshal(tc.policy, &policy)

		var errMessage string
		err := o.ValidatePolicyMutation(policy)
		if err != nil {
			errMessage = err.Error()
		}

		if errMessage != tc.errMessage {
			t.Errorf("\nTestcase [%v] failed:\nExpected Error:  %v\nGot Error:  %v", i+1, tc.errMessage, errMessage)
		}
	}

}

func Test_addDefaultFieldsToSchema(t *testing.T) {
	addingDefaultFieldsToSchema([]byte(`null`))
	addingDefaultFieldsToSchema(nil)
}
