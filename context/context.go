package rkctx

//import (
//	"github.com/rookie-ninja/rk-common/common"
//	"github.com/rookie-ninja/rk-entry/entry"
//	"github.com/rookie-ninja/rk-query"
//	"go.uber.org/zap"
//	"go.uber.org/zap/zapcore"
//)
//
//type EventContext struct {
//	EventId string `json:"eventId" yaml:"eventId"`
//	Event  rkquery.Event          `json:"-" yaml:"-"`
//	Logger *zap.Logger            `json:"-" yaml:"-"`
//	Errors []error                `json:"errors" yaml:"errors"`
//	Values map[string]interface{} `json:"values" yaml:"values"`
//}
//
//func New() *EventContext {
//	ctx := &EventContext{
//		EventId: rkcommon.GenerateRequestId(),
//		Errors: make([]error, 0),
//		Values: make(map[string]interface{}),
//	}
//
//	ctx.Event = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault().GetEventFactory().CreateEvent()
//	ctx.Event.SetEventId(ctx.EventId)
//
//	ctx.Logger = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault().GetLogger().With(zap.String("eventId", ctx.EventId))
//
//	return ctx
//}
//
//func (ctx *EventContext) WithEvent(event rkquery.Event) *EventContext {
//	if event != nil {
//		ctx.Event = event
//	}
//	return ctx
//}
//
//func (ctx *EventContext) WithLogger(logger *zap.Logger) *EventContext {
//	if logger != nil {
//		ctx.Logger = logger
//	}
//	return ctx
//}
//
//func (ctx *EventContext) WithEventId(id string) *EventContext {
//	if len(id) > 0 {
//		ctx.Event.SetEventId(id)
//		ctx.Logger = ctx.Logger.With(zap.String("eventId", id))
//	}
//
//	return ctx
//}
//
//func (ctx *EventContext) AddError(err error) {
//	if err != nil {
//		ctx.Errors = append(ctx.Errors, err)
//	}
//}
//
//func (ctx *EventContext) AddErrors(errs ...error) {
//	ctx.Errors = append(ctx.Errors, errs...)
//}
//
//func (ctx *EventContext) LogError(err error, level zapcore.Level) {
//	if err != nil {
//		ctx.Event.AddErr(err)
//		switch level {
//		case zapcore.InfoLevel:
//			ctx.Logger.Info()
//		}
//	}
//}
