package guidlog

import "log"

// GuidLog is common logger for guid
type GuidLog interface {
	Debugf(format string, fields ...interface{})
	Infof(format string, fields ...interface{})
	Warnf(format string, fields ...interface{})
	Errorf(format string, fields ...interface{})
}

type DefaultGuidLog struct {
}

func (l DefaultGuidLog) Debugf(format string, fields ...interface{}) {
	log.Printf(format, fields...)
}

func (l DefaultGuidLog) Infof(format string, fields ...interface{}) {
	log.Printf(format, fields...)
}
func (l DefaultGuidLog) Warnf(format string, fields ...interface{}) {
	log.Printf(format, fields...)
}

func (l DefaultGuidLog) Errorf(format string, fields ...interface{}) {
	log.Printf(format, fields...)
}
