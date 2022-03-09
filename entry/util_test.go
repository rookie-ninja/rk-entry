package rkentry

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
	"time"
)

func TestFileExists_ExpectTrue(t *testing.T) {
	filePath := path.Join(t.TempDir(), "ui-TestFileExist-ExpectTrue")
	assert.Nil(t, ioutil.WriteFile(filePath, []byte("unit-test"), 0777))
	assert.True(t, fileExists(filePath))
}

func TestFileExists_ExpectFalse(t *testing.T) {
	filePath := path.Join(t.TempDir(), "ui-TestFileExist-ExpectFalse")
	assert.False(t, fileExists(filePath))
	assert.False(t, fileExists(t.TempDir()))
}

func TestFileExists_WithEmptyFilePath(t *testing.T) {
	assert.False(t, fileExists(""))
}

func TestGetDefaultIfEmptyString_ExpectDefault(t *testing.T) {
	def := "unit-test-default"
	assert.Equal(t, def, getDefaultIfEmptyString("", def))
}

func TestGetDefaultIfEmptyString_ExpectOriginal(t *testing.T) {
	def := "unit-test-default"
	origin := "init-test-original"
	assert.Equal(t, origin, getDefaultIfEmptyString(origin, def))
}

func TestOverrideZapConfig_WithNilOverride(t *testing.T) {
	originOne := &zap.Config{}
	originTwo := &zap.Config{}

	overrideZapConfig(originOne, nil)
	assert.Equal(t, originOne, originTwo)
}

func TestOverrideZapConfig_WithSame(t *testing.T) {
	origin := &zap.Config{}
	override := &zap.Config{}

	overrideZapConfig(origin, override)
	assert.Equal(t, origin, override)
}

func TestOverrideZapConfig_HappyCase(t *testing.T) {
	origin := &zap.Config{
		Level: zap.NewAtomicLevelAt(zapcore.InfoLevel),
	}
	override := &zap.Config{
		Development:       true,
		DisableCaller:     true,
		DisableStacktrace: true,
		Encoding:          "json",
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		InitialFields:     map[string]interface{}{"key": "value"},
		ErrorOutputPaths:  []string{"logs"},
		OutputPaths:       []string{"logs"},
		Sampling:          &zap.SamplingConfig{},
		EncoderConfig: zapcore.EncoderConfig{
			CallerKey:        "ut-caller",
			ConsoleSeparator: "ut-separator",
			EncodeCaller:     func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {},
			EncodeDuration:   func(duration time.Duration, encoder zapcore.PrimitiveArrayEncoder) {},
			EncodeLevel:      func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {},
			EncodeName:       func(s string, encoder zapcore.PrimitiveArrayEncoder) {},
			EncodeTime:       func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {},
			MessageKey:       "ut-message",
			LevelKey:         "ut-level",
			TimeKey:          "ut-time",
			NameKey:          "ut-name",
			FunctionKey:      "ut-func",
			StacktraceKey:    "ut-stack",
			LineEnding:       "ut-line",
		},
	}

	overrideZapConfig(origin, override)
	assert.Equal(t, override.Development, origin.Development)
	assert.Equal(t, override.DisableCaller, origin.DisableCaller)
	assert.Equal(t, override.DisableStacktrace, origin.DisableStacktrace)
	assert.Equal(t, override.Encoding, origin.Encoding)
	assert.Equal(t, override.Level.String(), origin.Level.String())
	assert.Equal(t, override.InitialFields, origin.InitialFields)
	assert.Equal(t, override.ErrorOutputPaths, origin.ErrorOutputPaths)
	assert.Equal(t, override.OutputPaths, origin.OutputPaths)
	assert.Equal(t, override.Sampling, origin.Sampling)
	assert.Equal(t, override.EncoderConfig.CallerKey, origin.EncoderConfig.CallerKey)
	assert.Equal(t, override.EncoderConfig.ConsoleSeparator, origin.EncoderConfig.ConsoleSeparator)
	assert.Equal(t, reflect.ValueOf(override.EncoderConfig.EncodeCaller),
		reflect.ValueOf(origin.EncoderConfig.EncodeCaller))
	assert.Equal(t, reflect.ValueOf(override.EncoderConfig.EncodeDuration),
		reflect.ValueOf(origin.EncoderConfig.EncodeDuration))
	assert.Equal(t, reflect.ValueOf(override.EncoderConfig.EncodeLevel),
		reflect.ValueOf(origin.EncoderConfig.EncodeLevel))
	assert.Equal(t, reflect.ValueOf(override.EncoderConfig.EncodeName),
		reflect.ValueOf(origin.EncoderConfig.EncodeName))
	assert.Equal(t, reflect.ValueOf(override.EncoderConfig.EncodeTime),
		reflect.ValueOf(origin.EncoderConfig.EncodeTime))
	assert.Equal(t, override.EncoderConfig.MessageKey, origin.EncoderConfig.MessageKey)
	assert.Equal(t, override.EncoderConfig.LevelKey, origin.EncoderConfig.LevelKey)
	assert.Equal(t, override.EncoderConfig.TimeKey, origin.EncoderConfig.TimeKey)
	assert.Equal(t, override.EncoderConfig.NameKey, origin.EncoderConfig.NameKey)
	assert.Equal(t, override.EncoderConfig.FunctionKey, origin.EncoderConfig.FunctionKey)
	assert.Equal(t, override.EncoderConfig.StacktraceKey, origin.EncoderConfig.StacktraceKey)
	assert.Equal(t, override.EncoderConfig.LineEnding, origin.EncoderConfig.LineEnding)
}

func TestOverrideLumberjackConfig_WithNilTarget(t *testing.T) {
	originOne := &lumberjack.Logger{}
	originTwo := &lumberjack.Logger{}
	overrideLumberjackConfig(originOne, nil)
	assert.Equal(t, originOne, originTwo)
}

func TestOverrideLumberjackConfig_HappyCase(t *testing.T) {
	origin := &lumberjack.Logger{}
	override := &lumberjack.Logger{
		Compress:   true,
		LocalTime:  false,
		MaxAge:     1000,
		MaxBackups: 1000,
		MaxSize:    1000,
		Filename:   "ut-file-name",
	}
	overrideLumberjackConfig(origin, override)
	assert.Equal(t, origin, override)
}

func TestOverrideLumberjackConfig_WithSame(t *testing.T) {
	origin := &lumberjack.Logger{}
	override := &lumberjack.Logger{}
	overrideLumberjackConfig(origin, override)
	assert.Equal(t, origin, override)
}

func TestOverrideSlice_WithNilSource(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	// no panic expected
	override := make([]interface{}, 0)
	overrideSlice(nil, override)
}

func TestOverrideSlice_WithNilOverride(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	// no panic expected
	src := make([]interface{}, 0)
	overrideSlice(src, nil)
}

func TestOverrideSlice_WithUnMatchedType(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	// no panic expected
	src := []interface{}{"str"}
	override := []interface{}{false}

	overrideSlice(src, override)

	assert.Len(t, src, 1)
	assert.Equal(t, "str", src[0])

	assert.Len(t, override, 1)
	assert.Equal(t, false, override[0])
}

func TestOverrideSlice_WithMixedType(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	// no panic expected
	src := []interface{}{"str", true}
	override := []interface{}{"override-str", false}

	overrideSlice(src, override)

	assert.Len(t, src, 2)
	assert.Equal(t, "override-str", src[0])
	assert.Equal(t, false, src[1])

	assert.Len(t, override, 2)
	assert.Equal(t, "override-str", src[0])
	assert.Equal(t, false, src[1])
}

func TestOverrideSlice_WithGenericType(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	type MyStruct struct {
		Key string
	}

	// no panic expected
	src := []interface{}{
		[]*MyStruct{},
		map[string]*MyStruct{},
	}

	override := []interface{}{
		[]*MyStruct{{Key: "key"}},
		map[string]*MyStruct{"map-key": {Key: "key"}},
	}

	overrideSlice(src, override)

	assert.Equal(t, src, override)
}

func TestOverrideSlice_WithHappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	type MyStruct struct {
		Key string
	}

	// no panic expected
	src := []interface{}{
		"",
		[]string{},
		map[string]interface{}{},
		&MyStruct{},
	}

	override := []interface{}{
		"str",
		[]string{"one", "two"},
		map[string]interface{}{"key": "value"},
		&MyStruct{Key: "override"},
	}

	overrideSlice(src, override)

	// source map should be changed
	// validate string
	assert.Equal(t, "str", src[0])
	// validate list
	assert.Contains(t, src[1], "one")
	assert.Contains(t, src[1], "two")
	// validate map
	innerMap := src[2]
	assert.Equal(t, "value", innerMap.(map[string]interface{})["key"])
	// validate struct
	innerStruct := src[3]
	assert.NotNil(t, "override", innerStruct.(*MyStruct).Key)
}

func TestOverrideMap_WithNilSource(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	// no panic expected
	override := make(map[interface{}]interface{})
	overrideMap(nil, override)
}

func TestOverrideMap_WithNilOverride(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	// no panic expected
	src := make(map[interface{}]interface{})
	overrideMap(src, nil)
}

func TestOverrideMap_WithUnMatchedType(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	// no panic expected
	src := make(map[interface{}]interface{})
	src["ut-src-key"] = "ut-src-value"

	override := make(map[interface{}]interface{})
	override["ut-override-key"] = false

	overrideMap(src, override)

	// source map should keep the same
	assert.Equal(t, 1, len(src))
	assert.Equal(t, "ut-src-value", src["ut-src-key"])

	// override map should never change
	assert.Equal(t, 1, len(override))
	assert.Equal(t, false, override["ut-override-key"])
}

func TestOverrideMap_WithMixedType(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	// no panic expected
	src := make(map[interface{}]interface{})
	src["ut-src-key"] = "ut-src-value"

	override := make(map[interface{}]interface{})
	override["ut-override-key"] = false
	override["ut-src-key"] = "ut-override-value"

	overrideMap(src, override)

	// source map should be changed
	assert.Equal(t, 1, len(src))
	assert.Equal(t, "ut-override-value", src["ut-src-key"])

	// override map should never change
	assert.Equal(t, 2, len(override))
	assert.Equal(t, false, override["ut-override-key"])
	assert.Equal(t, "ut-override-value", src["ut-src-key"])
}

func TestOverrideMap_WithHappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	type MyStruct struct {
		Key string
	}

	// no panic expected
	src := make(map[interface{}]interface{})
	src["ut-str"] = ""
	src["ut-list"] = []string{}
	src["ut-map"] = map[string]interface{}{}
	src["ut-struct"] = &MyStruct{}

	override := make(map[interface{}]interface{})
	override["ut-str"] = "ut-str"
	override["ut-list"] = []string{"one", "two"}
	override["ut-map"] = map[string]interface{}{
		"key": "value",
	}
	override["ut-struct"] = &MyStruct{
		Key: "override",
	}

	overrideMap(src, override)

	// source map should be changed
	// validate string
	assert.Equal(t, "ut-str", src["ut-str"])
	// validate list
	assert.Contains(t, src["ut-list"], "one")
	assert.Contains(t, src["ut-list"], "two")
	// validate map
	innerMap := src["ut-map"]
	assert.Equal(t, "value", innerMap.(map[string]interface{})["key"])
	// validate struct
	innerStruct := src["ut-struct"]
	assert.NotNil(t, "override", innerStruct.(*MyStruct).Key)
}

func TestOverrideMap_WithGenericType(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// this should never be called in case of a bug
			assert.True(t, false)
		} else {
			// no panic expected
			assert.True(t, true)
		}
	}()

	type MyStruct struct {
		Key string
	}

	// no panic expected
	src := make(map[interface{}]interface{})
	src["ut-generic-list"] = []*MyStruct{}
	src["ut-generic-map"] = map[string]*MyStruct{}

	override := make(map[interface{}]interface{})
	override["ut-generic-list"] = []*MyStruct{
		{Key: "key"},
	}
	override["ut-generic-map"] = map[string]*MyStruct{
		"map-key": {Key: "key"},
	}

	overrideMap(src, override)

	// source map should be changed
	assert.Equal(t, override["ut-generic-list"], src["ut-generic-list"])
	assert.Equal(t, override["ut-generic-map"], src["ut-generic-map"])
}

func TestIsLocaleValid_WithEmptyLocale(t *testing.T) {
	assert.False(t, IsLocaleValid(""))
}

func TestIsLocaleValid_WithInvalidLocale(t *testing.T) {
	assert.False(t, IsLocaleValid("realm::region::az"))
}

func TestIsLocaleValid_WithEmptyRealmEnv(t *testing.T) {
	// with realm exist in locale
	assert.False(t, IsLocaleValid("fake-realm::*::*::*"))

	// with wildcard in realm
	assert.True(t, IsLocaleValid("*::*::*::*"))
}

func TestIsLocaleValid_WithRealmEnv(t *testing.T) {
	// set environment variable
	assert.Nil(t, os.Setenv("REALM", "ut"))

	// with realm exist in locale
	assert.True(t, IsLocaleValid("ut::*::*::*"))

	// with wildcard in realm
	assert.True(t, IsLocaleValid("*::*::*::*"))

	// with wrong realm
	assert.False(t, IsLocaleValid("rk::*::*::*"))

	assert.Nil(t, os.Setenv("REALM", ""))
}

func TestIsLocaleValid_WithRegionEnv(t *testing.T) {
	// set environment variable
	assert.Nil(t, os.Setenv("REGION", "ut"))

	// with region exist in locale
	assert.True(t, IsLocaleValid("*::ut::*::*"))

	// with wildcard in region
	assert.True(t, IsLocaleValid("*::*::*::*"))

	// with wrong region
	assert.False(t, IsLocaleValid("*::rk::*::*"))

	assert.Nil(t, os.Setenv("REGION", ""))
}

func TestIsLocaleValid_WithAZEnv(t *testing.T) {
	// set environment variable
	assert.Nil(t, os.Setenv("AZ", "ut"))

	// with az exist in locale
	assert.True(t, IsLocaleValid("*::*::ut::*"))

	// with wildcard in az
	assert.True(t, IsLocaleValid("*::*::*::*"))

	// with wrong az
	assert.False(t, IsLocaleValid("*::*::rk::*"))

	assert.Nil(t, os.Setenv("AZ", ""))
}

func TestIsLocaleValid_WithDomainEnv(t *testing.T) {
	// set environment variable
	assert.Nil(t, os.Setenv("DOMAIN", "ut"))

	// with domain exist in locale
	assert.True(t, IsLocaleValid("*::*::*::ut"))

	// with wildcard in domain
	assert.True(t, IsLocaleValid("*::*::*::*"))

	// with wrong domain
	assert.False(t, IsLocaleValid("*::*::*::rk"))

	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestShutdownWithError_WithNilError(t *testing.T) {
	defer assertPanic(t)
	ShutdownWithError(nil)
}

func TestShutdownWithError_HappyCase(t *testing.T) {
	defer assertPanic(t)
	ShutdownWithError(errors.New("error from unit test"))
}

func TestParseEnvOverrides(t *testing.T) {
	assert.Nil(t, os.Setenv("RK_GIN_NAME", "rookie"))

	m, err := parseEnvOverrides("rk")
	assert.Nil(t, err)
	assert.Len(t, m, 1)
	for k, v := range m {
		assert.Equal(t, "gin", k)
		for k1, v1 := range v.(map[interface{}]interface{}) {
			assert.Equal(t, "name", k1)
			assert.Equal(t, "rookie", v1)
		}
	}

	assert.Nil(t, os.Setenv("RK_GIN_NAME", ""))
}
