package guid

import (
	"sync"

	"go.uber.org/atomic"
)

type segment struct {
	MaxId int64
	Step  int
	Value atomic.Int64
}

func (s *segment) IsNotEnough() bool {
	return s.MaxId-s.Value.Load() > int64(s.Step>>1)
}

func (s *segment) increAndGet() int64 {
	return s.Value.Add(1)
}

type segmentBuffer struct {
	mu sync.RWMutex

	bizTag    string
	segments  [2]*segment
	index     int
	running   atomic.Bool
	nextReady atomic.Bool
}

func NewSegmentBuffer(bizTag string) *segmentBuffer {
	return &segmentBuffer{
		bizTag:   bizTag,
		segments: [2]*segment{{}, {}},
	}
}

func (buffer *segmentBuffer) currentSegment() *segment {
	return buffer.segments[buffer.index]
}

func (buffer *segmentBuffer) getNextSegment() *segment {
	return buffer.segments[buffer.nextIndex()]
}

func (buffer *segmentBuffer) nextIndex() int {
	return (buffer.index + 1) & 1
}

func (buffer *segmentBuffer) switchIndex() {
	buffer.index = buffer.nextIndex()
}
