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

func (s *Spec) ComputeDesiredScaleSpec(now time.Time) ScaleSpec {
	for _, rule := range s.ScaleRules {
		if rule.Range.IsActive(now) {
			return rule.ScaleSpec
		}
	}
	return s.DefaultScaleSpec
}

func (s *Spec) FindNextReconcileTime(now time.Time) (earliest time.Time) {
	for _, rule := range s.ScaleRules {
		edge := rule.Range.NextEdge(now)
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
	ScaleSpec ScaleSpec
}

type ScaleSpec struct {
	Replicas int32
}

type Status struct {
	NextReconcileTime time.Time
}
