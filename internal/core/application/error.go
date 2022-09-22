package application

import "errors"

var (
	ErrInvalidTimeFrame = errors.New("timeFrame must be smaller than timePeriod")
)
