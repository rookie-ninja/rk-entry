// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/consul/api"
	rkcommon "github.com/rookie-ninja/rk-common/common"
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const (
	// CredEntryName name of entry
	CredEntryName = "CredDefault"
	// CredEntryType type of entry
	CredEntryType = "CredEntry"
	// CredEntryDescription description of entry
	CredEntryDescription = "Internal RK entry which retrieves credentials from localFs, remoteFs, etcd or consul."
	// DefaultTimeout default timeout while connecting to retriever
	DefaultTimeout = 3 * time.Second
	// ProviderEtcd retriever type of etcd
	ProviderEtcd = "etcd"
	// ProviderEtcd retriever type of consul
	ProviderConsul = "consul"
	// ProviderEtcd retriever type of localFs
	ProviderLocalFs = "localFs"
	// ProviderEtcd retriever type of RemoteFs
	ProviderRemoteFs = "remoteFs"
)

// BootConfigCred defines bootstrapper config
type BootConfigCred struct {
	Cred []struct {
		Name        string   `yaml:"name" json:"name"`
		Description string   `yaml:"description" json:"description"`
		Provider    string   `yaml:"provider" json:"provider"`
		Locale      string   `yaml:"locale" json:"locale"`
		Endpoint    string   `yaml:"endpoint" json:"endpoint"`
		Datacenter  string   `yaml:"datacenter" json:"datacenter"`
		Token       string   `yaml:"token" json:"token"`
		BasicAuth   string   `yaml:"basicAuth" json:"basicAuth"`
		Paths       []string `yaml:"paths" json:"paths"`
		Logger      struct {
			ZapLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"zapLogger" json:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"cred" json:"cred"`
}

// CredStore is storage stores credentials retrieve via retriever
type CredStore struct {
	Cred map[string][]byte `json:"-" yaml:"-"`
}

// GetCred returns credential with name
func (store *CredStore) GetCred(path string) []byte {
	return store.Cred[path]
}

// MarshalJSON marshal entry
func (store *CredStore) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}

	for k := range store.Cred {
		m[k] = "Sensitive data!"
	}

	return json.Marshal(&m)
}

// UnmarshalJSON unmarshal entry
func (store *CredStore) UnmarshalJSON([]byte) error {
	return nil
}

// CredEntry defines credential entry
type CredEntry struct {
	EntryName        string            `json:"entryName" yaml:"entryName"`
	EntryType        string            `json:"entryType" yaml:"entryType"`
	EntryDescription string            `json:"entryDescription" yaml:"entryDescription"`
	ZapLoggerEntry   *ZapLoggerEntry   `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry `json:"-" yaml:"-"`
	Store            *CredStore        `json:"store" yaml:"store"`
	Retriever        Retriever         `json:"retriever" yaml:"retriever"`
}

// CredEntryOption Option which used while registering entry from codes.
type CredEntryOption func(entry *CredEntry)

// WithNameCred provide name.
func WithNameCred(name string) CredEntryOption {
	return func(entry *CredEntry) {
		entry.EntryName = name
	}
}

// WithDescriptionCred provide description.
func WithDescriptionCred(description string) CredEntryOption {
	return func(entry *CredEntry) {
		entry.EntryDescription = description
	}
}

// WithZapLoggerEntryCred provide ZapLoggerEntry.
func WithZapLoggerEntryCred(logger *ZapLoggerEntry) CredEntryOption {
	return func(entry *CredEntry) {
		if logger != nil {
			entry.ZapLoggerEntry = logger
		}
	}
}

// WithEventLoggerEntryCred provide EventLoggerEntry.
func WithEventLoggerEntryCred(logger *EventLoggerEntry) CredEntryOption {
	return func(entry *CredEntry) {
		if logger != nil {
			entry.EventLoggerEntry = logger
		}
	}
}

// WithRetrieverCred provide Retriever.
func WithRetrieverCred(retriever Retriever) CredEntryOption {
	return func(entry *CredEntry) {
		entry.Retriever = retriever
	}
}

// RegisterCredEntriesFromConfig implements rkentry.EntryRegFunc which generate Entry based on boot configuration file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterCredEntriesFromConfig(configFilePath string) map[string]Entry {
	config := &BootConfigCred{}

	rkcommon.UnmarshalBootConfig(configFilePath, config)

	res := make(map[string]Entry)

	for i := range config.Cred {
		element := config.Cred[i]

		if len(element.Name) < 1 || !rkcommon.MatchLocaleWithEnv(element.Locale) {
			continue
		}

		zapLoggerEntry := GlobalAppCtx.GetZapLoggerEntry(element.Logger.ZapLogger.Ref)
		if zapLoggerEntry == nil {
			zapLoggerEntry = GlobalAppCtx.GetZapLoggerEntryDefault()
		}

		eventLoggerEntry := GlobalAppCtx.GetEventLoggerEntry(element.Logger.EventLogger.Ref)
		if eventLoggerEntry == nil {
			eventLoggerEntry = GlobalAppCtx.GetEventLoggerEntryDefault()
		}

		var retriever Retriever

		switch element.Provider {
		case ProviderConsul:
			retriever = &CredRetrieverConsul{
				Provider:         element.Provider,
				Locale:           element.Locale,
				Endpoint:         element.Endpoint,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Datacenter:       element.Datacenter,
				Token:            element.Token,
				BasicAuth:        element.BasicAuth,
				Paths:            element.Paths,
			}
		case ProviderEtcd:
			retriever = &CredRetrieverEtcd{
				Provider:         element.Provider,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Locale:           element.Locale,
				Endpoint:         element.Endpoint,
				BasicAuth:        element.BasicAuth,
				Paths:            element.Paths,
			}
		case ProviderLocalFs:
			retriever = &CredRetrieverLocalFs{
				Provider:         element.Provider,
				Locale:           element.Locale,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Paths:            element.Paths,
			}
		case ProviderRemoteFs:
			retriever = &CredRetrieverRemoteFs{
				Provider:         element.Provider,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Locale:           element.Locale,
				Endpoint:         element.Endpoint,
				BasicAuth:        element.BasicAuth,
				Paths:            element.Paths,
			}
		}

		entry := RegisterCredEntry(
			WithNameCred(element.Name),
			WithDescriptionCred(element.Description),
			WithZapLoggerEntryCred(zapLoggerEntry),
			WithEventLoggerEntryCred(eventLoggerEntry),
			WithRetrieverCred(retriever))

		res[entry.GetName()] = entry
	}

	return res
}

// RegisterCredEntry create cred entry with options.
func RegisterCredEntry(opts ...CredEntryOption) *CredEntry {
	entry := &CredEntry{
		EventLoggerEntry: GlobalAppCtx.GetEventLoggerEntryDefault(),
		ZapLoggerEntry:   GlobalAppCtx.GetZapLoggerEntryDefault(),
		EntryName:        CredEntryName,
		EntryType:        CredEntryType,
		EntryDescription: CredEntryDescription,
		Store: &CredStore{
			Cred: make(map[string][]byte, 0),
		},
	}

	for i := range opts {
		opts[i](entry)
	}

	GlobalAppCtx.AddCredEntry(entry)

	return entry
}

// Bootstrap iterate retrievers and call Retrieve() for each of them.
func (entry *CredEntry) Bootstrap(ctx context.Context) {
	entry.Store = entry.Retriever.Retrieve(ctx)
}

// Interrupt entry.
func (entry *CredEntry) Interrupt(context.Context) {
	// no op
}

// String return string of entry.
func (entry *CredEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON marshal entry
func (entry *CredEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
		"store":            entry.Store,
		"retriever":        entry.Retriever,
	}

	return json.Marshal(&m)
}

// UnmarshalJSON unmarshal entry
func (entry *CredEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetName returns name of entry.
func (entry *CredEntry) GetName() string {
	return entry.EntryName
}

// GetType returns type of entry.
func (entry *CredEntry) GetType() string {
	return entry.EntryType
}

// GetDescription returns description of entry
func (entry *CredEntry) GetDescription() string {
	return entry.EntryDescription
}

// Retriever is an interface for retrieving credentials.
type Retriever interface {
	// Read credential files into byte array and store it into CredStore.
	Retrieve(context.Context) *CredStore

	// Return provider of retriever.
	GetProvider() string

	// Return list of paths
	ListPaths() []string

	// Return endpoint.
	GetEndpoint() string

	// Return locale.
	GetLocale() string
}

// ******************************
// ************ etcd ************
// ******************************

// CredRetrieverEtcd is retriever read from ETCD
// 1: Name: Name of section, required.
// 2: Provider: Provider of retriever, required.
// 3: Locale: <realm>::<region>::<az>::<domain>
// 4: Endpoint: Endpoint of ETCD server, http://x.x.x.x or x.x.x.x both acceptable.
// 5: BasicAuth: Basic auth for ETCD server, like <user:pass>.
// 6: Paths: Key of value needs to retrieve from ETCD server.
type CredRetrieverEtcd struct {
	Name             string            `yaml:"name" json:"name"`
	Provider         string            `yaml:"provider" json:"provider"`
	Locale           string            `yaml:"locale" json:"locale"`
	ZapLoggerEntry   *ZapLoggerEntry   `yaml:"-" json:"-"`
	EventLoggerEntry *EventLoggerEntry `yaml:"-" json:"-"`
	Endpoint         string            `yaml:"endpoint" json:"endpoint"`
	BasicAuth        string            `yaml:"-" json:"-"`
	Paths            []string          `yaml:"paths" json:"paths"`
}

// Retrieve call ETCD server and retrieve values based on keys.
func (retriever *CredRetrieverEtcd) Retrieve(context.Context) *CredStore {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{retriever.Endpoint},
		DialTimeout: DefaultTimeout,
		LogConfig:   retriever.ZapLoggerEntry.GetLoggerConfig(),
		Username:    rkcommon.GetUsernameFromBasicAuthString(retriever.BasicAuth),
		Password:    rkcommon.GetPasswordFromBasicAuthString(retriever.BasicAuth),
	})

	store := &CredStore{
		Cred: make(map[string][]byte, 0),
	}

	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to create etcd client v3",
			zap.Error(err))
		return store
	}

	defer client.Close()

	for i := range retriever.Paths {
		path := retriever.Paths[i]
		store.Cred[path] = retriever.getValueFromEtcd(client, path)
	}

	return store
}

// Inner utility function.
func (retriever *CredRetrieverEtcd) getValueFromEtcd(client *clientv3.Client, key string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	if resp, err := client.Get(ctx, key); err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to get credentials from etcd",
			zap.String("endpoint", retriever.Endpoint),
			zap.String("locale", retriever.Locale),
			zap.String("key", key),
			zap.Error(err))
		return nil
	} else {
		if len(resp.Kvs) > 0 {
			return resp.Kvs[0].Value
		}

		return nil
	}
}

// ListPaths return list of paths
func (retriever *CredRetrieverEtcd) ListPaths() []string {
	return retriever.Paths
}

// GetProvider returns provider of retriever.
func (retriever *CredRetrieverEtcd) GetProvider() string {
	return retriever.Provider
}

// GetEndpoint returns endpoint.
func (retriever *CredRetrieverEtcd) GetEndpoint() string {
	return retriever.Endpoint
}

// GetLocale returns locale.
func (retriever *CredRetrieverEtcd) GetLocale() string {
	return retriever.Locale
}

// ********************************
// ************ consul ************
// ********************************

// CredRetrieverConsul is retriever read from Consul
// 1: Provider: Provider of retriever, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: Endpoint: Endpoint of consul server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: Datacenter: Consul datacenter.
// 5: Token: Token for access Consul.
// 6: BasicAuth: Basic auth for Consul server, like <user:pass>.
// 7: Paths: Key of value needs to retrieve from Consul server.
type CredRetrieverConsul struct {
	Provider         string            `yaml:"provider" json:"provider"`
	ZapLoggerEntry   *ZapLoggerEntry   `yaml:"-" json:"-"`
	EventLoggerEntry *EventLoggerEntry `yaml:"-" json:"-"`
	Locale           string            `yaml:"locale" json:"locale"`
	Endpoint         string            `yaml:"endpoint" json:"endpoint"`
	Datacenter       string            `yaml:"datacenter" json:"datacenter"`
	Token            string            `yaml:"-" json:"-"`
	BasicAuth        string            `yaml:"-" json:"-"`
	Paths            []string          `yaml:"paths" json:"paths"`
}

// Retrieve call Consul server/agent and retrieve values based on keys.
func (retriever *CredRetrieverConsul) Retrieve(context.Context) *CredStore {
	scheme := rkcommon.ExtractSchemeFromURL(retriever.Endpoint)
	endpoint := retriever.Endpoint

	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.Trim(endpoint, "http://")
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.Trim(endpoint, "https://")
	}

	config := &api.Config{
		Address:    retriever.Endpoint,
		Datacenter: retriever.Datacenter,
		Token:      retriever.Token,
		Scheme:     scheme,
	}

	if len(retriever.BasicAuth) > 0 {
		auth := &api.HttpBasicAuth{
			Username: rkcommon.GetUsernameFromBasicAuthString(retriever.BasicAuth),
			Password: rkcommon.GetPasswordFromBasicAuthString(retriever.BasicAuth),
		}

		config.HttpAuth = auth
	}

	store := &CredStore{
		Cred: make(map[string][]byte, 0),
	}

	// Get a new client
	client, err := api.NewClient(config)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to create consul client v3",
			zap.Error(err))
		return store
	}

	for i := range retriever.Paths {
		path := retriever.Paths[i]
		store.Cred[path] = retriever.getValueFromConsul(client, path)
	}

	return store
}

// ListPaths return list of paths
func (retriever *CredRetrieverConsul) ListPaths() []string {
	return retriever.Paths
}

// GetProvider return provider of retriever.
func (retriever *CredRetrieverConsul) GetProvider() string {
	return retriever.Provider
}

// GetEndpoint return endpoint.
func (retriever *CredRetrieverConsul) GetEndpoint() string {
	return retriever.Endpoint
}

// GetLocale return locale.
func (retriever *CredRetrieverConsul) GetLocale() string {
	return retriever.Locale
}

// Inner utility function.
func (retriever *CredRetrieverConsul) getValueFromConsul(client *api.Client, key string) []byte {
	// Get a handle to the KV API
	kv := client.KV()

	// Lookup the pair
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to get credential from consul",
			zap.String("endpoint", retriever.Endpoint),
			zap.String("locale", retriever.Locale),
			zap.String("key", key),
			zap.Error(err))
		return nil
	}

	return pair.Value
}

// *********************************
// ************ localFs ************
// *********************************

// CredRetrieverLocalFs is retriever read from local file system
// 1: Provider: Type of retriever, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: Paths: Key of value need to retrieve from localFs.
type CredRetrieverLocalFs struct {
	Provider         string            `yaml:"provider" json:"provider"`
	Locale           string            `yaml:"locale" json:"locale"`
	ZapLoggerEntry   *ZapLoggerEntry   `yaml:"-" json:"-"`
	EventLoggerEntry *EventLoggerEntry `yaml:"-" json:"-"`
	Paths            []string          `yaml:"paths" json:"paths"`
}

// Retrieve read files from local file system and retrieve values based on keys.
func (retriever *CredRetrieverLocalFs) Retrieve(context.Context) *CredStore {
	wd, err := os.Getwd()

	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to get working directory", zap.Error(err))
		return nil
	}

	store := &CredStore{
		Cred: make(map[string][]byte, 0),
	}

	for i := range retriever.Paths {
		localPath := retriever.Paths[i]
		if len(localPath) < 1 {
			continue
		}

		if len(localPath) > 0 && !path.IsAbs(localPath) {
			localPath = path.Join(wd, retriever.Paths[i])
		}

		// Read files from local
		value, err := ioutil.ReadFile(localPath)
		if err != nil {
			retriever.ZapLoggerEntry.GetLogger().Warn("failed to read credential from localFs",
				zap.Error(err),
				zap.String("path", localPath))
			continue
		}

		store.Cred[retriever.Paths[i]] = value
	}

	return store
}

// ListPaths return list of paths
func (retriever *CredRetrieverLocalFs) ListPaths() []string {
	return retriever.Paths
}

// GetProvider returns provider of retriever.
func (retriever *CredRetrieverLocalFs) GetProvider() string {
	return retriever.Provider
}

// GetEndpoint returns endpoint.
func (retriever *CredRetrieverLocalFs) GetEndpoint() string {
	return "local"
}

// GetLocale returns locale.
func (retriever *CredRetrieverLocalFs) GetLocale() string {
	return retriever.Locale
}

// **********************************
// ************ remoteFs ************
// **********************************

// CredRetrieverRemoteFs is retriever read from remote file system
// 1: Provider: Provider of retriever, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: Endpoint: Endpoint of RemoteFileStore server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: BasicAuth: Basic auth for RemoteFileStore server, like <user:pass>.
// 5: Paths: Key of value need to retrieved form remote FS.
type CredRetrieverRemoteFs struct {
	Provider         string            `yaml:"provider" json:"provider"`
	ZapLoggerEntry   *ZapLoggerEntry   `yaml:"-" json:"-"`
	EventLoggerEntry *EventLoggerEntry `yaml:"-" json:"-"`
	Locale           string            `yaml:"locale" json:"locale"`
	Endpoint         string            `yaml:"endpoint" json:"endpoint"`
	BasicAuth        string            `yaml:"-" json:"-"`
	Paths            []string          `yaml:"paths" json:"paths"`
}

// CredRetrieverRemoteFs call remote file store and retrieve values based on keys.
func (retriever *CredRetrieverRemoteFs) Retrieve(context.Context) *CredStore {
	client := &http.Client{
		Timeout: DefaultTimeout,
	}

	store := &CredStore{
		Cred: make(map[string][]byte, 0),
	}

	for i := range retriever.Paths {
		path := retriever.Paths[i]
		store.Cred[path] = retriever.getValueFromRemoteFs(client, path)
	}

	return store
}

// Inner utility function.
func (retriever *CredRetrieverRemoteFs) getValueFromRemoteFs(client *http.Client, remotePath string) []byte {
	if !strings.HasPrefix(retriever.Endpoint, "http://") {
		retriever.Endpoint = "http://" + retriever.Endpoint
	}

	if !strings.HasPrefix(remotePath, "/") {
		remotePath = "/" + remotePath
	}

	req, err := http.NewRequest("GET", retriever.Endpoint+remotePath, nil)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed create http request",
			zap.String("endpoint", retriever.Endpoint),
			zap.String("locale", retriever.Locale),
			zap.String("remotePath", remotePath),
			zap.Error(err))
		return nil
	}

	username := rkcommon.GetUsernameFromBasicAuthString(retriever.BasicAuth)
	password := rkcommon.GetPasswordFromBasicAuthString(retriever.BasicAuth)
	if len(username) > 0 && len(password) > 0 {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to get credential from remote file store",
			zap.String("endpoint", retriever.Endpoint),
			zap.String("locale", retriever.Locale),
			zap.String("remotePath", remotePath),
			zap.Error(err))
		return nil
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to read credential from remote file store",
			zap.String("endpoint", retriever.Endpoint),
			zap.String("locale", retriever.Locale),
			zap.String("remotePath", remotePath),
			zap.Error(err))
		return nil
	}

	return res
}

// ListPaths return list of paths
func (retriever *CredRetrieverRemoteFs) ListPaths() []string {
	return retriever.Paths
}

// GetProvider returns provider of retriever.
func (retriever *CredRetrieverRemoteFs) GetProvider() string {
	return retriever.Provider
}

// GetEndpoint returns endpoint.
func (retriever *CredRetrieverRemoteFs) GetEndpoint() string {
	return retriever.Endpoint
}

// GetLocale returns locale.
func (retriever *CredRetrieverRemoteFs) GetLocale() string {
	return retriever.Locale
}
