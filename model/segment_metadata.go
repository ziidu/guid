package model

import "time"

// SegmentMetadata is database table that preserve segment information
type SegmentMetadata struct {
	BizTag      string
	MaxId       int64
	Step        int
	Description string
	UpdateTime  time.Time
}
