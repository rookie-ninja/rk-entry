package rkentry

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithNameProm_HappyCase(t *testing.T) {
	entry := RegisterPromEntry(WithNameProm("ut-name"))

	assert.Equal(t, "ut-name", entry.EntryName)
	assert.NotEmpty(t, entry.GetDescription())
}

func TestPromEntry_UnmarshalJSON(t *testing.T) {
	entry := RegisterPromEntry()
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestPromEntry_RegisterCollectors(t *testing.T) {
	entry := RegisterPromEntry(WithPromRegistryProm(prometheus.NewRegistry()))

	assert.Nil(t, entry.RegisterCollectors(prometheus.NewGoCollector()))
	assert.NotNil(t, entry.RegisterCollectors(prometheus.NewGoCollector()))
}

func TestPromEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterPromEntry()

	entry.Bootstrap(context.TODO())
}

func TestPromEntry_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterPromEntry()

	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestWithPortProm_HappyCase(t *testing.T) {
	entry := RegisterPromEntry(WithPortProm(1949))

	assert.Equal(t, uint64(1949), entry.Port)
}

func TestWithPathProm_HappyCase(t *testing.T) {
	entry := RegisterPromEntry(WithPathProm("ut"))

	assert.Equal(t, "/ut", entry.Path)
}

func TestWithZapLoggerEntryProm_HappyCase(t *testing.T) {
	loggerEntry := NoopZapLoggerEntry()

	entry := RegisterPromEntry(WithZapLoggerEntryProm(loggerEntry))

	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestWithEventLoggerEntryProm_HappyCase(t *testing.T) {
	loggerEntry := NoopEventLoggerEntry()

	entry := RegisterPromEntry(WithEventLoggerEntryProm(loggerEntry))

	assert.Equal(t, loggerEntry, entry.EventLoggerEntry)
}

func TestWithPusherProm_HappyCase(t *testing.T) {
	pusher, _ := NewPushGatewayPusher()

	entry := RegisterPromEntry(WithPusherProm(pusher))

	assert.Equal(t, pusher, entry.Pusher)
}

func TestWithPromRegistryProm_HappyCase(t *testing.T) {
	registry := prometheus.NewRegistry()

	entry := RegisterPromEntry(WithPromRegistryProm(registry))

	assert.Equal(t, registry, entry.Registry)
}

func TestNewPromEntry_HappyCase(t *testing.T) {
	port := uint64(1949)
	path := "/ut"
	zapLoggerEntry := NoopZapLoggerEntry()
	eventLoggerEntry := NoopEventLoggerEntry()
	pusher, _ := NewPushGatewayPusher()
	registry := prometheus.NewRegistry()

	entry := RegisterPromEntry(
		WithPortProm(port),
		WithPathProm(path),
		WithZapLoggerEntryProm(zapLoggerEntry),
		WithEventLoggerEntryProm(eventLoggerEntry),
		WithPusherProm(pusher),
		WithPromRegistryProm(registry))

	assert.Equal(t, port, entry.Port)
	assert.Equal(t, path, entry.Path)
	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
	assert.Equal(t, pusher, entry.Pusher)
	assert.Equal(t, registry, entry.Registry)
}

func TestPromEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterPromEntry()
	assert.Equal(t, PromEntryNameDefault, entry.GetName())
}

func TestPromEntry_GetType_HappyCase(t *testing.T) {
	entry := RegisterPromEntry()
	assert.Equal(t, PromEntryType, entry.GetType())
}

func TestPromEntry_String_HappyCase(t *testing.T) {
	entry := RegisterPromEntry()

	str := entry.String()

	assert.Contains(t, str, "entryName")
	assert.Contains(t, str, "entryType")
	assert.Contains(t, str, "entryDescription")
	assert.Contains(t, str, "path")
	assert.Contains(t, str, "port")
}

func TestRegisterPromEntryWithConfig(t *testing.T) {
	// with disabled
	config := &BootConfigProm{
		Enabled: false,
	}
	assert.Nil(t, RegisterPromEntryWithConfig(config, "", 1, nil, nil, nil))

	// with enabled
	config.Enabled = true
	config.Pusher.Enabled = true
	assert.NotNil(t, RegisterPromEntryWithConfig(config, "", 1, nil, nil, nil))
}
