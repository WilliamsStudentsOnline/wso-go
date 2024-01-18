package admin

import (
	"time"
)

/**
 * These types are copied from k8s.io/api/batch/v1/types.go
 *
 * They are used solely for swagger documentation generation for path `@Router /admin/jobs/{jobID}/status [get]`
 * They are not used in the code.
 *
 * They are added so swaggo does not need to parse external dependencies.
 */

// disable type unused warnings
var _ k8sJobStatus

type k8sJobConditionType string
type k8sConditionStatus string

type k8sUID string

// https://pkg.go.dev/k8s.io/apimachinery@v0.23.4/pkg/apis/meta/v1#Time
type k8sTime struct {
	time.Time `protobuf:"-"`
}

// https://pkg.go.dev/k8s.io/api@v0.23.4/batch/v1#JobCondition
type k8sJobCondition struct {
	// Type of job condition, Complete or Failed.
	Type k8sJobConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=JobConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status k8sConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Last time the condition was checked.
	// +optional
	LastProbeTime k8sTime `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime k8sTime `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// https://pkg.go.dev/k8s.io/api@v0.23.4/batch/v1#UncountedTerminatedPods
type k8sUncountedTerminatedPods struct {
	// Succeeded holds UIDs of succeeded Pods.
	// +listType=set
	// +optional
	Succeeded []k8sUID `json:"succeeded,omitempty" protobuf:"bytes,1,rep,name=succeeded,casttype=k8s.io/apimachinery/pkg/types.UID"`

	// Failed holds UIDs of failed Pods.
	// +listType=set
	// +optional
	Failed []k8sUID `json:"failed,omitempty" protobuf:"bytes,2,rep,name=failed,casttype=k8s.io/apimachinery/pkg/types.UID"`
}

// https://pkg.go.dev/k8s.io/api@v0.23.4/batch/v1#JobStatus
type k8sJobStatus struct {
	// The latest available observations of an object's current state. When a Job
	// fails, one of the conditions will have type "Failed" and status true. When
	// a Job is suspended, one of the conditions will have type "Suspended" and
	// status true; when the Job is resumed, the status of this condition will
	// become false. When a Job is completed, one of the conditions will have
	// type "Complete" and status true.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/jobs-run-to-completion/
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=atomic
	Conditions []k8sJobCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	// Represents time when the job controller started processing a job. When a
	// Job is created in the suspended state, this field is not set until the
	// first time it is resumed. This field is reset every time a Job is resumed
	// from suspension. It is represented in RFC3339 form and is in UTC.
	// +optional
	StartTime *k8sTime `json:"startTime,omitempty" protobuf:"bytes,2,opt,name=startTime"`

	// Represents time when the job was completed. It is not guaranteed to
	// be set in happens-before order across separate operations.
	// It is represented in RFC3339 form and is in UTC.
	// The completion time is only set when the job finishes successfully.
	// +optional
	CompletionTime *k8sTime `json:"completionTime,omitempty" protobuf:"bytes,3,opt,name=completionTime"`

	// The number of pending and running pods.
	// +optional
	Active int32 `json:"active,omitempty" protobuf:"varint,4,opt,name=active"`

	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32 `json:"succeeded,omitempty" protobuf:"varint,5,opt,name=succeeded"`

	// The number of pods which reached phase Failed.
	// +optional
	Failed int32 `json:"failed,omitempty" protobuf:"varint,6,opt,name=failed"`

	// CompletedIndexes holds the completed indexes when .spec.completionMode =
	// "Indexed" in a text format. The indexes are represented as decimal integers
	// separated by commas. The numbers are listed in increasing order. Three or
	// more consecutive numbers are compressed and represented by the first and
	// last element of the series, separated by a hyphen.
	// For example, if the completed indexes are 1, 3, 4, 5 and 7, they are
	// represented as "1,3-5,7".
	// +optional
	CompletedIndexes string `json:"completedIndexes,omitempty" protobuf:"bytes,7,opt,name=completedIndexes"`

	// UncountedTerminatedPods holds the UIDs of Pods that have terminated but
	// the job controller hasn't yet accounted for in the status counters.
	//
	// The job controller creates pods with a finalizer. When a pod terminates
	// (succeeded or failed), the controller does three steps to account for it
	// in the job status:
	// (1) Add the pod UID to the arrays in this field.
	// (2) Remove the pod finalizer.
	// (3) Remove the pod UID from the arrays while increasing the corresponding
	//     counter.
	//
	// This field is beta-level. The job controller only makes use of this field
	// when the feature gate JobTrackingWithFinalizers is enabled (enabled
	// by default).
	// Old jobs might not be tracked using this field, in which case the field
	// remains null.
	// +optional
	UncountedTerminatedPods *k8sUncountedTerminatedPods `json:"uncountedTerminatedPods,omitempty" protobuf:"bytes,8,opt,name=uncountedTerminatedPods"`

	// The number of pods which have a Ready condition.
	//
	// This field is alpha-level. The job controller populates the field when
	// the feature gate JobReadyPods is enabled (disabled by default).
	// +optional
	Ready *int32 `json:"ready,omitempty" protobuf:"varint,9,opt,name=ready"`
}
