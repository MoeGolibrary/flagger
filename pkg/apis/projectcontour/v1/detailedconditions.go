// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +k8s:deepcopy-gen=package

// Package v1 is the v1 version of the API.
// +groupName=projectcontour.io
package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionStatus describes the state of a condition.
// +kubebuilder:validation:Enum=True;False;Unknown
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// Condition contains details for one aspect of the current state of this API Resource.
type Condition struct {
	// type of condition in CamelCase or in foo.example.com/CamelCase.
	// ---
	// Many .condition.type values are consistent across resources like Available, but because arbitrary conditions can be
	// useful (see .node.status.conditions), the ability to deconflict is important.
	// The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`
	// +kubebuilder:validation:MaxLength=316
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// status of the condition, one of True, False, Unknown.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`
	// observedGeneration represents the .metadata.generation that the condition was set based upon.
	// For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
	// with respect to the current state of the instance.
	// +optional
	// +kubebuilder:validation:Minimum=0
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,3,opt,name=observedGeneration"`
	// lastTransitionTime is the last time the condition transitioned from one status to another.
	// This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	LastTransitionTime metav1.Time `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// reason contains a programmatic identifier indicating the reason for the condition's last transition.
	// Producers of specific condition types may define expected values and meanings for this field,
	// and whether the values are considered a guaranteed API.
	// The value should be a CamelCase string.
	// This field may not be empty.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$`
	Reason string `json:"reason" protobuf:"bytes,5,opt,name=reason"`
	// message is a human readable message indicating details about the transition.
	// This may be an empty string.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=32768
	Message string `json:"message" protobuf:"bytes,6,opt,name=message"`
}

// SubCondition is a Condition-like type intended for use as a subcondition inside a DetailedCondition.
//
// It contains a subset of the Condition fields.
//
// It is intended for warnings and errors, so `type` names should use abnormal-true polarity,
// that is, they should be of the form "ErrorPresent: true".
//
// The expected lifecycle for these errors is that they should only be present when the error or warning is,
// and should be removed when they are not relevant.
type SubCondition struct {
	// Type of condition in `CamelCase` or in `foo.example.com/CamelCase`.
	//
	// This must be in abnormal-true polarity, that is, `ErrorFound` or `controller.io/ErrorFound`.
	//
	// The regex it matches is (dns1123SubdomainFmt/)?(qualifiedNameFmt)
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$`
	// +kubebuilder:validation:MaxLength=316
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`
	// Status of the condition, one of True, False, Unknown.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status"`
	// Reason contains a programmatic identifier indicating the reason for the condition's last transition.
	// Producers of specific condition types may define expected values and meanings for this field,
	// and whether the values are considered a guaranteed API.
	//
	// The value should be a CamelCase string.
	//
	// This field may not be empty.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$`
	Reason string `json:"reason" protobuf:"bytes,3,opt,name=reason"`
	// Message is a human readable message indicating details about the transition.
	//
	// This may be an empty string.
	//
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=32768
	Message string `json:"message" protobuf:"bytes,4,opt,name=message"`
}

// DetailedCondition is an extension of the normal Kubernetes conditions, with two extra
// fields to hold sub-conditions, which provide more detailed reasons for the state (True or False)
// of the condition.
//
// `errors` holds information about sub-conditions which are fatal to that condition and render its state False.
//
// `warnings` holds information about sub-conditions which are not fatal to that condition and do not force the state to be False.
//
// Remember that Conditions have a type, a status, and a reason.
//
// The type is the type of the condition, the most important one in this CRD set is `Valid`.
// `Valid` is a positive-polarity condition: when it is `status: true` there are no problems.
//
// In more detail, `status: true` means that the object is has been ingested into Contour with no errors.
// `warnings` may still be present, and will be indicated in the Reason field. There must be zero entries in the `errors`
// slice in this case.
//
// `Valid`, `status: false` means that the object has had one or more fatal errors during processing into Contour.
//
//	The details of the errors will be present under the `errors` field. There must be at least one error in the `errors`
//
// slice if `status` is `false`.
//
// For DetailedConditions of types other than `Valid`, the Condition must be in the negative polarity.
// When they have `status` `true`, there is an error. There must be at least one entry in the `errors` Subcondition slice.
// When they have `status` `false`, there are no serious errors, and there must be zero entries in the `errors` slice.
// In either case, there may be entries in the `warnings` slice.
//
// Regardless of the polarity, the `reason` and `message` fields must be updated with either the detail of the reason
// (if there is one and only one entry in total across both the `errors` and `warnings` slices), or
// `MultipleReasons` if there is more than one entry.
type DetailedCondition struct {
	Condition `json:",inline"`
	// Errors contains a slice of relevant error subconditions for this object.
	//
	// Subconditions are expected to appear when relevant (when there is a error), and disappear when not relevant.
	// An empty slice here indicates no errors.
	// +optional
	Errors []SubCondition `json:"errors,omitempty"`
	// Warnings contains a slice of relevant warning subconditions for this object.
	//
	// Subconditions are expected to appear when relevant (when there is a warning), and disappear when not relevant.
	// An empty slice here indicates no warnings.
	// +optional
	Warnings []SubCondition `json:"warnings,omitempty"`
}

const (
	// ValidConditionType describes an valid condition.
	ValidConditionType = "Valid"

	// ConditionTypeAuthError describes an error condition related to Auth.
	ConditionTypeAuthError = "AuthError"

	// ConditionTypeCORSError describes an error condition related to CORS.
	ConditionTypeCORSError = "CORSError"

	// ConditionTypeIncludeError describes an error condition with
	// inclusion of another HTTPProxy resource.
	ConditionTypeIncludeError = "IncludeError"

	// ConditionTypeOrphanedError describes an error condition
	// with an HTTPProxy resource which is not part of a delegation chain.
	ConditionTypeOrphanedError = "Orphaned"

	// ConditionTypePrefixReplaceError describes an error condition with
	// an HTTPProxy path prefix replacement issue.
	ConditionTypePrefixReplaceError = "PrefixReplaceError"

	// ConditionTypeRootNamespaceError describes an error condition
	// with an HTTPProxy resource created in non-root namespace.
	ConditionTypeRootNamespaceError = "RootNamespaceError"

	// ConditionTypeRouteError describes an error condition that
	// relates to Routes within an HTTPProxy.
	ConditionTypeRouteError = "RouteError"

	// ConditionTypeServiceError describes an error condition that
	// relates to a Service error within an HTTPProxy.
	ConditionTypeServiceError = "ServiceError"

	// ConditionTypeSpecError describes an error condition that
	// relates to the Spec of an HTTPProxy resource.
	ConditionTypeSpecError = "SpecError"

	// ConditionTypeTCPProxyIncludeError describes an error condition
	// with inclusion of another HTTPProxy TCP Proxy resource.
	ConditionTypeTCPProxyIncludeError = "TCPProxyIncludeError"

	// ConditionTypeTCPProxyError describes an error condition relating
	// to a TCP Proxy HTTPProxy resource.
	ConditionTypeTCPProxyError = "TCPProxyError"

	// ConditionTypeTLSError describes an error condition relating
	// to TLS configuration.
	ConditionTypeTLSError = "TLSError"

	// ConditionTypeVirtualHostError describes an error condition relating
	// to the VirtualHost configuration section of an HTTPProxy resource.
	ConditionTypeVirtualHostError = "VirtualHostError"
)
