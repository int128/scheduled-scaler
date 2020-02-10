package scheduledpodscaler

import (
	"time"

	"github.com/int128/scheduled-scaler/pkg/domain/schedule"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScheduledPodScaler struct {
	TypeMeta   metav1.TypeMeta
	ObjectMeta metav1.ObjectMeta

	Spec   Spec
	Status Status
}

type Spec struct {
	ScaleTarget      ScaleTarget
	ScaleRules       []ScaleRule
	DefaultScaleSpec ScaleSpec
}

// ComputeDesiredScaleSpec returns the ScaleSpec corresponding to the current time.
// This finds the active ScaleRule in order.
func (s *Spec) ComputeDesiredScaleSpec(now time.Time) ScaleSpec {
	for _, rule := range s.ScaleRules {
		if rule.IsActive(now) {
			return rule.ScaleSpec
		}
	}
	return s.DefaultScaleSpec
}

// FindNextReconcileTime returns the next time to reconcile.
// This finds the earliest ScaleRule in order.
func (s *Spec) FindNextReconcileTime(now time.Time) (earliest time.Time) {
	for _, rule := range s.ScaleRules {
		edge := rule.NextEdge(now)
		if earliest.IsZero() || edge.Before(earliest) {
			earliest = edge
		}
	}
	return
}

type ScaleTarget struct {
	Selectors map[string]string
}

type ScaleRule struct {
	Range     schedule.Range
	Timezone  *time.Location // must be non-nil
	ScaleSpec ScaleSpec
}

func (r *ScaleRule) IsActive(now time.Time) bool {
	return r.Range.IsActive(now.In(r.Timezone))
}

func (r *ScaleRule) NextEdge(now time.Time) time.Time {
	return r.Range.NextEdge(now.In(r.Timezone))
}

type ScaleSpec struct {
	Replicas int32
}

type Status struct {
	NextReconcileTime time.Time
}
