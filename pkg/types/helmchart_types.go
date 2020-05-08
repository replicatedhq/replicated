package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Type=""
type MappedChartValue struct {
	Value string `json:"-"`

	valueType string `json:"-"`

	strValue   string  `json:"-"`
	boolValue  bool    `json:"-"`
	floatValue float64 `json:"-"`

	children map[string]*MappedChartValue `json:"-"`
	array    []*MappedChartValue          `json:"-"`
}

type ChartIdentifier struct {
	Name         string `json:"name",yaml:"name"`
	ChartVersion string `json:"chartVersion",yaml:"chartVersion"`
}

type OptionalValue struct {
	When string `json:"when"`

	Values map[string]MappedChartValue `json:"values,omitempty"`
}

// HelmChartSpec defines the desired state of HelmChartSpec
type HelmChartSpec struct {
	Chart          ChartIdentifier             `json:"chart"`
	Exclude        string                      `json:"exclude,omitempty"`
	Namespace      string                      `json:"namespace,omitempty"`
	Values         map[string]MappedChartValue `json:"values,omitempty"`
	OptionalValues []*OptionalValue            `json:"optionalValues,omitempty"`
	Builder        map[string]MappedChartValue `json:"builder,omitempty"`
}

// HelmChartStatus defines the observed state of HelmChart
type HelmChartStatus struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// HelmChart is the Schema for the helmchart API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type HelmChart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelmChartSpec   `json:"spec,omitempty"`
	Status HelmChartStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HelmChartList contains a list of HelmCharts
type HelmChartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HelmChart `json:"items"`
}

