package workbench

import (
	"errors"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
)

// Step is authored intent, never a result, receipt or inferred shared population.
type StepScope struct {
	Environment string `json:"environment"`
	FromMS      int64  `json:"fromMs"`
	ToMS        int64  `json:"toMs"`
	FromNS      string `json:"fromNS,omitempty"`
	ToNS        string `json:"toNS,omitempty"`
}
type Step struct {
	Version     int                      `json:"version"`
	Kind        string                   `json:"kind"`
	Scope       StepScope                `json:"scope"`
	View        string                   `json:"view"`
	Record      *exploration.Query       `json:"record,omitempty"`
	Measurement *measurements.Definition `json:"measurement,omitempty"`
}

func (s Step) Validate() error {
	if s.Version != 1 {
		return errors.New("unsupported question step version")
	}
	switch s.Kind {
	case "record":
		if s.Scope.FromNS != "" || s.Scope.ToNS != "" || s.Record == nil || s.Measurement != nil || (s.View != "Records" && s.View != "Distribution") {
			return errors.New("record step requires one record question and supported view")
		}
		q := s.Record
		if q.Environment != s.Scope.Environment || q.FromMS != s.Scope.FromMS || q.ToMS != s.Scope.ToMS {
			return errors.New("record question differs from step scope")
		}
		return q.Validate()
	case "measurement":
		if s.Measurement == nil || s.Record != nil || s.View != "Overview" || s.Measurement.Environment != s.Scope.Environment {
			return errors.New("measurement step requires one definition matching scope")
		}
		return (measurements.Question{Definition: *s.Measurement, FromMS: s.Scope.FromMS, ToMS: s.Scope.ToMS, FromNS: s.Scope.FromNS, ToNS: s.Scope.ToNS}).Validate()
	default:
		return errors.New("unsupported question step kind")
	}
}
