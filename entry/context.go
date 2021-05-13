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
	DefaultZapLoggerEntryName   = "zapLoggerDefault"
	DefaultEventLoggerEntryName = "eventLoggerDefault"
)

var (
	// Global application context
	GlobalAppCtx = &appContext{
		startTime:          time.Now(),
		appInfoEntry:       AppInfoEntryDefault(),  // internal entry
		zapLoggerEntries:   make(map[string]Entry), // internal entry
		eventLoggerEntries: make(map[string]Entry), // internal entry
		configEntries:      make(map[string]Entry), // internal entry
		certEntries:        make(map[string]Entry), // internal entry
		externalEntries:    make(map[string]Entry), // external entry
		shutdownSig:        make(chan os.Signal),
		shutdownHooks:      make(map[string]ShutdownHook),
		userValues:         make(map[string]interface{}),
	}
	// list of entry registration function
	entryRegFuncList = make([]EntryRegFunc, 0)

	// list of internal entry registration function
	internalEntryRegFuncList = make([]EntryRegFunc, 0)
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

	eventLoggerEntry.LoggerConfig = defaultEventLoggerConfig

	// register rk style entries here including RKEntry which contains basic information,
	// and application logger and event logger, otherwise, we will have import cycle
	internalEntryRegFuncList = append(internalEntryRegFuncList,
		RegisterAppInfoEntriesFromConfig,
		RegisterZapLoggerEntriesWithConfig,
		RegisterEventLoggerEntriesWithConfig,
		RegisterConfigEntriesWithConfig,
		RegisterCertEntriesFromConfig)

	signal.Notify(GlobalAppCtx.shutdownSig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
}

// Application context which contains bellow fields.
// 1: startTime - Application start time.
// 2: appInfoEntry - AppInfoEntry.
// 3: zapLoggerEntries - List of ZapLoggerEntry.
// 4: eventLoggerEntries - List of EventLoggerEntry.
// 5: configEntries - List of ConfigEntry.
// 6: certEntries - List of ConfigEntry.
// 7: externalEntries - User entries registered from user code.
// 8: userValues - User K/V registered from code.
// 9: shutdownSig - Shutdown signals which contains syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT.
// 10: shutdownHooks - Shutdown hooks registered from user code.
type appContext struct {
	// It is not recommended to override this value since StartTime would be assigned to current time
	// at beginning of go process in init() function.
	startTime          time.Time               `json:"startTime" yaml:"startTime"`
	appInfoEntry       Entry                   `json:"appInfoEntry" yaml:"appInfoEntry"`
	zapLoggerEntries   map[string]Entry        `json:"zapLoggerEntries" yaml:"zapLoggerEntries"`
	eventLoggerEntries map[string]Entry        `json:"eventLoggerEntries" yaml:"eventLoggerEntries"`
	configEntries      map[string]Entry        `json:"configEntries" yaml:"configEntries"`
	certEntries        map[string]Entry        `json:"certEntries" yaml:"certEntries"`
	externalEntries    map[string]Entry        `json:"externalEntries" yaml:"externalEntries"`
	userValues         map[string]interface{}  `json:"userValues" yaml:"userValues"`
	shutdownSig        chan os.Signal          `json:"shutdownSig" yaml:"shutdownSig"`
	shutdownHooks      map[string]ShutdownHook `json:"shutdownHooks" yaml:"shutdownHooks"`
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
func RegisterInternalEntriesFromConfig(configFilePath string) {
	ctx := context.Background()

	for i := range internalEntryRegFuncList {
		entries := internalEntryRegFuncList[i](configFilePath)
		for _, v := range entries {
			v.Bootstrap(ctx)
		}
	}
}

// ********************************
// ****** User value related ******
// ********************************

// Add value to GlobalAppCtx.
func (ctx *appContext) AddValue(key string, value interface{}) {
	ctx.userValues[key] = value
}

// Get value from GlobalAppCtx.
func (ctx *appContext) GetValue(key string) interface{} {
	return ctx.userValues[key]
}

// List values from GlobalAppCtx.
func (ctx *appContext) ListValues() map[string]interface{} {
	return ctx.userValues
}

// Remove value from GlobalAppCtx.
func (ctx *appContext) RemoveValue(key string) {
	delete(ctx.userValues, key)
}

// Clear values from GlobalAppCtx.
func (ctx *appContext) ClearValues() {
	for k := range ctx.userValues {
		delete(ctx.userValues, k)
	}
}

// **************************************
// ****** Zap Logger Entry related ******
// **************************************

// Returns entry name if entry is not nil, otherwise, return an empty string.
func (ctx *appContext) AddZapLoggerEntry(entry *ZapLoggerEntry) string {
	if entry == nil {
		return ""
	}

	ctx.zapLoggerEntries[entry.GetName()] = entry
	return entry.GetName()
}

// Get zap logger entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ZapLoggerEntry, we will just return nil.
func (ctx *appContext) GetZapLoggerEntry(name string) *ZapLoggerEntry {
	if val, ok := GlobalAppCtx.zapLoggerEntries[name]; ok {
		if res, ok := val.(*ZapLoggerEntry); ok {
			return res
		}
	}

	return nil
}

// Remove zap logger entry.
func (ctx *appContext) RemoveZapLoggerEntry(name string) bool {
	if val, ok := GlobalAppCtx.zapLoggerEntries[name]; ok {
		if _, ok := val.(*ZapLoggerEntry); ok {
			delete(GlobalAppCtx.zapLoggerEntries, name)
			return true
		}
	}

	return false
}

// Returns map of zap logger entries.
func (ctx *appContext) ListZapLoggerEntries() map[string]*ZapLoggerEntry {
	res := make(map[string]*ZapLoggerEntry, 0)

	for k, v := range ctx.zapLoggerEntries {
		if logger, ok := v.(*ZapLoggerEntry); ok {
			res[k] = logger
		}
	}
	return res
}

// Returns map of zap logger entries as Entry.
func (ctx *appContext) ListZapLoggerEntriesRaw() map[string]Entry {
	return ctx.zapLoggerEntries
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
	if val, ok := GlobalAppCtx.zapLoggerEntries[name]; ok {
		if res, ok := val.(*ZapLoggerEntry); ok {
			return res.GetLogger()
		}
	}

	return nil
}

// Get zap logger config from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ZapLoggerEntry, we will just return nil.
func (ctx *appContext) GetZapLoggerConfig(name string) *zap.Config {
	if val, ok := GlobalAppCtx.zapLoggerEntries[name]; ok {
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
func (ctx *appContext) AddEventLoggerEntry(entry *EventLoggerEntry) string {
	if entry == nil {
		return ""
	}

	ctx.eventLoggerEntries[entry.GetName()] = entry
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
	if val, ok := GlobalAppCtx.eventLoggerEntries[name]; ok {
		if res, ok := val.(*EventLoggerEntry); ok {
			return res
		}
	}

	return nil
}

// Remove event logger entry.
func (ctx *appContext) RemoveEventLoggerEntry(name string) bool {
	if val, ok := GlobalAppCtx.eventLoggerEntries[name]; ok {
		if _, ok := val.(*EventLoggerEntry); ok {
			delete(GlobalAppCtx.eventLoggerEntries, name)
			return true
		}
	}

	return false
}

// Returns map of event logger entries.
func (ctx *appContext) ListEventLoggerEntries() map[string]*EventLoggerEntry {
	res := make(map[string]*EventLoggerEntry, 0)

	for k, v := range ctx.eventLoggerEntries {
		if logger, ok := v.(*EventLoggerEntry); ok {
			res[k] = logger
		}
	}
	return res
}

// Returns map of zap logger entries as Entry.
func (ctx *appContext) ListEventLoggerEntriesRaw() map[string]Entry {
	return ctx.eventLoggerEntries
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
func (ctx *appContext) AddCertEntry(entry *CertEntry) string {
	if entry == nil {
		return ""
	}

	ctx.certEntries[entry.GetName()] = entry
	return entry.GetName()
}

// Get event logger entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.CertEntry, we will just return nil.
func (ctx *appContext) GetCertEntry(name string) *CertEntry {
	if val, ok := GlobalAppCtx.certEntries[name]; ok {
		if res, ok := val.(*CertEntry); ok {
			return res
		}
	}

	return nil
}

// Remove cert entry.
func (ctx *appContext) RemoveCertEntry(name string) bool {
	if val, ok := GlobalAppCtx.certEntries[name]; ok {
		if _, ok := val.(*CertEntry); ok {
			delete(GlobalAppCtx.certEntries, name)
			return true
		}
	}

	return false
}

// Returns map of cert entries.
func (ctx *appContext) ListCertEntries() map[string]*CertEntry {
	res := make(map[string]*CertEntry)
	for k, v := range ctx.certEntries {
		res[k] = v.(*CertEntry)
	}

	return res
}

// Returns map of cert entries as Entry.
func (ctx *appContext) ListCertEntriesRaw() map[string]Entry {
	return GlobalAppCtx.certEntries
}

// Internal use only.
func (ctx *appContext) clearCertEntries() {
	for k := range ctx.certEntries {
		delete(ctx.certEntries, k)
	}
}

// ************************************
// ****** App info Entry related ******
// ************************************

// Returns entry name if entry is not nil, otherwise, return an empty string.
// Entry will be added into map of appContext.BasicEntries in order to distinguish between user entries
// and RK default entries.
//
// Please do NOT add other entries by calling this function although it would do no harm to context.
func (ctx *appContext) SetAppInfoEntry(entry Entry) string {
	if entry == nil {
		return ""
	}

	ctx.appInfoEntry = entry
	return entry.GetName()
}

// Get rkentry.AppInfoEntry.
func (ctx *appContext) GetAppInfoEntry() *AppInfoEntry {
	return ctx.appInfoEntry.(*AppInfoEntry)
}

// Get rkentry.AppInfoEntry.
func (ctx *appContext) GetAppInfoEntryRaw() Entry {
	return ctx.appInfoEntry
}

// Get up time of application from StartTime.
func (ctx *appContext) GetUpTime() time.Duration {
	return time.Since(ctx.startTime)
}

// Get start time of application.
func (ctx *appContext) GetStartTime() time.Time {
	return ctx.startTime
}

// **********************************
// ****** Config Entry related ******
// **********************************

// Add config entry into GlobalAppCtx.
func (ctx *appContext) AddConfigEntry(entry *ConfigEntry) string {
	if entry == nil {
		return ""
	}

	ctx.configEntries[entry.GetName()] = entry
	return entry.GetName()
}

// Get config entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ConfigEntry, we will just return nil.
func (ctx *appContext) GetConfigEntry(name string) *ConfigEntry {
	if val, ok := ctx.configEntries[name]; ok {
		return val.(*ConfigEntry)
	}

	return nil
}

// Returns map of config entries.
func (ctx *appContext) ListConfigEntries() map[string]*ConfigEntry {
	res := make(map[string]*ConfigEntry)
	for k, v := range ctx.configEntries {
		res[k] = v.(*ConfigEntry)
	}

	return res
}

// Returns map of config entries as Entry.
func (ctx *appContext) ListConfigEntriesRaw() map[string]Entry {
	return ctx.configEntries
}

// Remove config entry.
func (ctx *appContext) RemoveConfigEntry(name string) bool {
	if val, ok := GlobalAppCtx.configEntries[name]; ok {
		if _, ok := val.(*ConfigEntry); ok {
			delete(GlobalAppCtx.configEntries, name)
			return true
		}
	}

	return false
}

// Internal use only.
func (ctx *appContext) clearConfigEntries() {
	for k := range ctx.configEntries {
		delete(ctx.configEntries, k)
	}
}

// ***********************************
// ****** User entry related *********
// ***********************************

// Add user entry into GlobalAppCtx.
func (ctx *appContext) AddEntry(entry Entry) {
	if entry == nil {
		return
	}
	ctx.externalEntries[entry.GetName()] = entry
}

// Get user entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.Entry, we will just return nil.
func (ctx *appContext) GetEntry(name string) Entry {
	return ctx.externalEntries[name]
}

// Merge entries.
func (ctx *appContext) MergeEntries(entries map[string]Entry) {
	for k, v := range entries {
		ctx.externalEntries[k] = v
	}
}

// Remove entry.
func (ctx *appContext) RemoveEntry(name string) bool {
	if _, ok := GlobalAppCtx.externalEntries[name]; ok {
		delete(GlobalAppCtx.externalEntries, name)
		return true
	}

	return false
}

// Returns map of config entries.
func (ctx *appContext) ListEntries() map[string]Entry {
	return ctx.externalEntries
}

// Internal use only.
func (ctx *appContext) clearEntries() {
	for k := range ctx.externalEntries {
		delete(ctx.externalEntries, k)
	}
}

// ***********************************
// ****** Shutdown hook related ******
// ***********************************

// Add shutdown hook with name.
func (ctx *appContext) AddShutdownHook(name string, f ShutdownHook) {
	if f == nil {
		return
	}
	ctx.shutdownHooks[name] = f
}

// Get shutdown hook with name.
func (ctx *appContext) GetShutdownHook(name string) ShutdownHook {
	return ctx.shutdownHooks[name]
}

// List shutdown hooks.
func (ctx *appContext) ListShutdownHooks() map[string]ShutdownHook {
	return ctx.shutdownHooks
}

// Remove shutdown hook.
func (ctx *appContext) RemoveShutdownHook(name string) bool {
	if _, ok := GlobalAppCtx.configEntries[name]; ok {
		delete(GlobalAppCtx.configEntries, name)
		return true
	}

	return false
}

// Internal use only.
func (ctx *appContext) clearShutdownHooks() {
	for k := range ctx.shutdownHooks {
		delete(ctx.shutdownHooks, k)
	}
}

// *************************************
// ****** Shutdown sig related *********
// *************************************

// Wait for shutdown signal.
func (ctx *appContext) WaitForShutdownSig() {
	<-ctx.shutdownSig
}

// Get shutdown signal.
func (ctx *appContext) GetShutdownSig() chan os.Signal {
	return ctx.shutdownSig
}
