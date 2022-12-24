package rkops

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rookie-ninja/rk-entry/v2/cursor"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"strings"
	"time"
)

func NewPointer(payload *rkcursor.CursorPayload) rkcursor.Pointer {
	return &Pointer{
		start:     payload.StartTime,
		parent:    payload.ParentFunc,
		operation: payload.Operation,
		event:     payload.Event,
		logger:    payload.Logger,
		entryName: payload.EntryName,
		entryType: payload.EntryType,
	}
}

type Pointer struct {
	start     time.Time
	parent    string
	operation string
	err       error
	event     rkquery.Event
	logger    *zap.Logger
	entryName string
	entryType string
}

func (c *Pointer) PrintError(err error) {
	if err == nil || c.logger == nil {
		return
	}

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	var builder bytes.Buffer
	builder.WriteString(err.Error())

	conv, ok := err.(stackTracer)

	if ok {
		builder.WriteString(fmt.Sprintf("\nStackTrace:"))

		st := conv.StackTrace()

		ss := errors.StackTrace{}
		for i := range st {
			frame := st[i]
			str := fmt.Sprintf("'%+s", frame)
			if strings.Contains(str, "@") {
				break
			}

			ss = append(ss, frame)
		}
		builder.WriteString(fmt.Sprintf("%+v", ss))
		builder.WriteString("\n")
	}

	c.logger.WithOptions(zap.AddCallerSkip(1)).Error(builder.String())
}

func (c *Pointer) ObserveError(err error) error {
	if err == nil {
		return nil
	}

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	_, ok := err.(stackTracer)

	if !ok {
		err = errors.WithStack(err)
	}

	if c.event != nil {
		c.event.IncCounter(strings.Join([]string{c.operation, "ERROR"}, "."), 1)
	}

	return err
}

func (c *Pointer) Release() {
	elapsedNano := time.Now().Sub(c.start).Nanoseconds()

	success := true
	if c.err != nil {
		success = false
	}

	observer, _ := rkcursor.SummaryVec().GetMetricWithLabelValues(
		rkcursor.PromLabels().GetValues(c.parent, c.operation, c.entryName, c.entryType, success)...)
	if observer == nil {
		return
	}

	if c.event != nil {
		c.event.EndTimer(c.operation)
	}

	observer.Observe(float64(elapsedNano))
}
