package rkentry

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
)

// UnmarshalBoot this function will parse boot config file with ENV and pflag overrides.
//
// User who want to implement his/her own entry, may use this function to parse YAML config into struct.
// This function would also parse --rkset flags.
//
// This function would do the following:
// 1: Read config file and unmarshal content into a map.
// 2: Read --rkset flags and override values in map unmarshalled at above step.
// 3: Unmarshal map into user provided struct.
//
//
// [Flag]: Override boot config value with flag of rkset:
//
// pflag.FlagSet which contains rkset as key.
//
// Receives flattened boot config file(YAML) keys and override them in provided boot config.
// We follow the way of HELM does while overriding keys, refer to https://helm.sh/docs/intro/using_helm/
// example:
//
// Lets assuming we have boot config YAML file as bellow:
//
// example-boot.yaml:
// gin:
//   - port: 1949
//     commonService:
//       enabled: true
//
// We can override values in example-boot.yaml file as bellow:
//
// ./your_compiled_binary --rkset "gin[0].port=2008,gin[0].commonService.enabled=false"
//
// Basic rules:
// 1: Using comma(,) to separate different k/v section.
// 2: Using [index] to access arrays in YAML file.
// 3: Using equal sign(=) to distinguish key and value.
// 4: Using dot(.) to access map in YAML file.
//
// [Environment variable]: Override boot config value
//
// Prefix of "RK" will be used as environment variable key. The schema follows above.
//
// example-boot.yaml:
// gin:
//   - port: 1949
//     commonService:
//       enabled: true
//
// We can override values in example-boot.yaml file as bellow:
//
// os.Setenv("RK_GIN[0]_PORT", "2008")
// os.Setenv("RK_GIN[0]_COMMONSERVICE_ENABLED", "false")
//
// ./your_compiled_binary
func UnmarshalBoot(raw []byte, config interface{}) {
	// 1: unmarshal original
	originalBootM := map[interface{}]interface{}{}
	vp := viper.New()
	vp.SetConfigType("yaml")
	vp.ReadConfig(bytes.NewReader(raw))
	if err := vp.Unmarshal(&originalBootM); err != nil {
		ShutdownWithError(err)
	}

	// 2: get ENV overrides
	envOverridesBootM, err := parseEnvOverrides("RK")
	if err != nil {
		ShutdownWithError(err)
	}

	// 3: get flag overrides
	pFlag := pflag.NewFlagSet("rk", pflag.ContinueOnError)
	pFlag.String("rkset", "", "")
	flagOverridesBootM, err := parseFlagOverrides(pFlag)
	if err != nil {
		ShutdownWithError(err)
	}

	// 4: override environment first, and then flags
	overrideMap(originalBootM, envOverridesBootM)
	overrideMap(originalBootM, flagOverridesBootM)

	// 5: unmarshal to struct
	if err := mapstructure.Decode(originalBootM, config); err != nil {
		ShutdownWithError(err)
	}

	GlobalAppCtx.bootConfig = config
}

// ShutdownWithError shuts down and panic.
func ShutdownWithError(err error) {
	panic(err)
}

// IsLocaleValid mainly used in entry config.
// RK use <realm>::<region>::<az>::<domain> to distinguish different environment.
// Variable of <locale> could be composed as form of <realm>::<region>::<az>::<domain>
// - realm: It could be a company, department and so on, like RK-Corp.
//          Environment variable: REALM
//          Eg: RK-Corp
//          Wildcard: supported
//
// - region: Please see AWS web site:
//   https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
//           Environment variable: REGION
//           Eg: us-east
//           Wildcard: supported
//
// - az: Availability zone, please see AWS web site for details:
//  https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html
//       Environment variable: AZ
//       Eg: us-east-1
//       Wildcard: supported
//
// - domain: Stands for different environment, like dev, test, prod and so on, users can define it by themselves.
//           Environment variable: DOMAIN
//           Eg: prod
//           Wildcard: supported
//
// How it works?
// First, we will split locale with "::" and extract realm, region, az and domain.
// Second, get environment variable named as REALM, REGION, AZ and DOMAIN.
// Finally, compare every element in locale variable and environment variable.
// If variables in locale represented as wildcard(*), we will ignore comparison step.
//
// Example:
// # let's assuming we are going to define DB address which is different based on environment.
// # Then, user can distinguish DB address based on locale.
// # We recommend to include locale with wildcard.
// ---
// DB:
//   - name: redis-default
//     locale: "*::*::*::*"
//     addr: "192.0.0.1:6379"
//   - name: redis-in-test
//     locale: "*::*::*::test"
//     addr: "192.0.0.1:6379"
//   - name: redis-in-prod
//     locale: "*::*::*::prod"
//     addr: "176.0.0.1:6379"
func IsLocaleValid(locale string) bool {
	if len(locale) < 1 {
		return false
	}

	tokens := strings.Split(locale, "::")
	if len(tokens) != 4 {
		return false
	}

	realmFromEnv := getDefaultIfEmptyString(os.Getenv("REALM"), "*")
	regionFromEnv := getDefaultIfEmptyString(os.Getenv("REGION"), "*")
	azFromEnv := getDefaultIfEmptyString(os.Getenv("AZ"), "*")
	domainFromEnv := getDefaultIfEmptyString(os.Getenv("DOMAIN"), "*")

	if tokens[0] != "*" && realmFromEnv != "*" {
		if tokens[0] != realmFromEnv {
			return false
		}
	}

	if tokens[1] != "*" && regionFromEnv != "*" {
		if tokens[1] != regionFromEnv {
			return false
		}
	}

	if tokens[2] != "*" && azFromEnv != "*" {
		if tokens[0] != azFromEnv {
			return false
		}
	}

	if tokens[3] != "*" && domainFromEnv != "*" {
		if tokens[3] != domainFromEnv {
			return false
		}
	}

	return true
}

// readFile wil read try to read file with bellow sequence.
//
// 1: Read from embed.FS if not nil
// 2: Read from local FS
func readFile(filePath string, fs *embed.FS) []byte {
	if fs != nil {
		data, err := fs.ReadFile(filePath)
		if err != nil {
			ShutdownWithError(err)
		}
		return data
	}

	wd, _ := os.Getwd()

	if !path.IsAbs(filePath) {
		filePath = path.Join(wd, filePath)
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		ShutdownWithError(err)
	}
	return data
}

// parseBootOverrides parses a set line.
//
// A set line is of the form name1=value1,name2=value2
func parseBootOverrides(s string) (map[interface{}]interface{}, error) {
	vals := map[interface{}]interface{}{}
	scanner := bytes.NewBufferString(s)
	t := newParser(scanner, vals, false)
	err := t.parse()
	return vals, err
}

// overrideMap override source map with new map items.
// It will iterate through all items in map and check map and slice types of item to recursively override values
//
// Mainly used for unmarshalling YAML to map.
func overrideMap(src map[interface{}]interface{}, override map[interface{}]interface{}) {
	if src == nil || override == nil {
		return
	}

	for k, overrideItem := range override {
		originalItem, ok := src[k]
		if ok && reflect.TypeOf(originalItem) == reflect.TypeOf(overrideItem) {
			switch overrideItem.(type) {
			case []interface{}:
				overrideSlice(originalItem.([]interface{}), overrideItem.([]interface{}))
			case map[interface{}]interface{}:
				overrideMap(originalItem.(map[interface{}]interface{}), overrideItem.(map[interface{}]interface{}))
			default:
				src[k] = overrideItem
			}
		}
	}
}

// overrideSlice override source slice with new slice items.
// It will iterate through all items in slice and check map and slice types of item to recursively override values
//
// Mainly used for unmarshalling YAML to map.
func overrideSlice(src []interface{}, override []interface{}) {
	if src == nil || override == nil {
		return
	}

	for i := range override {
		if override[i] != nil && len(src)-1 >= i && reflect.TypeOf(override[i]) == reflect.TypeOf(src[i]) {
			overrideItem := override[i]
			originalItem := src[i]
			switch overrideItem.(type) {
			case []interface{}:
				overrideSlice(originalItem.([]interface{}), overrideItem.([]interface{}))
			case map[interface{}]interface{}:
				overrideMap(originalItem.(map[interface{}]interface{}), overrideItem.(map[interface{}]interface{}))
			default:
				src[i] = override[i]
			}
		}
	}
}

// parseEnvOverrides read environment variables and convert to map
func parseEnvOverrides(prefix string) (map[interface{}]interface{}, error) {
	overrideValueList := make([]string, 0)

	// 1: iterate ENV values and filter with prefix
	for _, val := range os.Environ() {
		if !strings.HasPrefix(val, strings.ToUpper(prefix)+"_") {
			continue
		}

		tokens := strings.SplitN(val, "=", 2)
		if len(tokens) != 2 {
			continue
		}

		// convert key
		newKey := strings.ToLower(strings.ReplaceAll(strings.TrimPrefix(tokens[0], strings.ToUpper(prefix)+"_"), "_", "."))
		newValue := tokens[1]

		overrideValueList = append(overrideValueList, fmt.Sprintf("%s=%s", newKey, newValue))
	}

	// 2: flatten values
	overrideValueFlatten := strings.Join(overrideValueList, ",")

	// 3: parse to map
	return parseBootOverrides(overrideValueFlatten)
}

// parseEnvOverrides read flag values and convert to map
func parseFlagOverrides(set *pflag.FlagSet) (map[interface{}]interface{}, error) {
	overrideValueList := make([]string, 0)

	// 1: iterate pFlag values and filter with prefix
	set.ParseAll(os.Args[1:], func(flag *pflag.Flag, value string) error {
		overrideValueList = append(overrideValueList, value)
		return nil
	})

	// 2: flatten values
	overrideValueFlatten := strings.Join(overrideValueList, ",")

	// 3: parse to map
	return parseBootOverrides(overrideValueFlatten)
}

// overrideLumberjackConfig override lumberjack config.
// This function will override fields of non empty and non-nil.
func overrideLumberjackConfig(origin *lumberjack.Logger, override *lumberjack.Logger) {
	if override == nil {
		return
	}
	origin.Compress = override.Compress
	origin.LocalTime = override.LocalTime
	if override.MaxAge > 0 {
		origin.MaxAge = override.MaxAge
	}

	if override.MaxBackups > 0 {
		origin.MaxBackups = override.MaxBackups
	}

	if override.MaxSize > 0 {
		origin.MaxSize = override.MaxSize
	}

	if len(override.Filename) > 0 {
		origin.Filename = override.Filename
	}
}

// overrideZapConfig overrides zap config.
// This function will override fields of non empty and non-nil.
func overrideZapConfig(origin *zap.Config, override *zap.Config) {
	if override == nil {
		return
	}

	// by default, these fields would be false
	// so just override it with new config
	origin.Development = override.Development
	origin.DisableCaller = override.DisableCaller
	origin.DisableStacktrace = override.DisableStacktrace

	if len(override.Encoding) > 0 {
		origin.Encoding = override.Encoding
	}

	if !reflect.ValueOf(override.Level).Field(0).IsNil() {
		origin.Level.SetLevel(override.Level.Level())
	}

	if len(override.InitialFields) > 0 {
		origin.InitialFields = override.InitialFields
	}

	if len(override.ErrorOutputPaths) > 0 {
		origin.ErrorOutputPaths = override.ErrorOutputPaths
	}

	if len(override.OutputPaths) > 0 {
		origin.OutputPaths = override.OutputPaths
	}

	if override.Sampling != nil {
		origin.Sampling = override.Sampling
	}

	// deal with encoder config
	if len(override.EncoderConfig.CallerKey) > 0 {
		origin.EncoderConfig.CallerKey = override.EncoderConfig.CallerKey
	}

	if len(override.EncoderConfig.ConsoleSeparator) > 0 {
		origin.EncoderConfig.ConsoleSeparator = override.EncoderConfig.ConsoleSeparator
	}

	if override.EncoderConfig.EncodeCaller != nil {
		origin.EncoderConfig.EncodeCaller = override.EncoderConfig.EncodeCaller
	}

	if override.EncoderConfig.EncodeDuration != nil {
		origin.EncoderConfig.EncodeDuration = override.EncoderConfig.EncodeDuration
	}

	if override.EncoderConfig.EncodeLevel != nil {
		origin.EncoderConfig.EncodeLevel = override.EncoderConfig.EncodeLevel
	}

	if override.EncoderConfig.EncodeName != nil {
		origin.EncoderConfig.EncodeName = override.EncoderConfig.EncodeName
	}

	if override.EncoderConfig.EncodeTime != nil {
		origin.EncoderConfig.EncodeTime = override.EncoderConfig.EncodeTime
	}

	if len(override.EncoderConfig.MessageKey) > 0 {
		origin.EncoderConfig.MessageKey = override.EncoderConfig.MessageKey
	}

	if len(override.EncoderConfig.LevelKey) > 0 {
		origin.EncoderConfig.LevelKey = override.EncoderConfig.LevelKey
	}

	if len(override.EncoderConfig.TimeKey) > 0 {
		origin.EncoderConfig.TimeKey = override.EncoderConfig.TimeKey
	}

	if len(override.EncoderConfig.NameKey) > 0 {
		origin.EncoderConfig.NameKey = override.EncoderConfig.NameKey
	}

	if len(override.EncoderConfig.FunctionKey) > 0 {
		origin.EncoderConfig.FunctionKey = override.EncoderConfig.FunctionKey
	}

	if len(override.EncoderConfig.StacktraceKey) > 0 {
		origin.EncoderConfig.StacktraceKey = override.EncoderConfig.StacktraceKey
	}

	if len(override.EncoderConfig.LineEnding) > 0 {
		origin.EncoderConfig.LineEnding = override.EncoderConfig.LineEnding
	}
}

// getDefaultIfEmptyString returns default value if original string is empty.
func getDefaultIfEmptyString(origin, def string) string {
	if len(origin) < 1 {
		return def
	}

	return origin
}

// fileExists checks File existence, file path should be full path.
func fileExists(filePath string) bool {
	if file, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	} else if file.IsDir() {
		return false
	}
	return true
}
