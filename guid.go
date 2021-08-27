package guid

import (
	"context"
	"sync"
	"time"

	"github.com/ziidu/guid/dao"
	"github.com/ziidu/guid/guidlog"
	"github.com/ziidu/guid/model"
	"github.com/ziidu/guid/singleflight"
	"go.uber.org/atomic"
)

// Getter get a unique id (for a key)
type Getter interface {
	// Get returns reponse
	Get(bizTag string) model.Response
}

var _ (Getter) = (*SegmentUID)(nil)

// SegmentUID generate unique id by db segment
type SegmentUID struct {
	cache cache
	log   guidlog.GuidLog
	dao   dao.GuidAllocDao
	group singleflight.Group
}

// NewSegmentUID create segmentUID a instance, and
// load all segmentMetadata to cache
func NewSegmentUID(dao dao.GuidAllocDao) *SegmentUID {
	segmentUID := &SegmentUID{log: guidlog.DefaultGuidLog{}, dao: dao}
	if err := segmentUID.loadAllLeafFromDb(); err != nil {
		panic(err)
	}
	return segmentUID
}

func (s *SegmentUID) UseLogger(logger guidlog.GuidLog) {
	s.log = logger
}

// load load segmentMetadata from db.
// when mutil goroutine call this method, only one of those is really executed,
// and other goroutine will wait the result and return directly
func (s *SegmentUID) load(bizTag string) (*segmentBuffer, error) {
	val, err := s.group.Do(bizTag, func() (interface{}, error) {
		segmentMetadata, err := s.dao.GetById(bizTag)
		if err != nil {
			return nil, err
		}
		return s.buildSegmentBuffer(bizTag, segmentMetadata), nil
	})
	return val.(*segmentBuffer), err
}

func (s *SegmentUID) Get(bizTag string) model.Response {
	var (
		buffer *segmentBuffer
		hit    bool
	)
	buffer, hit = s.cache.get(bizTag)
	if !hit {
		loadBuffer, err := s.load(bizTag)
		if err != nil {
			return model.ResponseErr(model.NotFoundErrCode)
		}
		buffer = loadBuffer
	}
LOOP:
	buffer.mu.RLock()
	segment := buffer.currentSegment()
	if !buffer.nextReady.Load() && segment.IsNotEnough() && buffer.running.CAS(false, true) {
		go s.preparedNextSegment(buffer)
	}
	value := segment.increAndGet()
	if value <= segment.MaxId {
		buffer.mu.RUnlock()
		return model.ResponseOk(value)
	}
	buffer.mu.RUnlock()
	waitAndSleep(buffer)
	buffer.mu.Lock()
	segment = buffer.currentSegment()
	value = segment.increAndGet()
	if value <= segment.MaxId {
		buffer.mu.Unlock()
		return model.ResponseOk(value)
	}
	if buffer.nextReady.Load() {
		buffer.switchIndex()
		buffer.nextReady.Store(false)
	}
	buffer.mu.Unlock()
	goto LOOP
}

func (s *SegmentUID) buildSegmentBuffer(bizTag string, sm model.SegmentMetadata) *segmentBuffer {
	buffer := NewSegmentBuffer(sm.BizTag)
	segment := buffer.currentSegment()
	segment.MaxId = sm.MaxId
	segment.Step = sm.Step
	segment.Value = *atomic.NewInt64(sm.MaxId - int64(sm.Step))
	s.cache.put(sm.BizTag, buffer)
	return buffer
}

func (s *SegmentUID) loadAllLeafFromDb() error {
	segmentsMetadatas, err := s.dao.GetAllSegmentMetadatas()
	if err != nil {
		return err
	}
	for _, sm := range segmentsMetadatas {
		updatedSm, err := s.dao.GetAndUpdateMaxId(sm.BizTag)
		if err != nil {
			return err
		}
		s.buildSegmentBuffer(sm.BizTag, updatedSm)
	}
	return nil
}

func (s *SegmentUID) preparedNextSegment(buffer *segmentBuffer) {
	defer func() {
		if err := recover(); err != nil {
			s.log.Errorf("prepared next segment panic: %v", err)
		}
		buffer.running.Store(false)
	}()
	var count = 0
	nextSegment := buffer.getNextSegment()
RETRY:
	metadata, err := s.dao.UpdateMaxIdAndGet(buffer.bizTag)
	// when some error occur, retry 3 time
	if err != nil {
		if count < 3 {
			count++
			goto RETRY
		}
		s.log.Errorf("dao UpdateMaxIdAndGet %s", err)
		return
	}
	nextSegment.MaxId = metadata.MaxId
	nextSegment.Step = metadata.Step
	nextSegment.Value = *atomic.NewInt64(metadata.MaxId - int64(metadata.Step))
	buffer.nextReady.Store(true)
}

func waitAndSleep(buffer *segmentBuffer) {
	for buffer.running.Load() {
		time.Sleep(10 * time.Millisecond)
	}
}

var _ (Getter) = (*SnowflakeUID)(nil)

const (
	timeBits    = 30
	workBits    = 8
	sequenceBit = 25

	workShift = sequenceBit
	timeShift = workBits + sequenceBit

	sequenceMask = 1<<sequenceBit - 1
)

type SnowflakeUID struct {
	sequence      int
	workId        int
	lastTimestamp int64
	epoch         int64
	lock          sync.Mutex
}

func NewSnowflakeUID(holder IWorkIDHolder) *SnowflakeUID {
	var err error
	snow := &SnowflakeUID{epoch: time.Date(2021, 7, 8, 0, 0, 0, 0, time.Local).Unix()}
	if snow.workId, err = holder.WorkId(context.Background()); err != nil {
		panic(err)
	}
	return snow
}

func (snow *SnowflakeUID) Get(bizTag string) model.Response {
	snow.lock.Lock()
	defer snow.lock.Unlock()
	var (
		timestamp = time.Now().Unix()
		id        int64
	)
	if timestamp < snow.lastTimestamp {
		return model.ResponseErr(model.TimeBackErrCode)
	}
	if timestamp == snow.lastTimestamp {
		snow.sequence = (snow.sequence + 1) & sequenceMask
		if snow.sequence == 0 {
			timestamp = waitNextSecond(snow.lastTimestamp)
		}
	} else {
		snow.sequence = 0
	}
	snow.lastTimestamp = timestamp
	id = (timestamp-snow.epoch)<<timeShift | int64(snow.workId)<<workShift | int64(snow.sequence)
	return model.ResponseOk(id)
}

func waitNextSecond(lastTimestamp int64) int64 {
	timestamp := time.Now().Unix()
	for timestamp < lastTimestamp {
		time.Sleep(time.Second * (time.Duration(lastTimestamp - timestamp)))
		timestamp = time.Now().Unix()
	}
	return timestamp
}
