/*
Copyright 2021 The Crossplane Authors.

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

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ActivityParameters defines the desired state of Activity
type ActivityParameters struct {
	// Region is which region the Activity will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// The name of the activity to create. This name must be unique for your Amazon
	// Web Services account and region for 90 days. For more information, see Limits
	// Related to State Machine Executions (https://docs.aws.amazon.com/step-functions/latest/dg/limits.html#service-limits-state-machine-executions)
	// in the Step Functions Developer Guide.
	//
	// A name must not contain:
	//
	//    * white space
	//
	//    * brackets < > { } [ ]
	//
	//    * wildcard characters ? *
	//
	//    * special characters " # % \ ^ | ~ ` $ & , ; : /
	//
	//    * control characters (U+0000-001F, U+007F-009F)
	//
	// To enable logging with CloudWatch Logs, the name should only contain 0-9,
	// A-Z, a-z, - and _.
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// The list of tags to add to a resource.
	//
	// An array of key-value pairs. For more information, see Using Cost Allocation
	// Tags (https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/cost-alloc-tags.html)
	// in the Amazon Web Services Billing and Cost Management User Guide, and Controlling
	// Access Using IAM Tags (https://docs.aws.amazon.com/IAM/latest/UserGuide/access_iam-tags.html).
	//
	// Tags may only contain Unicode letters, digits, white space, or these symbols:
	// _ . : / = + - @.
	Tags                     []*Tag `json:"tags,omitempty"`
	CustomActivityParameters `json:",inline"`
}

// ActivitySpec defines the desired state of Activity
type ActivitySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ActivityParameters `json:"forProvider"`
}

// ActivityObservation defines the observed state of Activity
type ActivityObservation struct {
	// The Amazon Resource Name (ARN) that identifies the created activity.
	ActivityARN *string `json:"activityARN,omitempty"`
	// The date the activity is created.
	CreationDate *metav1.Time `json:"creationDate,omitempty"`
}

// ActivityStatus defines the observed state of Activity.
type ActivityStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ActivityObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Activity is the Schema for the Activities API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Activity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ActivitySpec   `json:"spec"`
	Status            ActivityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ActivityList contains a list of Activities
type ActivityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Activity `json:"items"`
}

// Repository type metadata.
var (
	ActivityKind             = "Activity"
	ActivityGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ActivityKind}.String()
	ActivityKindAPIVersion   = ActivityKind + "." + GroupVersion.String()
	ActivityGroupVersionKind = GroupVersion.WithKind(ActivityKind)
)

func init() {
	SchemeBuilder.Register(&Activity{}, &ActivityList{})
}
