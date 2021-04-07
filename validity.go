package sari

import (
	"time"

	"github.com/asaskevich/govalidator"
)

type ValidityPeriod struct {
	NotValidBefore time.Time `yaml:"not_valid_before"`
	NotValidAfter  time.Time `yaml:"not_valid_after"`
}

func init() {
	govalidator.CustomTypeTagMap.Set("period", func(field interface{}, context interface{}) bool {
		switch v := field.(type) {
		case ValidityPeriod:
			return v.IsWellDefined()
		default:
			return false
		}
	})
}

func (v ValidityPeriod) IsWellDefined() bool {
	return v.NotValidBefore.IsZero() || v.NotValidAfter.IsZero() || v.NotValidBefore.Before(v.NotValidAfter)
}

type ValidityChecker interface {
	Check(period ValidityPeriod) bool
}

type TrackingValidityChecker struct {
	timeRef        time.Time
	nextTransition time.Time
}

func NewTrackingValidityChecker(timeRef time.Time) *TrackingValidityChecker {
	return &TrackingValidityChecker{timeRef: timeRef}
}

func (c *TrackingValidityChecker) Check(period ValidityPeriod) bool {
	if !period.NotValidBefore.IsZero() && c.timeRef.Before(period.NotValidBefore) {
		c.updateNextTransition(period.NotValidBefore)
		return false
	}
	if !period.NotValidAfter.IsZero() {
		if c.timeRef.After(period.NotValidAfter) {
			return false
		}
		c.updateNextTransition(period.NotValidAfter)
	}
	return true
}

func (c TrackingValidityChecker) GetNextTransition() time.Time {
	return c.nextTransition
}

func (c *TrackingValidityChecker) updateNextTransition(nextTransition time.Time) {
	if c.nextTransition.IsZero() || nextTransition.Before(c.nextTransition) {
		c.nextTransition = nextTransition
	}
}
