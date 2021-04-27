// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	DefaultZapLoggerEntryName   = "zap-logger-default"
	DefaultEventLoggerEntryName = "event-logger-default"
)

var (
	// Global application context
	GlobalAppCtx = &appContext{
		StartTime:     time.Now(),
		BasicEntries:  make(map[string]Entry),
		Entries:       make(map[string]Entry),
		ViperConfigs:  make(map[string]Entry),
		ShutdownSig:   make(chan os.Signal),
		ShutdownHooks: make(map[string]ShutdownHook),
		UserValues:    make(map[string]interface{}),
	}
	// list of entry registration function
	entryRegFuncList = make([]EntryRegFunc, 0)
	// list of basic entry registration function
	basicEntryRegFuncList = make([]EntryRegFunc, 0)
)

type ShutdownHook func()

// init global app context with bellow fields.
// 1: Default zap logger entry with name of "zap-logger-default" whose output path is stdout.
//    Please refer to rklogger.NewZapStdoutConfig.
// 2: Default event logger entry with name of "event-logger-default" whose output path is stdout with RK format.
//    Please refer to rkquery.NewZapEventConfig.
// 3: RK Entry registration function would be registered.
func init() {
	// init application logger with zap logger.
	defaultZapLoggerConfig := rklogger.NewZapStdoutConfig()
	defaultZapLogger, _ := defaultZapLoggerConfig.Build()

	// add application logger of zap logger into GlobalAppCtx.
	RegisterZapLoggerEntry(
		WithNameZap(DefaultZapLoggerEntryName),
		WithLoggerZap(defaultZapLogger, defaultZapLoggerConfig, nil))

	// init event logger.
	defaultEventLoggerConfig := rklogger.NewZapEventConfig()
	defaultEventLogger, _ := defaultEventLoggerConfig.Build()

	// add event logger with zap logger injected into GlobalAppCtx.
	eventLoggerEntry := RegisterEventLoggerEntry(
		WithNameEvent(DefaultEventLoggerEntryName),
		WithEventFactoryEvent(
			rkquery.NewEventFactory(
				rkquery.WithLogger(defaultEventLogger))))

	eventLoggerEntry.loggerConfig = defaultEventLoggerConfig

	// register rk style entries here including RKEntry which contains basic information,
	// and application logger and event logger, otherwise, we will have import cycle
	basicEntryRegFuncList = append(basicEntryRegFuncList,
		RegisterAppInfoEntriesFromConfig,
		RegisterZapLoggerEntriesWithConfig,
		RegisterEventLoggerEntriesWithConfig,
		RegisterViperEntriesWithConfig,
		RegisterCertEntriesFromConfig)

	signal.Notify(GlobalAppCtx.ShutdownSig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
}

// Application context which contains bellow fields.
// 1: StartTime - Application start time.
// 2: BasicEntries - Basic entries contains default zap logger entry, event logger entry and rk entry by default.
// 3: Entries - User entries registered from user code.
// 4: ViperConfigs - Viper configs registered from user code.
// 5: UserValues - User K/V registered from code.
// 6: ShutdownSig - Shutdown signals which contains syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT.
// 7: ShutdownHooks - Shutdown hooks registered from user code.
type appContext struct {
	// It is not recommended to override this value since StartTime would be assigned to current time
	// at beginning of go process in init() function.
	StartTime     time.Time               `json:"start_time"`
	BasicEntries  map[string]Entry        `json:"basic_entries"`
	Entries       map[string]Entry        `json:"entries"`
	ViperConfigs  map[string]Entry        `json:"viper_configs"`
	UserValues    map[string]interface{}  `json:"user_values"`
	ShutdownSig   chan os.Signal          `json:"shutdown_sig"`
	ShutdownHooks map[string]ShutdownHook `json:"shutdown_hooks"`
}

// Register user defined registration function.
// rkboot.Bootstrap will iterate every registration function and call it
func RegisterEntryRegFunc(regFunc EntryRegFunc) {
	if regFunc == nil {
		return
	}
	entryRegFuncList = append(entryRegFuncList, regFunc)
}

// List user defined registration functions.
func ListEntryRegFunc() []EntryRegFunc {
	// make a copy of the list
	res := make([]EntryRegFunc, 0)
	for i := range entryRegFuncList {
		res = append(res, entryRegFuncList[i])
	}

	return res
}

// Internal use only, please do not add entries registration function via this.
// Please use RegisterEntryRegFunc instead.
func RegisterBasicEntriesFromConfig(configFilePath string) {
	for i := range basicEntryRegFuncList {
		entries := basicEntryRegFuncList[i](configFilePath)

		for k, v := range entries {
			GlobalAppCtx.BasicEntries[k] = v
			v.Bootstrap(context.Background())
		}
	}
}

// ********************************
// ****** User value related ******
// ********************************

// Add value to GlobalAppCtx.
func (ctx *appContext) AddValue(key string, value interface{}) {
	ctx.UserValues[key] = value
}

// Get value from GlobalAppCtx.
func (ctx *appContext) GetValue(key string) interface{} {
	return ctx.UserValues[key]
}

// List values from GlobalAppCtx.
func (ctx *appContext) ListValues() map[string]interface{} {
	return ctx.UserValues
}

// Remove value from GlobalAppCtx.
func (ctx *appContext) RemoveValue(key string) {
	delete(ctx.UserValues, key)
}

// Clear values from GlobalAppCtx.
func (ctx *appContext) ClearValues() {
	for k := range ctx.UserValues {
		delete(ctx.UserValues, k)
	}
}

// **************************************
// ****** Zap Logger Entry related ******
// **************************************

// Returns entry name if entry is not nil, otherwise, return an empty string.
// Entry will be added into map of appContext.BasicEntries in order to distinguish between user entries and basic entries.
func (ctx *appContext) addZapLoggerEntry(entry *ZapLoggerEntry) string {
	if entry == nil {
		return ""
	}

	ctx.BasicEntries[entry.GetName()] = entry
	return entry.GetName()
}

// Get zap logger entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ZapLoggerEntry, we will just return nil.
func (ctx *appContext) GetZapLoggerEntry(name string) *ZapLoggerEntry {
	if val, ok := GlobalAppCtx.BasicEntries[name]; ok {
		if res, ok := val.(*ZapLoggerEntry); ok {
			return res
		}
	}

	return nil
}

// Remove zap logger entry.
func (ctx *appContext) RemoveZapLoggerEntry(name string) bool {
	if val, ok := GlobalAppCtx.BasicEntries[name]; ok {
		if _, ok := val.(*ZapLoggerEntry); ok {
			delete(GlobalAppCtx.BasicEntries, name)
			return true
		}
	}

	return false
}

// Returns map of zap logger entries in appContext.BasicEntries.
func (ctx *appContext) ListZapLoggerEntries() map[string]*ZapLoggerEntry {
	res := make(map[string]*ZapLoggerEntry, 0)

	for k, v := range ctx.BasicEntries {
		if logger, ok := v.(*ZapLoggerEntry); ok {
			res[k] = logger
		}
	}
	return res
}

// Get zap logger entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ZapLoggerEntry, we will just return nil.
//
// Please call bellow functions if you want to get logger entry instead of raw entry.
// zap.Logger related function:
// 1: GetZapLogger
// 2: GetZapLoggerConfig
// 3: GetZapLoggerEntry
// 4: GetZapLoggerDefault
// 5: GetZapLoggerConfigDefault
// 6: GetZapLoggerEntryDefault
func (ctx *appContext) GetZapLogger(name string) *zap.Logger {
	if val, ok := GlobalAppCtx.BasicEntries[name]; ok {
		if res, ok := val.(*ZapLoggerEntry); ok {
			return res.GetLogger()
		}
	}

	return nil
}

// Get zap logger config from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ZapLoggerEntry, we will just return nil.
func (ctx *appContext) GetZapLoggerConfig(name string) *zap.Config {
	if val, ok := GlobalAppCtx.BasicEntries[name]; ok {
		if res, ok := val.(*ZapLoggerEntry); ok {
			return res.GetLoggerConfig()
		}
	}

	return nil
}

// Get default rkentry.ZapLoggerEntry from GlobalAppCtx.
func (ctx *appContext) GetZapLoggerEntryDefault() *ZapLoggerEntry {
	return ctx.GetZapLoggerEntry(DefaultZapLoggerEntryName)
}

// Get default zap.Logger from GlobalAppCtx.
func (ctx *appContext) GetZapLoggerDefault() *zap.Logger {
	return ctx.GetZapLogger(DefaultZapLoggerEntryName)
}

// Get default zap.Config from GlobalAppCtx.
func (ctx *appContext) GetZapLoggerConfigDefault() *zap.Config {
	return ctx.GetZapLoggerConfig(DefaultZapLoggerEntryName)
}

// ****************************************
// ****** Event Logger Entry related ******
// ****************************************

// Returns entry name if entry is not nil, otherwise, return an empty string.
// Entry will be added into map of appContext.BasicEntries in order to distinguish between user entries
// and RK default entries.
//
// Please do NOT add other entries by calling this function although it would do no harm to context.
func (ctx *appContext) addEventLoggerEntry(entry *EventLoggerEntry) string {
	if entry == nil {
		return ""
	}

	ctx.BasicEntries[entry.GetName()] = entry
	return entry.GetName()
}

// Get event logger entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.EventLoggerEntry, we will just return nil.
//
// rkquery.Event related function:
// 1: GetEventLoggerEntry
// 2: GetEventLoggerEntryDefault
// 3: GetEventFactory
// 4: GetEventHelper
func (ctx *appContext) GetEventLoggerEntry(name string) *EventLoggerEntry {
	if val, ok := GlobalAppCtx.BasicEntries[name]; ok {
		if res, ok := val.(*EventLoggerEntry); ok {
			return res
		}
	}

	return nil
}

// Remove event logger entry.
func (ctx *appContext) RemoveEventLoggerEntry(name string) bool {
	if val, ok := GlobalAppCtx.BasicEntries[name]; ok {
		if _, ok := val.(*EventLoggerEntry); ok {
			delete(GlobalAppCtx.BasicEntries, name)
			return true
		}
	}

	return false
}

// Returns map of event logger entries in appContext.BasicEntries.
func (ctx *appContext) ListEventLoggerEntries() map[string]*EventLoggerEntry {
	res := make(map[string]*EventLoggerEntry, 0)

	for k, v := range ctx.BasicEntries {
		if logger, ok := v.(*EventLoggerEntry); ok {
			res[k] = logger
		}
	}
	return res
}

// Get rkquery.EventFactory from GlobalAppCtx with name.
func (ctx *appContext) GetEventFactory(name string) *rkquery.EventFactory {
	if entry := ctx.GetEventLoggerEntry(name); entry != nil {
		return entry.GetEventFactory()
	}

	return nil
}

// Get rkquery.EventHelper from GlobalAppCtx with name.
func (ctx *appContext) GetEventHelper(name string) *rkquery.EventHelper {
	if entry := ctx.GetEventLoggerEntry(name); entry != nil {
		return entry.GetEventHelper()
	}

	return nil
}

// Get default rkentry.EventLoggerEntry from GlobalAppCtx.
func (ctx *appContext) GetEventLoggerEntryDefault() *EventLoggerEntry {
	return ctx.GetEventLoggerEntry(DefaultEventLoggerEntryName)
}

// ************************************
// ******** Cert Entry related ********
// ************************************

// Returns entry name if entry is not nil, otherwise, return an empty string.
// Entry will be added into map of appContext.BasicEntries in order to distinguish between user entries
// and RK default entries.
//
// Please do NOT add other entries by calling this function although it would do no harm to context.
func (ctx *appContext) addCertEntry(entry Entry) string {
	if entry == nil {
		return ""
	}

	ctx.BasicEntries[entry.GetName()] = entry
	return entry.GetName()
}

// Get event logger entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.EventLoggerEntry, we will just return nil.
func (ctx *appContext) GetCertEntry() *CertEntry {
	if val, ok := GlobalAppCtx.BasicEntries[CertEntryName]; ok {
		if res, ok := val.(*CertEntry); ok {
			return res
		}
	}

	return nil
}

// ************************************
// ****** App info Entry related ******
// ************************************

// Returns entry name if entry is not nil, otherwise, return an empty string.
// Entry will be added into map of appContext.BasicEntries in order to distinguish between user entries
// and RK default entries.
//
// Please do NOT add other entries by calling this function although it would do no harm to context.
func (ctx *appContext) addAppInfoEntry(entry Entry) string {
	if entry == nil {
		return ""
	}

	ctx.BasicEntries[entry.GetName()] = entry
	return entry.GetName()
}

// Get rkentry.AppInfoEntry.
func (ctx *appContext) GetAppInfoEntry() *AppInfoEntry {
	if rawEntry, ok := ctx.BasicEntries[AppInfoEntryName]; ok {
		if res, ok := rawEntry.(*AppInfoEntry); ok {
			return res
		}
	}

	return AppInfoEntryDefault()
}

// Get up time of application from StartTime.
func (ctx *appContext) GetUpTime() time.Duration {
	return time.Since(ctx.StartTime)
}

// *********************************
// ****** Viper Entry related ******
// *********************************

// Add viper config entry into GlobalAppCtx.
func (ctx *appContext) AddViperEntry(entry *ViperEntry) {
	if entry == nil {
		return
	}

	ctx.ViperConfigs[entry.GetName()] = entry
}

// Get viper config.
func (ctx *appContext) GetViperEntry(name string) *ViperEntry {
	if val, ok := ctx.ViperConfigs[name]; ok {
		return val.(*ViperEntry)
	}

	return nil
}

func (ctx *appContext) ListViperEntries() map[string]*ViperEntry {
	res := make(map[string]*ViperEntry)
	for k, v := range ctx.ViperConfigs {
		res[k] = v.(*ViperEntry)
	}

	return res
}

func (ctx *appContext) clearViperEntries() {
	for k := range ctx.ViperConfigs {
		delete(ctx.ViperConfigs, k)
	}
}

// ***********************************
// ****** Shutdown hook related ******
// ***********************************

func (ctx *appContext) AddShutdownHook(name string, f ShutdownHook) {
	if f == nil {
		return
	}
	ctx.ShutdownHooks[name] = f
}

func (ctx *appContext) GetShutdownHook(name string) ShutdownHook {
	return ctx.ShutdownHooks[name]
}

func (ctx *appContext) ListShutdownHooks() map[string]ShutdownHook {
	return ctx.ShutdownHooks
}

func (ctx *appContext) clearShutdownHooks() {
	for k := range ctx.ShutdownHooks {
		delete(ctx.ShutdownHooks, k)
	}
}

// ***********************************
// ****** User entry related *********
// ***********************************

func (ctx *appContext) AddEntry(entry Entry) {
	if entry == nil {
		return
	}
	ctx.Entries[entry.GetName()] = entry
}

func (ctx *appContext) GetEntry(name string) Entry {
	return ctx.Entries[name]
}

func (ctx *appContext) MergeEntries(entries map[string]Entry) {
	for k, v := range entries {
		ctx.Entries[k] = v
	}
}

func (ctx *appContext) ListEntries() map[string]Entry {
	return ctx.Entries
}

func (ctx *appContext) clearEntries() {
	for k := range ctx.Entries {
		delete(ctx.Entries, k)
	}
}

// *************************************
// ****** Shutdown sig related *********
// *************************************

// Wait for shutdown signal
func (ctx *appContext) WaitForShutdownSig() {
	<-ctx.ShutdownSig
}
