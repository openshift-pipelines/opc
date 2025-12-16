package opc

import (
	"context"
	"testing"

	"github.com/tektoncd/cli/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetRedHatOpenShiftPipelinesVersion(t *testing.T) {
	testParams := []struct {
		name      string
		namespace string
		configMap *corev1.ConfigMap
		want      string
	}{
		{
			name:      "get version from product field",
			namespace: "openshift-pipelines",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tekton-operator-info",
				},
				Data: map[string]string{
					"product": "1.8.0",
				},
			},
			want: "1.8.0",
		},
		{
			name:      "get version from version field with embedded product string",
			namespace: "openshift-pipelines",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tekton-operator-info",
				},
				Data: map[string]string{
					"version": "0.56.0 (Red Hat OpenShift Pipelines 1.8.1)",
				},
			},
			want: "1.8.1",
		},
		{
			name:      "get version from version field with only parentheses",
			namespace: "openshift-pipelines",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tekton-operator-info",
				},
				Data: map[string]string{
					"version": "0.56.0 (1.8.2)",
				},
			},
			want: "1.8.2",
		},
		{
			name:      "get version from rhProduct field as fallback",
			namespace: "openshift-pipelines",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tekton-operator-info",
				},
				Data: map[string]string{
					"rhProduct": "1.8.3",
				},
			},
			want: "1.8.3",
		},
		{
			name:      "no relevant version fields return empty",
			namespace: "openshift-pipelines",
			configMap: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tekton-operator-info",
				},
				Data: map[string]string{
					"unrelated": "data",
				},
			},
			want: "",
		},
	}

	for _, tp := range testParams {
		t.Run(tp.name, func(t *testing.T) {
			cs, _ := test.SeedV1beta1TestData(t, test.Data{})
			p := &test.Params{Kube: cs.Kube}
			cls, err := p.Clients()
			if err != nil {
				t.Fatalf("failed to get client: %v", err)
			}

			if _, err := cls.Kube.CoreV1().ConfigMaps(tp.namespace).Create(context.Background(), tp.configMap, metav1.CreateOptions{}); err != nil {
				t.Fatalf("failed to create configmap: %v", err)
			}

			got, _ := GetRedHatOpenShiftPipelinesVersion(cls, tp.namespace)
			if got != tp.want {
				t.Errorf("unexpected version, got %q, want %q", got, tp.want)
			}
		})
	}
}

func TestGetRedHatOpenShiftPipelinesVersion_ConfigMapNotFound(t *testing.T) {
	cs, _ := test.SeedV1beta1TestData(t, test.Data{})
	p := &test.Params{Kube: cs.Kube}
	cls, err := p.Clients()
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}

	got, _ := GetRedHatOpenShiftPipelinesVersion(cls, "openshift-pipelines")
	if got != "" {
		t.Errorf("expected empty string when ConfigMap not found, got %q", got)
	}
}
