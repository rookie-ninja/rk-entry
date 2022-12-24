package rkcursor

import (
	"bytes"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"runtime"
	"strings"
	"sync"
	"time"
)

const metricsKey = "elapsedNano"

var (
	summaryVec *prometheus.SummaryVec
	logger     *zap.Logger
	label      *promLabel
)

func init() {
	// 1: init labels
	label = &promLabel{
		keys: []string{
			"entryName",
			"entryType",
			"domain",
			"instance",
			"parent",
			"operation",
			"status",
		},
		values: []string{
			"",
			"",
			rkmid.Domain.String,
			rkmid.LocalIp.String,
		},
	}

	// 2: init summary vec and register to default registerer
	opts := prometheus.SummaryOpts{
		Namespace:  "rk",
		Subsystem:  "cursor",
		Name:       metricsKey,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001},
		Help:       fmt.Sprintf("Summary of cursor with labels:%s", label.keys),
	}
	summaryVec = prometheus.NewSummaryVec(opts, label.keys)
	prometheus.DefaultRegisterer.Register(summaryVec)

	// 3: init logger
	logger = rkentry.NewLoggerEntryStdout().Logger
}

// ************* Global *************

func OverrideEntryNameAndType(entryName, entryType string) {
	label.mutex.Lock()
	defer label.mutex.Unlock()

	label.values[0] = entryName
	label.values[1] = entryType
}

func OverrideLogger(l *zap.Logger) {
	if l != nil {
		logger = l
	}
}

func SummaryVec() *prometheus.SummaryVec {
	return summaryVec
}

func PromLabels() *promLabel {
	return label
}

func Click() *pointer {
	return &pointer{
		start:     time.Now(),
		parent:    parentName(),
		operation: operationName(),
	}
}

func Error(err error) {
	if err == nil {
		return
	}

	stack := stacks()

	var builder bytes.Buffer

	// print error message
	builder.WriteString(fmt.Sprintf("%s\n", err.Error()))
	// print stack function
	for i := range stack {
		pc := stack[i] - 1
		builder.WriteString(fmt.Sprintf("%d)\t%s\n", i, fileline(pc)))
	}

	logger.WithOptions(zap.AddCallerSkip(1)).Error(builder.String())
}

func AddField(key, val string) {
	logger = logger.With(zap.String(key, val))
}

// ************* Instance *************

type Option func(c *Cursor)

func WithEntryNameAndType(entryName, entryType string) Option {
	return func(c *Cursor) {
		c.EntryName = entryName
		c.EntryType = entryType
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(c *Cursor) {
		if l != nil {
			c.Logger = l
		}
	}
}

func WithEvent(e rkquery.Event) Option {
	return func(c *Cursor) {
		if e != nil {
			c.Event = e
		}
	}
}

func NewCursor(opts ...Option) *Cursor {
	c := &Cursor{
		Logger:    rkentry.LoggerEntryStdout.Logger,
		Event:     rkentry.EventEntryNoop.CreateEventNoop(),
		EntryName: "",
		EntryType: "",
		Now:       time.Now(),
		Creator:   rkPointerGenerator,
	}

	for i := range opts {
		opts[i](c)
	}

	return c
}

type CursorPayload struct {
	Logger    *zap.Logger
	Event     rkquery.Event
	StartTime time.Time

	EntryName  string
	EntryType  string
	Operation  string
	ParentFunc string
}

type PointerCreator func(payload *CursorPayload) Pointer

type Cursor struct {
	Logger *zap.Logger
	Event  rkquery.Event
	Now    time.Time

	EntryName string
	EntryType string

	Creator PointerCreator
}

func (c *Cursor) Click() Pointer {
	operation := operationName()

	if c.Event != nil {
		c.Event.StartTimer(operation)
	}

	payload := &CursorPayload{
		Logger:     c.Logger,
		Event:      c.Event,
		StartTime:  time.Now(),
		EntryName:  c.EntryName,
		EntryType:  c.EntryType,
		Operation:  operation,
		ParentFunc: parentName(),
	}

	return c.Creator(payload)
}

func (c *Cursor) Error(err error) {
	if err == nil {
		return
	}

	stack := stacks()

	var builder bytes.Buffer

	// print error message
	builder.WriteString(fmt.Sprintf("%s\n", err.Error()))
	// print stack function
	for i := range stack {
		pc := stack[i] - 1
		builder.WriteString(fmt.Sprintf("%d)\t%s\n", i, fileline(pc)))
	}

	c.Logger.WithOptions(zap.AddCallerSkip(1)).Error(builder.String())
}

func (c *Cursor) AddField(key, val string) {
	c.Logger = c.Logger.With(zap.String(key, val))
	c.Event.AddPair(key, val)
}

// ************* Prometheus labels *************

type promLabel struct {
	mutex  sync.Mutex
	keys   []string
	values []string
}

func (l *promLabel) GetValues(parent, op, entryName, entryType string, success bool) []string {
	label.mutex.Lock()
	defer label.mutex.Unlock()

	status := "OK"
	if !success {
		status = "ERROR"
	}

	res := append(l.values, parent, op, status)
	res[0] = entryName
	res[1] = entryType

	return res
}

// ************* Cursor *************

type Pointer interface {
	PrintError(err error)

	ObserveError(err error) error

	Release()
}

// ************* Default pointer *************

func rkPointerGenerator(payload *CursorPayload) Pointer {
	return &pointer{
		start:     payload.StartTime,
		parent:    payload.ParentFunc,
		operation: payload.Operation,
		event:     payload.Event,
		logger:    payload.Logger,
		entryName: payload.EntryName,
		entryType: payload.EntryType,
	}
}

type pointer struct {
	start     time.Time
	parent    string
	operation string
	err       error
	event     rkquery.Event
	logger    *zap.Logger
	entryName string
	entryType string
}

func (c *pointer) PrintError(err error) {
	stack := stacks()

	var builder bytes.Buffer

	// print error message
	builder.WriteString(fmt.Sprintf("%s\n", err.Error()))
	// print stack function
	for i := range stack {
		pc := stack[i] - 1
		builder.WriteString(fmt.Sprintf("%d)\t%s\n", i, fileline(pc)))
	}

	if c.logger != nil {
		c.logger.WithOptions(zap.AddCallerSkip(1)).Error(builder.String())
	} else {
		logger.WithOptions(zap.AddCallerSkip(1)).Error(builder.String())
	}
}

func (c *pointer) ObserveError(err error) error {
	if err == nil {
		return nil
	}

	c.err = err

	if c.event != nil {
		c.event.IncCounter(strings.Join([]string{c.operation, "ERROR"}, "."), 1)
	}

	return err
}

func (c *pointer) Release() {
	elapsedNano := time.Now().Sub(c.start).Nanoseconds()

	success := true
	if c.err != nil {
		success = false
	}

	observer, _ := summaryVec.GetMetricWithLabelValues(
		label.GetValues(c.parent, c.operation, c.entryName, c.entryType, success)...)
	if observer == nil {
		return
	}

	if c.event != nil {
		c.event.EndTimer(c.operation)
	}

	observer.Observe(float64(elapsedNano))
}

// ************* helper functions *************

func operationName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}

	fName := runtime.FuncForPC(pc).Name()
	// 1: try to check whether it is nested, trim prefix of file path
	fName = fName[strings.LastIndex(fName, "/")+1:]
	fName = strings.ReplaceAll(fName, "(", "")
	fName = strings.ReplaceAll(fName, ")", "")
	fName = strings.ReplaceAll(fName, "*", "")

	return fName
}

func parentName() string {
	pc, file, _, ok := runtime.Caller(3)
	if !ok {
		return "-"
	}

	fName := runtime.FuncForPC(pc).Name()
	if strings.Contains(file, "@") {
		return "-"
	}

	// 1: try to check whether it is nested, trim prefix of file path
	fName = fName[strings.LastIndex(fName, "/")+1:]
	fName = strings.ReplaceAll(fName, "(", "")
	fName = strings.ReplaceAll(fName, ")", "")
	fName = strings.ReplaceAll(fName, "*", "")

	return fName
}

func stacks() []uintptr {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])

	index := n
	for i := range pcs[:n] {
		pc := pcs[i]
		if strings.Contains(file(pc), "@") {
			index = i
			break
		}
	}

	return pcs[:index]
}

func file(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	file, _ := fn.FileLine(pc)
	return file
}

func fileline(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	file, line := fn.FileLine(pc)
	return fmt.Sprintf("%s\t%s:%d", fn.Name(), file, line)
}
