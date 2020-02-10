/*
Copyright 2019 Hidetake Iwata.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScheduledPodScalerSpec defines the desired state of ScheduledPodScaler
type ScheduledPodScalerSpec struct {
	ScaleTarget      ScaleTarget `json:"scaleTarget,omitempty"`
	ScaleRules       []ScaleRule `json:"schedule,omitempty"`
	DefaultScaleSpec ScaleSpec   `json:"default,omitempty"`
}

// ScaleTarget represents the resource to scale.
// For now only Deployment is supported.
type ScaleTarget struct {
	// +optional
	Selectors map[string]string `json:"selectors,omitempty"`
}

// ScaleRule represents a rule of scaling schedule.
type ScaleRule struct {
	ScaleSpec ScaleSpec `json:"spec,omitempty"`
	// Timezone, default to UTC.
	// +optional
	Timezone string `json:"timezone,omitempty"`
	// +optional
	Daily *DailyRule `json:"daily,omitempty"`
}

// DailyRule represents a rule to apply everyday.
type DailyRule struct {
	// Time format in 00:00:00.
	// If EndTime < StartTime, it treats the EndTime as the next day.
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
}

// ScaleSpec represents the desired state to scale the resource.
type ScaleSpec struct {
	Replicas int32 `json:"replicas,omitempty"`
}

// ScheduledPodScalerStatus defines the observed state of ScheduledPodScaler
type ScheduledPodScalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	NextReconcileTime string `json:"nextReconcileTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ScheduledPodScaler is the Schema for the scheduledpodscalers API
type ScheduledPodScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledPodScalerSpec   `json:"spec,omitempty"`
	Status ScheduledPodScalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScheduledPodScalerList contains a list of ScheduledPodScaler
type ScheduledPodScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScheduledPodScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScheduledPodScaler{}, &ScheduledPodScalerList{})
}
