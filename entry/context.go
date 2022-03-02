// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// GlobalAppCtx global application context
	GlobalAppCtx = &appContext{
		startTime: time.Now(),
		entries: map[string]map[string]Entry{
			appInfoEntryType: {
				appInfoEntryName: appInfoEntryDefault(),
			},
		},
		appInfoEntry:  appInfoEntryDefault(),
		shutdownSig:   make(chan os.Signal),
		shutdownHooks: make(map[string]ShutdownHook),
		userValues:    make(map[string]interface{}),
	}

	// List of entry registration function
	entryRegFuncList = []RegFunc{
		registerAppInfoEntry,
		registerLoggerEntry,
		registerEventEntry,
		registerConfigEntry,
		registerCertEntry,
	}

	LoggerEntryNoop   = NewLoggerEntryNoop()
	LoggerEntryStdout = NewLoggerEntryStdout()
	EventEntryNoop    = NewEventEntryNoop()
	EventEntryStdout  = NewEventEntryStdout()
)

// ShutdownHook defines interface of shutdown hook
type ShutdownHook func()

// Init global app context with bellow fields.
func init() {
	signal.Notify(GlobalAppCtx.shutdownSig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
}

// Application context which contains bellow fields.
type appContext struct {
	// It is not recommended override this value since StartTime would be assigned to current time
	// at beginning of go process in init() function.
	startTime    time.Time     `json:"startTime" yaml:"startTime"`
	appInfoEntry *appInfoEntry `json:"appInfoEntry" yaml:"appInfoEntry"`
	bootConfig   interface{}   `json:"-" yaml:"-"`

	entries map[string]map[string]Entry `json:"-" yaml:"-"`

	userValues    map[string]interface{}  `json:"userValues" yaml:"userValues"`
	shutdownSig   chan os.Signal          `json:"shutdownSig" yaml:"shutdownSig"`
	shutdownHooks map[string]ShutdownHook `json:"shutdownHooks" yaml:"shutdownHooks"`
}

// RegisterEntryRegFunc register user defined registration function.
// rkboot.Bootstrap will iterate every registration function and call it
func RegisterEntryRegFunc(regFunc RegFunc) {
	if regFunc == nil {
		return
	}
	entryRegFuncList = append(entryRegFuncList, regFunc)
}

// ListEntryRegFunc list user defined registration functions.
func ListEntryRegFunc() []RegFunc {
	// make a copy of the list
	res := make([]RegFunc, 0)
	for i := range entryRegFuncList {
		res = append(res, entryRegFuncList[i])
	}

	return res
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

// ************************************
// ****** App info Entry related ******
// ************************************

func (ctx *appContext) GetAppInfoEntry() *appInfoEntry {
	return ctx.appInfoEntry
}

func (ctx *appContext) GetConfigEntry(entryName string) *ConfigEntry {
	entries := ctx.entries[ConfigEntryType]

	if v, ok := entries[entryName]; ok {
		return v.(*ConfigEntry)
	}

	return nil
}

func (ctx *appContext) GetLoggerEntry(entryName string) *LoggerEntry {
	entries := ctx.entries[LoggerEntryType]

	if v, ok := entries[entryName]; ok {
		return v.(*LoggerEntry)
	}

	return nil
}

func (ctx *appContext) GetEventEntry(entryName string) *LoggerEntry {
	entries := ctx.entries[EventEntryType]

	if v, ok := entries[entryName]; ok {
		return v.(*LoggerEntry)
	}

	return nil
}

func (ctx *appContext) GetCertEntry(entryName string) *CertEntry {
	entries := ctx.entries[CertEntryType]

	if v, ok := entries[entryName]; ok {
		return v.(*CertEntry)
	}

	return nil
}

func (ctx *appContext) AddEntry(entry Entry) {
	if entry == nil {
		return
	}

	if v, ok := ctx.entries[entry.GetType()]; !ok {
		ctx.entries[entry.GetType()] = map[string]Entry{
			entry.GetName(): entry,
		}
	} else {
		v[entry.GetName()] = entry
	}
}

func (ctx *appContext) clearEntries() {
	ctx.entries = map[string]map[string]Entry{}
}

func (ctx *appContext) GetEntry(entryType, entryName string) Entry {
	if v, ok := ctx.entries[entryType]; ok {
		return v[entryName]
	}

	return nil
}

func (ctx *appContext) RemoveEntry(entry Entry) {
	if v, ok := ctx.entries[entry.GetType()]; ok {
		delete(v, entry.GetName())
	}
}

func (ctx *appContext) RemoveEntryByType(entryType string) {
	delete(ctx.entries, entryType)
}

func (ctx *appContext) ListEntriesByType(entryType string) map[string]Entry {
	if v, ok := ctx.entries[entryType]; ok {
		return v
	}

	return map[string]Entry{}
}

func (ctx *appContext) ListEntries() map[string]map[string]Entry {
	return ctx.entries
}

// ***********************************
// ****** Shutdown hook related ******
// ***********************************

// GetUpTime returns uptime of application from StartTime.
func (ctx *appContext) GetUpTime() time.Duration {
	return time.Since(ctx.startTime)
}

// GetStartTime returns start time of application.
func (ctx *appContext) GetStartTime() time.Time {
	return ctx.startTime
}

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
