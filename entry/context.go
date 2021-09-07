// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
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
	// DefaultZapLoggerEntryName is default ZapLoggerEntry name
	DefaultZapLoggerEntryName = "zapLoggerDefault"
	// DefaultEventLoggerEntryName is default EventLoggerEntry name
	DefaultEventLoggerEntryName = "eventLoggerDefault"
)

var (
	// GlobalAppCtx global application context
	GlobalAppCtx = &appContext{
		startTime:          time.Now(),
		appInfoEntry:       AppInfoEntryDefault(),  // internal entry
		zapLoggerEntries:   make(map[string]Entry), // internal entry
		eventLoggerEntries: make(map[string]Entry), // internal entry
		configEntries:      make(map[string]Entry), // internal entry
		certEntries:        make(map[string]Entry), // internal entry
		credEntries:        make(map[string]Entry), // internal entry
		externalEntries:    make(map[string]Entry), // external entry
		shutdownSig:        make(chan os.Signal),
		shutdownHooks:      make(map[string]ShutdownHook),
		userValues:         make(map[string]interface{}),
	}
	// List of entry registration function
	entryRegFuncList = make([]EntryRegFunc, 0)

	// List of internal entry registration function
	internalEntryRegFuncList = make([]EntryRegFunc, 0)
)

// ShutdownHook defines interface of shutdown hook
type ShutdownHook func()

// Init global app context with bellow fields.
// 1: Default zap logger entry with name of "zap-logger-default" whose output path is stdout.
//    Please refer to rklogger.NewZapStdoutConfig.
// 2: Default event logger entry with name of "event-logger-default" whose output path is stdout with RK format.
//    Please refer to rkquery.NewZapEventConfig.
// 3: RK Entry registration function would be registered.
func init() {
	// Init application logger with zap logger.
	defaultZapLoggerConfig := rklogger.NewZapStdoutConfig()
	defaultZapLogger, _ := defaultZapLoggerConfig.Build()

	// RegisterZapLoggerEntry add application logger of zap logger into GlobalAppCtx.
	RegisterZapLoggerEntry(
		WithNameZap(DefaultZapLoggerEntryName),
		WithLoggerZap(defaultZapLogger, defaultZapLoggerConfig, nil))

	// Init event logger.
	defaultEventLoggerConfig := rklogger.NewZapEventConfig()
	defaultEventLogger, _ := defaultEventLoggerConfig.Build()

	// Add event logger with zap logger injected into GlobalAppCtx.
	eventLoggerEntry := RegisterEventLoggerEntry(
		WithNameEvent(DefaultEventLoggerEntryName),
		WithEventFactoryEvent(
			rkquery.NewEventFactory(
				rkquery.WithZapLogger(defaultEventLogger))))

	eventLoggerEntry.LoggerConfig = defaultEventLoggerConfig

	// Register rk style entries here including RKEntry which contains basic information,
	// and application logger and event logger, otherwise, we will have import cycle
	internalEntryRegFuncList = append(internalEntryRegFuncList,
		RegisterRkMetaEntriesFromConfig,
		RegisterAppInfoEntriesFromConfig,
		RegisterZapLoggerEntriesWithConfig,
		RegisterEventLoggerEntriesWithConfig,
		RegisterConfigEntriesWithConfig,
		RegisterCertEntriesFromConfig,
		RegisterCredEntriesFromConfig)

	signal.Notify(GlobalAppCtx.shutdownSig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
}

// Application context which contains bellow fields.
// 1: startTime - Application start time.
// 2: appInfoEntry - AppInfoEntry.
// 3: rkMetaEntry - RkMetaEntry.
// 4: zapLoggerEntries - List of ZapLoggerEntry.
// 5: eventLoggerEntries - List of EventLoggerEntry.
// 6: configEntries - List of ConfigEntry.
// 7: certEntries - List of ConfigEntry.
// 8: externalEntries - User entries registered from user code.
// 9: userValues - User K/V registered from code.
// 10: shutdownSig - Shutdown signals which contains syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT.
// 11: shutdownHooks - Shutdown hooks registered from user code.
type appContext struct {
	// It is not recommended to override this value since StartTime would be assigned to current time
	// at beginning of go process in init() function.
	startTime          time.Time               `json:"startTime" yaml:"startTime"`
	appInfoEntry       Entry                   `json:"appInfoEntry" yaml:"appInfoEntry"`
	rkMetaEntry        Entry                   `json:"rkMetaEntry" yaml:"rkMetaEntry"`
	zapLoggerEntries   map[string]Entry        `json:"zapLoggerEntries" yaml:"zapLoggerEntries"`
	eventLoggerEntries map[string]Entry        `json:"eventLoggerEntries" yaml:"eventLoggerEntries"`
	configEntries      map[string]Entry        `json:"configEntries" yaml:"configEntries"`
	certEntries        map[string]Entry        `json:"certEntries" yaml:"certEntries"`
	credEntries        map[string]Entry        `json:"credEntries" yaml:"credEntries"`
	externalEntries    map[string]Entry        `json:"externalEntries" yaml:"externalEntries"`
	userValues         map[string]interface{}  `json:"userValues" yaml:"userValues"`
	shutdownSig        chan os.Signal          `json:"shutdownSig" yaml:"shutdownSig"`
	shutdownHooks      map[string]ShutdownHook `json:"shutdownHooks" yaml:"shutdownHooks"`
}

// RegisterEntryRegFunc register user defined registration function.
// rkboot.Bootstrap will iterate every registration function and call it
func RegisterEntryRegFunc(regFunc EntryRegFunc) {
	if regFunc == nil {
		return
	}
	entryRegFuncList = append(entryRegFuncList, regFunc)
}

// ListEntryRegFunc list user defined registration functions.
func ListEntryRegFunc() []EntryRegFunc {
	// make a copy of the list
	res := make([]EntryRegFunc, 0)
	for i := range entryRegFuncList {
		res = append(res, entryRegFuncList[i])
	}

	return res
}

// RegisterInternalEntriesFromConfig is internal use only, please do not add entries registration function via this.
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

// AddValue add value to GlobalAppCtx.
func (ctx *appContext) AddValue(key string, value interface{}) {
	ctx.userValues[key] = value
}

// GetValue returns value from GlobalAppCtx.
func (ctx *appContext) GetValue(key string) interface{} {
	return ctx.userValues[key]
}

// ListValues list values from GlobalAppCtx.
func (ctx *appContext) ListValues() map[string]interface{} {
	return ctx.userValues
}

// RemoveValue remove value from GlobalAppCtx.
func (ctx *appContext) RemoveValue(key string) {
	delete(ctx.userValues, key)
}

// ClearValues clear values from GlobalAppCtx.
func (ctx *appContext) ClearValues() {
	for k := range ctx.userValues {
		delete(ctx.userValues, k)
	}
}

// **************************************
// ****** Zap Logger Entry related ******
// **************************************

// AddZapLoggerEntry returns entry name if entry is not nil, otherwise, return an empty string.
func (ctx *appContext) AddZapLoggerEntry(entry *ZapLoggerEntry) string {
	if entry == nil {
		return ""
	}

	ctx.zapLoggerEntries[entry.GetName()] = entry
	return entry.GetName()
}

// GetZapLoggerEntry returns zap logger entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ZapLoggerEntry, we will just return nil.
func (ctx *appContext) GetZapLoggerEntry(name string) *ZapLoggerEntry {
	if val, ok := GlobalAppCtx.zapLoggerEntries[name]; ok {
		if res, ok := val.(*ZapLoggerEntry); ok {
			return res
		}
	}

	return nil
}

// RemoveZapLoggerEntry remove zap logger entry.
func (ctx *appContext) RemoveZapLoggerEntry(name string) bool {
	if val, ok := GlobalAppCtx.zapLoggerEntries[name]; ok {
		if _, ok := val.(*ZapLoggerEntry); ok {
			delete(GlobalAppCtx.zapLoggerEntries, name)
			return true
		}
	}

	return false
}

// ListZapLoggerEntries returns map of zap logger entries.
func (ctx *appContext) ListZapLoggerEntries() map[string]*ZapLoggerEntry {
	res := make(map[string]*ZapLoggerEntry, 0)

	for k, v := range ctx.zapLoggerEntries {
		if logger, ok := v.(*ZapLoggerEntry); ok {
			res[k] = logger
		}
	}
	return res
}

// ListZapLoggerEntriesRaw returns map of zap logger entries as Entry.
func (ctx *appContext) ListZapLoggerEntriesRaw() map[string]Entry {
	return ctx.zapLoggerEntries
}

// GetZapLogger return ZapLoggerEntry from GlobalAppCtx with name.
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

// GetZapLoggerConfig return zap.Config from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ZapLoggerEntry, we will just return nil.
func (ctx *appContext) GetZapLoggerConfig(name string) *zap.Config {
	if val, ok := GlobalAppCtx.zapLoggerEntries[name]; ok {
		if res, ok := val.(*ZapLoggerEntry); ok {
			return res.GetLoggerConfig()
		}
	}

	return nil
}

// GetZapLoggerEntryDefault returns default rkentry.ZapLoggerEntry from GlobalAppCtx.
func (ctx *appContext) GetZapLoggerEntryDefault() *ZapLoggerEntry {
	return ctx.GetZapLoggerEntry(DefaultZapLoggerEntryName)
}

// GetZapLoggerDefault returns default zap.Logger from GlobalAppCtx.
func (ctx *appContext) GetZapLoggerDefault() *zap.Logger {
	return ctx.GetZapLogger(DefaultZapLoggerEntryName)
}

// GetZapLoggerConfigDefault returns default zap.Config from GlobalAppCtx.
func (ctx *appContext) GetZapLoggerConfigDefault() *zap.Config {
	return ctx.GetZapLoggerConfig(DefaultZapLoggerEntryName)
}

// ****************************************
// ****** Event Logger Entry related ******
// ****************************************

// AddEventLoggerEntry teturns entry name if entry is not nil, otherwise, return an empty string.
func (ctx *appContext) AddEventLoggerEntry(entry *EventLoggerEntry) string {
	if entry == nil {
		return ""
	}

	ctx.eventLoggerEntries[entry.GetName()] = entry
	return entry.GetName()
}

// GetEventLoggerEntry returns EventLoggerEntry from GlobalAppCtx with name.
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

// RemoveEventLoggerEntry remove event logger entry.
func (ctx *appContext) RemoveEventLoggerEntry(name string) bool {
	if val, ok := GlobalAppCtx.eventLoggerEntries[name]; ok {
		if _, ok := val.(*EventLoggerEntry); ok {
			delete(GlobalAppCtx.eventLoggerEntries, name)
			return true
		}
	}

	return false
}

// ListEventLoggerEntries Returns map of EventLoggerEntry.
func (ctx *appContext) ListEventLoggerEntries() map[string]*EventLoggerEntry {
	res := make(map[string]*EventLoggerEntry, 0)

	for k, v := range ctx.eventLoggerEntries {
		if logger, ok := v.(*EventLoggerEntry); ok {
			res[k] = logger
		}
	}
	return res
}

// ListEventLoggerEntriesRaw returns map of zap logger entries as Entry.
func (ctx *appContext) ListEventLoggerEntriesRaw() map[string]Entry {
	return ctx.eventLoggerEntries
}

// GetEventFactory returns rkquery.EventFactory from GlobalAppCtx with name.
func (ctx *appContext) GetEventFactory(name string) *rkquery.EventFactory {
	if entry := ctx.GetEventLoggerEntry(name); entry != nil {
		return entry.GetEventFactory()
	}

	return nil
}

// GetEventHelper returns rkquery.EventHelper from GlobalAppCtx with name.
func (ctx *appContext) GetEventHelper(name string) *rkquery.EventHelper {
	if entry := ctx.GetEventLoggerEntry(name); entry != nil {
		return entry.GetEventHelper()
	}

	return nil
}

// GetEventLoggerEntryDefault returns default rkentry.EventLoggerEntry from GlobalAppCtx.
func (ctx *appContext) GetEventLoggerEntryDefault() *EventLoggerEntry {
	return ctx.GetEventLoggerEntry(DefaultEventLoggerEntryName)
}

// ************************************
// ******** Cred Entry related ********
// ************************************

// AddCredEntry returns entry name if entry is not nil, otherwise, return an empty string.
// Entry will be added into map of appContext.BasicEntries in order to distinguish between user entries
// and RK default entries.
//
// Please do NOT add other entries by calling this function although it would do no harm to context.
func (ctx *appContext) AddCredEntry(entry *CredEntry) string {
	if entry == nil {
		return ""
	}

	ctx.credEntries[entry.GetName()] = entry
	return entry.GetName()
}

// GetCredEntry returns EventLoggerEntry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.CredEntry, we will just return nil.
func (ctx *appContext) GetCredEntry(name string) *CredEntry {
	if val, ok := GlobalAppCtx.credEntries[name]; ok {
		if res, ok := val.(*CredEntry); ok {
			return res
		}
	}

	return nil
}

// RemoveCredEntry remove cred entry.
func (ctx *appContext) RemoveCredEntry(name string) bool {
	if val, ok := GlobalAppCtx.credEntries[name]; ok {
		if _, ok := val.(*CredEntry); ok {
			delete(GlobalAppCtx.credEntries, name)
			return true
		}
	}

	return false
}

// ListCredEntries returns map of cred entries.
func (ctx *appContext) ListCredEntries() map[string]*CredEntry {
	res := make(map[string]*CredEntry)
	for k, v := range ctx.credEntries {
		res[k] = v.(*CredEntry)
	}

	return res
}

// ListCredEntriesRaw returns map of cred entries as Entry.
func (ctx *appContext) ListCredEntriesRaw() map[string]Entry {
	return GlobalAppCtx.credEntries
}

// Internal use only.
func (ctx *appContext) clearCredEntries() {
	for k := range ctx.credEntries {
		delete(ctx.credEntries, k)
	}
}

// ************************************
// ******** Cert Entry related ********
// ************************************

// AddCertEntry returns entry name if entry is not nil, otherwise, return an empty string.
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

// GetCertEntry returns event logger entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.CertEntry, we will just return nil.
func (ctx *appContext) GetCertEntry(name string) *CertEntry {
	if val, ok := GlobalAppCtx.certEntries[name]; ok {
		if res, ok := val.(*CertEntry); ok {
			return res
		}
	}

	return nil
}

// RemoveCertEntry remove cert entry.
func (ctx *appContext) RemoveCertEntry(name string) bool {
	if val, ok := GlobalAppCtx.certEntries[name]; ok {
		if _, ok := val.(*CertEntry); ok {
			delete(GlobalAppCtx.certEntries, name)
			return true
		}
	}

	return false
}

// ListCertEntries returns map of cert entries.
func (ctx *appContext) ListCertEntries() map[string]*CertEntry {
	res := make(map[string]*CertEntry)
	for k, v := range ctx.certEntries {
		res[k] = v.(*CertEntry)
	}

	return res
}

// ListCertEntriesRaw returns map of cert entries as Entry.
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

// SetAppInfoEntry returns entry name if entry is not nil, otherwise, return an empty string.
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

// GetAppInfoEntry returns rkentry.AppInfoEntry.
func (ctx *appContext) GetAppInfoEntry() *AppInfoEntry {
	return ctx.appInfoEntry.(*AppInfoEntry)
}

// GetAppInfoEntryRaw returns rkentry.AppInfoEntry.
func (ctx *appContext) GetAppInfoEntryRaw() Entry {
	return ctx.appInfoEntry
}

// GetUpTime returns up time of application from StartTime.
func (ctx *appContext) GetUpTime() time.Duration {
	return time.Since(ctx.startTime)
}

// GetStartTime returns start time of application.
func (ctx *appContext) GetStartTime() time.Time {
	return ctx.startTime
}

// ***********************************
// ****** Rk Meta Entry related ******
// ***********************************

// SetRkMetaEntry returns entry name if entry is not nil, otherwise, return an empty string.
// Entry will be added into map of appContext.BasicEntries in order to distinguish between user entries
// and RK default entries.
//
// Please do NOT add other entries by calling this function although it would do no harm to context.
func (ctx *appContext) SetRkMetaEntry(entry Entry) string {
	if entry == nil {
		return ""
	}

	ctx.rkMetaEntry = entry
	return entry.GetName()
}

// GetRkMetaEntry returns rkentry.RkMetaEntry.
func (ctx *appContext) GetRkMetaEntry() *RkMetaEntry {
	if ctx.rkMetaEntry != nil {
		return ctx.rkMetaEntry.(*RkMetaEntry)
	}

	return nil
}

// GetRkMetaEntryRaw returns rkentry.RkMetaEntry.
func (ctx *appContext) GetRkMetaEntryRaw() Entry {
	return ctx.rkMetaEntry
}

// **********************************
// ****** Config Entry related ******
// **********************************

// AddConfigEntry adds config entry into GlobalAppCtx.
func (ctx *appContext) AddConfigEntry(entry *ConfigEntry) string {
	if entry == nil {
		return ""
	}

	ctx.configEntries[entry.GetName()] = entry
	return entry.GetName()
}

// GetConfigEntry returns config entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.ConfigEntry, we will just return nil.
func (ctx *appContext) GetConfigEntry(name string) *ConfigEntry {
	if val, ok := ctx.configEntries[name]; ok {
		return val.(*ConfigEntry)
	}

	return nil
}

// ListConfigEntries returns map of config entries.
func (ctx *appContext) ListConfigEntries() map[string]*ConfigEntry {
	res := make(map[string]*ConfigEntry)
	for k, v := range ctx.configEntries {
		res[k] = v.(*ConfigEntry)
	}

	return res
}

// ListConfigEntriesRaw returns map of config entries as Entry.
func (ctx *appContext) ListConfigEntriesRaw() map[string]Entry {
	return ctx.configEntries
}

// RemoveConfigEntry remove config entry.
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

// AddEntry add user entry into GlobalAppCtx.
func (ctx *appContext) AddEntry(entry Entry) {
	if entry == nil {
		return
	}
	ctx.externalEntries[entry.GetName()] = entry
}

// GetEntry returns user entry from GlobalAppCtx with name.
// If entry retrieved with provided name was not type of rkentry.Entry, we will just return nil.
func (ctx *appContext) GetEntry(name string) Entry {
	return ctx.externalEntries[name]
}

// MergeEntries merge entries.
func (ctx *appContext) MergeEntries(entries map[string]Entry) {
	for k, v := range entries {
		ctx.externalEntries[k] = v
	}
}

// RemoveEntry removes entry.
func (ctx *appContext) RemoveEntry(name string) bool {
	if _, ok := GlobalAppCtx.externalEntries[name]; ok {
		delete(GlobalAppCtx.externalEntries, name)
		return true
	}

	return false
}

// ListEntries returns map of config entries.
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

// AddShutdownHook add shutdown hook with name.
func (ctx *appContext) AddShutdownHook(name string, f ShutdownHook) {
	if f == nil {
		return
	}
	ctx.shutdownHooks[name] = f
}

// GetShutdownHook returns shutdown hook with name.
func (ctx *appContext) GetShutdownHook(name string) ShutdownHook {
	return ctx.shutdownHooks[name]
}

// ListShutdownHooks list shutdown hooks.
func (ctx *appContext) ListShutdownHooks() map[string]ShutdownHook {
	return ctx.shutdownHooks
}

// RemoveShutdownHook remove shutdown hook.
func (ctx *appContext) RemoveShutdownHook(name string) bool {
	if _, ok := GlobalAppCtx.shutdownHooks[name]; ok {
		delete(GlobalAppCtx.shutdownHooks, name)
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

// WaitForShutdownSig waits for shutdown signal.
func (ctx *appContext) WaitForShutdownSig() {
	<-ctx.shutdownSig
}

// GetShutdownSig returns shutdown signal.
func (ctx *appContext) GetShutdownSig() chan os.Signal {
	return ctx.shutdownSig
}
