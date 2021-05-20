// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/grantae/certinfo"
	"github.com/hashicorp/consul/api"
	"github.com/rookie-ninja/rk-common/common"
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
	CertEntryName        = "CertDefault"
	CertEntryType        = "CertEntry"
	CertEntryDescription = "Internal RK entry which retrieves certificates from localFs, remoteFs, etcd or consul."
	DefaultTimeout       = 3 * time.Second
	ProviderEtcd         = "etcd"
	ProviderConsul       = "consul"
	ProviderLocalFs      = "localFs"
	ProviderRemoteFs     = "remoteFs"
)

// Bootstrap config of CertEntry.
// etcd:
// 1: Cert.Name: Name of section, required.
// 2: Cert.Provider: etcd.
// 3: Cert.Locale: <realm>::<region>::<az>::<domain>
// 4: Cert.Endpoint: Endpoint of etcd server, http://x.x.x.x or x.x.x.x both acceptable.
// 5: Cert.BasicAuth: Basic auth for etcd server, like <user:pass>.
// 6: Cert.ServerCertPath: Key of server cert in etcd server.
// 7: Cert.ServerKeyPath: Key of server key in etcd server.
// 8: Cert.ClientCertPath: Key of client cert in etcd server.
// 9: Cert.ClientKeyPath: Key of client cert in etcd server.
//
// localFs
// 1: Cert.Local.Name: Name of section, required.
// 2: Cert.Provider: localFS.
// 3: Cert.Locale: <realm>::<region>::<az>::<domain>
// 4: Cert.ServerCertPath: Key of server cert in local fs.
// 5: Cert.ServerKeyPath: Key of server key in local fs.
// 6: Cert.ClientCertPath: Key of client cert in local fs.
// 7: Cert.ClientKeyPath: Key of client cert in local fs.
//
// consul
// 1: Cert.Name: Name of section, required.
// 2: Cert.Provider: consul.
// 3: Cert.Locale: <realm>::<region>::<az>::<domain>
// 4: Cert.Endpoint: Endpoint of consul server, http://x.x.x.x or x.x.x.x both acceptable.
// 5: Cert.Datacenter: Consul datacenter.
// 6: Cert.Token: Token for access consul.
// 7: Cert.BasicAuth: Basic auth for consul server, like <user:pass>.
// 8: Cert.ServerCertPath: Key of server cert in consul server.
// 9: Cert.ServerKeyPath: Key of server key in consul server.
// 10: Cert.ClientCertPath: Key of client cert in consul server.
// 11: Cert.ClientKeyPath: Key of client cert in consul server.
//
// remoteFs:
// 1: Cert.Name: Name of section, required.
// 2: Cert.Provider: remoteFs.
// 3: Cert.Locale: <realm>::<region>::<az>::<domain>
// 4: Cert.Endpoint: Endpoint of remoteFs server, http://x.x.x.x or x.x.x.x both acceptable.
// 5: Cert.BasicAuth: Basic auth for remoteFs server, like <user:pass>.
// 6: Cert.ServerCertPath: Key of server cert in remoteFs server.
// 7: Cert.ServerKeyPath: Key of server key in remoteFs server.
// 8: Cert.ClientCertPath: Key of client cert in remoteFs server.
// 9: Cert.ClientKeyPath: Key of client cert in remoteFs server.
//
// Logger:
// 1: Cert.Logger.ZapLogger.Ref: Name of zap logger entry defined in ZapLoggerEntry.
// 2: Cert.Logger.EventLogger.Ref: Name of event logger entry defined in EventLoggerEntry.
type BootConfigCert struct {
	Cert []struct {
		Name           string `yaml:"name" json:"name"`
		Description    string `yaml:"description" json:"description"`
		Provider       string `yaml:"provider" json:"provider"`
		Locale         string `yaml:"locale" json:"locale"`
		Endpoint       string `yaml:"endpoint" json:"endpoint"`
		Datacenter     string `yaml:"datacenter" json:"datacenter"`
		Token          string `yaml:"token" json:"token"`
		BasicAuth      string `yaml:"basicAuth" json:"basicAuth"`
		ServerCertPath string `yaml:"serverCertPath" json:"serverCertPath"`
		ServerKeyPath  string `yaml:"serverKeyPath" json:"serverKeyPath"`
		ClientCertPath string `yaml:"clientCertPath" json:"clientCertPath"`
		ClientKeyPath  string `yaml:"clientKeyPath" json:"clientKeyPath"`
		Logger         struct {
			ZapLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"zapLogger" json:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"cert" json:"cert"`
}

// Stores certificate as byte array.
// ServerCert: Server certificate.
// ServerKey: Private key of server certificate.
// ClientCert: Client certificate.
// ClientKey: Private key of client certificate.
type CertStore struct {
	// Server certificate
	ServerCert []byte `json:"-" yaml:"-"`
	// Server key
	ServerKey []byte `json:"-" yaml:"-"`
	// Client certificate, useful while client authentication was enabled from server
	ClientCert []byte `json:"-" yaml:"-"`
	// Client key (private), useful while client authentication was enabled from server
	ClientKey []byte `json:"-" yaml:"-"`
}

// Parse server certificate to human readable string.
func (store *CertStore) SeverCertString() string {
	if len(store.ServerCert) < 1 {
		return ""
	}

	cert, err := store.parseCert(store.ServerCert)
	if err != nil {
		return ""
	}

	res, err := certinfo.CertificateText(cert)
	if err != nil {
		return ""
	}

	return res
}

// Parse client certificate to human readable string.
func (store *CertStore) ClientCertString() string {
	if len(store.ServerCert) < 1 {
		return ""
	}

	cert, err := store.parseCert(store.ClientCert)
	if err != nil {
		return ""
	}

	res, err := certinfo.CertificateText(cert)
	if err != nil {
		return ""
	}

	return res
}

// Parse bytes into Certificate instance, used for stringfy certificate.
func (store *CertStore) parseCert(bytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, errors.New("failed to decode cert")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// Marshal entry
func (store *CertStore) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"serverCert": store.SeverCertString(),
		"clientCert": store.ClientCertString(),
	}

	return json.Marshal(&m)
}

// Unmarshal entry
func (store *CertStore) UnmarshalJSON([]byte) error {
	return nil
}

// CertEntry contains bellow fields.
// 1: EntryName: Name of entry.
// 2: EntryType: Type of entry which is CertEntry.
// 3: EntryDescription: Description of entry.
// 4: ZapLoggerEntry: ZapLoggerEntry was initialized at the beginning.
// 5: EventLoggerEntry: EventLoggerEntry was initialized at the beginning.
// 6: Store: Certificate store.
// 7: Retriever: Certificate retriever.
type CertEntry struct {
	EntryName        string            `json:"entryName" yaml:"entryName"`
	EntryType        string            `json:"entryType" yaml:"entryType"`
	EntryDescription string            `json:"entryDescription" yaml:"entryDescription"`
	ZapLoggerEntry   *ZapLoggerEntry   `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry `json:"-" yaml:"-"`
	Store            *CertStore        `json:"store" yaml:"store"`
	Retriever        CertRetriever     `json:"retrievers" yaml:"retrievers"`
}

// CertEntryOption Option which used while registering entry from codes.
type CertEntryOption func(entry *CertEntry)

// Provide name.
func WithNameCert(name string) CertEntryOption {
	return func(entry *CertEntry) {
		entry.EntryName = name
	}
}

// Provide description.
func WithDescriptionCert(description string) CertEntryOption {
	return func(entry *CertEntry) {
		entry.EntryDescription = description
	}
}

// Provide ZapLoggerEntry.
func WithZapLoggerEntryCert(logger *ZapLoggerEntry) CertEntryOption {
	return func(entry *CertEntry) {
		if logger != nil {
			entry.ZapLoggerEntry = logger
		}
	}
}

// Provide EventLoggerEntry.
func WithEventLoggerEntryCert(logger *EventLoggerEntry) CertEntryOption {
	return func(entry *CertEntry) {
		if logger != nil {
			entry.EventLoggerEntry = logger
		}
	}
}

// Provide list of CertRetriever.
func WithCertRetrieverCert(retriever CertRetriever) CertEntryOption {
	return func(entry *CertEntry) {
		entry.Retriever = retriever
	}
}

// Implements rkentry.EntryRegFunc which generate Entry based on boot configuration file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterCertEntriesFromConfig(configFilePath string) map[string]Entry {
	config := &BootConfigCert{}

	rkcommon.UnmarshalBootConfig(configFilePath, config)

	res := make(map[string]Entry)

	for i := range config.Cert {
		element := config.Cert[i]

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

		var retriever CertRetriever

		switch element.Provider {
		case ProviderConsul:
			retriever = &CertRetrieverConsul{
				Provider:         element.Provider,
				Locale:           element.Locale,
				Endpoint:         element.Endpoint,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Datacenter:       element.Datacenter,
				Token:            element.Token,
				BasicAuth:        element.BasicAuth,
				ServerCertPath:   element.ServerCertPath,
				ServerKeyPath:    element.ServerKeyPath,
				ClientCertPath:   element.ClientCertPath,
				ClientKeyPath:    element.ClientKeyPath,
			}
		case ProviderEtcd:
			retriever = &CertRetrieverEtcd{
				Provider:         element.Provider,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Locale:           element.Locale,
				Endpoint:         element.Endpoint,
				BasicAuth:        element.BasicAuth,
				ServerCertPath:   element.ServerCertPath,
				ServerKeyPath:    element.ServerKeyPath,
				ClientCertPath:   element.ClientCertPath,
				ClientKeyPath:    element.ClientKeyPath,
			}
		case ProviderLocalFs:
			retriever = &CertRetrieverLocalFs{
				Provider:         element.Provider,
				Locale:           element.Locale,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				ServerCertPath:   element.ServerCertPath,
				ServerKeyPath:    element.ServerKeyPath,
				ClientCertPath:   element.ClientCertPath,
				ClientKeyPath:    element.ClientKeyPath,
			}
		case ProviderRemoteFs:
			retriever = &CertRetrieverRemoteFs{
				Provider:         element.Provider,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Locale:           element.Locale,
				Endpoint:         element.Endpoint,
				BasicAuth:        element.BasicAuth,
				ServerCertPath:   element.ServerCertPath,
				ServerKeyPath:    element.ServerKeyPath,
				ClientCertPath:   element.ClientCertPath,
				ClientKeyPath:    element.ClientKeyPath,
			}
		}

		entry := RegisterCertEntry(
			WithNameCert(element.Name),
			WithDescriptionCert(element.Description),
			WithZapLoggerEntryCert(zapLoggerEntry),
			WithEventLoggerEntryCert(eventLoggerEntry),
			WithCertRetrieverCert(retriever))

		res[entry.GetName()] = entry
	}

	return res
}

// Create cert entry with options.
func RegisterCertEntry(opts ...CertEntryOption) *CertEntry {
	entry := &CertEntry{
		EventLoggerEntry: GlobalAppCtx.GetEventLoggerEntryDefault(),
		ZapLoggerEntry:   GlobalAppCtx.GetZapLoggerEntryDefault(),
		EntryName:        CertEntryName,
		EntryType:        CertEntryType,
		EntryDescription: CertEntryDescription,
	}

	for i := range opts {
		opts[i](entry)
	}

	GlobalAppCtx.AddCertEntry(entry)

	return entry
}

// Iterate retrievers and call Retrieve() for each of them.
func (entry *CertEntry) Bootstrap(ctx context.Context) {
	entry.Store = entry.Retriever.Retrieve(ctx)
}

// Interrupt entry.
func (entry *CertEntry) Interrupt(context.Context) {
	// no op
}

// Return string of entry.
func (entry *CertEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// Marshal entry
func (entry *CertEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":        entry.EntryName,
		"entryType":        entry.EntryType,
		"entryDescription": entry.EntryDescription,
		"eventLoggerEntry": entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":   entry.ZapLoggerEntry.GetName(),
		"store":            entry.Store,
		"retriever":        entry.Retriever,
	}

	return json.Marshal(&m)
}

// Unmarshal entry
func (entry *CertEntry) UnmarshalJSON([]byte) error {
	return nil
}

// Get name of entry.
func (entry *CertEntry) GetName() string {
	return entry.EntryName
}

// Get type of entry.
func (entry *CertEntry) GetType() string {
	return entry.EntryType
}

// Return description of entry
func (entry *CertEntry) GetDescription() string {
	return entry.EntryDescription
}

// Interface for retrieving certificates.
type CertRetriever interface {
	// Read certificate files into byte array and store it into CertStore.
	Retrieve(context.Context) *CertStore

	// Return privider of retriever.
	GetProvider() string

	// Return server cert path.
	GetServerCertPath() string

	// Return server key path.
	GetServerKeyPath() string

	// Return client cert path.
	GetClientCertPath() string

	// Return client key path.
	GetClientKeyPath() string

	// Return endpoint.
	GetEndpoint() string

	// Return locale.
	GetLocale() string
}

// ******************************
// ************ etcd ************
// ******************************

// 1: Name: Name of section, required.
// 2: Provider: Provider of retriever, required.
// 3: Locale: <realm>::<region>::<az>::<domain>
// 4: Endpoint: Endpoint of ETCD server, http://x.x.x.x or x.x.x.x both acceptable.
// 5: BasicAuth: Basic auth for ETCD server, like <user:pass>.
// 6: ServerCertPath: Key of server cert in ETCD server.
// 7: ServerKeyPath: Key of server key in ETCD server.
// 8: ClientCertPath: Key of client cert in ETCD server.
// 9: ClientKeyPath: Key of client cert in ETCD server.
type CertRetrieverEtcd struct {
	Name             string            `yaml:"name" json:"name"`
	Provider         string            `yaml:"provider" json:"provider"`
	Locale           string            `yaml:"locale" json:"locale"`
	ZapLoggerEntry   *ZapLoggerEntry   `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry `json:"-" yaml:"-"`
	Endpoint         string            `yaml:"endpoint" json:"endpoint"`
	BasicAuth        string            `json:"-" yaml:"-"`
	ServerCertPath   string            `yaml:"serverCertPath" json:"serverCertPath"`
	ServerKeyPath    string            `yaml:"serverKeyPath" json:"serverKeyPath"`
	ClientCertPath   string            `yaml:"clientCertPath" json:"clientCertPath"`
	ClientKeyPath    string            `yaml:"clientKeyPath" json:"clientKeyPath"`
}

// Call ETCD server and retrieve values based on keys.
func (retriever *CertRetrieverEtcd) Retrieve(context.Context) *CertStore {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{retriever.Endpoint},
		DialTimeout: DefaultTimeout,
		LogConfig:   retriever.ZapLoggerEntry.GetLoggerConfig(),
		Username:    rkcommon.GetUsernameFromBasicAuthString(retriever.BasicAuth),
		Password:    rkcommon.GetPasswordFromBasicAuthString(retriever.BasicAuth),
	})

	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to create etcd client v3",
			zap.Error(err))
		return nil
	}

	defer client.Close()

	return &CertStore{
		ServerCert: retriever.getValueFromEtcd(client, retriever.ServerCertPath),
		ServerKey:  retriever.getValueFromEtcd(client, retriever.ServerKeyPath),
		ClientCert: retriever.getValueFromEtcd(client, retriever.ClientCertPath),
		ClientKey:  retriever.getValueFromEtcd(client, retriever.ClientKeyPath),
	}
}

// Inner utility function.
func (retriever *CertRetrieverEtcd) getValueFromEtcd(client *clientv3.Client, key string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	if resp, err := client.Get(ctx, key); err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to get cert from etcd",
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

// Get provider of retriever.
func (retriever *CertRetrieverEtcd) GetProvider() string {
	return retriever.Provider
}

// Return server cert path.
func (retriever *CertRetrieverEtcd) GetServerCertPath() string {
	return retriever.ServerCertPath
}

// Return server key path.
func (retriever *CertRetrieverEtcd) GetServerKeyPath() string {
	return retriever.ServerKeyPath
}

// Return client cert path.
func (retriever *CertRetrieverEtcd) GetClientCertPath() string {
	return retriever.ClientCertPath
}

// Return client key path.
func (retriever *CertRetrieverEtcd) GetClientKeyPath() string {
	return retriever.ClientKeyPath
}

// Return endpoint.
func (retriever *CertRetrieverEtcd) GetEndpoint() string {
	return retriever.Endpoint
}

// Return locale.
func (retriever *CertRetrieverEtcd) GetLocale() string {
	return retriever.Locale
}

// ********************************
// ************ consul ************
// ********************************

// 1: Provider: Provider of retriever, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: Endpoint: Endpoint of consul server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: Datacenter: Consul datacenter.
// 5: Token: Token for access Consul.
// 6: BasicAuth: Basic auth for Consul server, like <user:pass>.
// 7: ServerCertPath: Path of server cert in Consul server.
// 8: ServerKeyPath: Path of server key in Consul server.
// 9: ClientCertPath: Path of client cert in Consul server.
// 10: ClientKeyPath: Path of client cert in Consul server.
type CertRetrieverConsul struct {
	Provider         string            `yaml:"provider" json:"provider"`
	ZapLoggerEntry   *ZapLoggerEntry   `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry `json:"-" yaml:"-"`
	Locale           string            `yaml:"locale" json:"locale"`
	Endpoint         string            `yaml:"endpoint" json:"endpoint"`
	Datacenter       string            `yaml:"datacenter" json:"datacenter"`
	Token            string            `json:"-" yaml:"-"`
	BasicAuth        string            `json:"-" yaml:"-"`
	ServerCertPath   string            `yaml:"serverCertPath" json:"serverCertPath"`
	ServerKeyPath    string            `yaml:"serverKeyPath" json:"serverKeyPath"`
	ClientCertPath   string            `yaml:"clientCertPath" json:"clientCertPath"`
	ClientKeyPath    string            `yaml:"clientKeyPath" json:"clientKeyPath"`
}

// Call Consul server/agent and retrieve values based on keys.
func (retriever *CertRetrieverConsul) Retrieve(context.Context) *CertStore {
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

	// Get a new client
	client, err := api.NewClient(config)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to create consul client v3",
			zap.Error(err))
		return nil
	}

	return &CertStore{
		ServerCert: retriever.getValueFromConsul(client, retriever.ServerCertPath),
		ServerKey:  retriever.getValueFromConsul(client, retriever.ServerKeyPath),
		ClientCert: retriever.getValueFromConsul(client, retriever.ClientCertPath),
		ClientKey:  retriever.getValueFromConsul(client, retriever.ClientKeyPath),
	}
}

// Get provider of retriever.
func (retriever *CertRetrieverConsul) GetProvider() string {
	return retriever.Provider
}

// Return server cert path.
func (retriever *CertRetrieverConsul) GetServerCertPath() string {
	return retriever.ServerCertPath
}

// Return server key path.
func (retriever *CertRetrieverConsul) GetServerKeyPath() string {
	return retriever.ServerKeyPath
}

// Return client cert path.
func (retriever *CertRetrieverConsul) GetClientCertPath() string {
	return retriever.ClientCertPath
}

// Return client key path.
func (retriever *CertRetrieverConsul) GetClientKeyPath() string {
	return retriever.ClientKeyPath
}

// Return endpoint.
func (retriever *CertRetrieverConsul) GetEndpoint() string {
	return retriever.Endpoint
}

// Return locale.
func (retriever *CertRetrieverConsul) GetLocale() string {
	return retriever.Locale
}

// Inner utility function.
func (retriever *CertRetrieverConsul) getValueFromConsul(client *api.Client, key string) []byte {
	// Get a handle to the KV API
	kv := client.KV()

	// Lookup the pair
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to get cert from consul",
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

// 1: Provider: Type of retriever, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: ServerCertPath: Path of server cert in localFs.
// 4: ServerKeyPath: Path of server key in localFs.
// 5: ClientCertPath: Path of client cert in localFs.
// 6: ClientKeyPath: Path of client cert in localFs.
type CertRetrieverLocalFs struct {
	Provider         string            `yaml:"provider" json:"provider"`
	Locale           string            `yaml:"locale" json:"locale"`
	ZapLoggerEntry   *ZapLoggerEntry   `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry `json:"-" yaml:"-"`
	ServerCertPath   string            `yaml:"serverCertPath" json:"serverCertPath""`
	ServerKeyPath    string            `yaml:"serverKeyPath" json:"serverKeyPath"`
	ClientCertPath   string            `yaml:"clientCertPath" json:"clientCertPath""`
	ClientKeyPath    string            `yaml:"clientKeyPath" json:"clientKeyPath"`
}

// Read files from local file system and retrieve values based on keys.
func (retriever *CertRetrieverLocalFs) Retrieve(context.Context) *CertStore {
	wd, err := os.Getwd()

	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to get working directory", zap.Error(err))
		return nil
	}

	if len(retriever.ServerCertPath) > 0 && !path.IsAbs(retriever.ServerCertPath) {
		retriever.ServerCertPath = path.Join(wd, retriever.ServerCertPath)
	}

	if len(retriever.ServerKeyPath) > 0 && !path.IsAbs(retriever.ServerKeyPath) {
		retriever.ServerKeyPath = path.Join(wd, retriever.ServerKeyPath)
	}

	if len(retriever.ClientCertPath) > 0 && !path.IsAbs(retriever.ClientCertPath) {
		retriever.ClientCertPath = path.Join(wd, retriever.ClientCertPath)
	}

	if len(retriever.ClientKeyPath) > 0 && !path.IsAbs(retriever.ClientKeyPath) {
		retriever.ClientKeyPath = path.Join(wd, retriever.ClientKeyPath)
	}

	var serverCert, serverKey []byte
	if len(retriever.ServerCertPath) > 0 {
		serverCert, err = ioutil.ReadFile(retriever.ServerCertPath)
		if err != nil {
			retriever.ZapLoggerEntry.GetLogger().Warn("failed to read server cert",
				zap.Error(err),
				zap.String("path", retriever.ServerCertPath))
			return nil
		}
	}

	if len(retriever.ServerKeyPath) > 0 {
		serverKey, err = ioutil.ReadFile(retriever.ServerKeyPath)
		if err != nil {
			retriever.ZapLoggerEntry.GetLogger().Warn("failed to read server key",
				zap.Error(err),
				zap.String("path", retriever.ServerKeyPath))
			return nil
		}
	}

	var clientCert, clientKey []byte
	if len(retriever.ClientCertPath) > 0 {
		clientCert, err = ioutil.ReadFile(retriever.ClientCertPath)
		if err != nil {
			retriever.ZapLoggerEntry.GetLogger().Warn("failed to read client cert",
				zap.Error(err),
				zap.String("path", retriever.ClientCertPath))
			return nil
		}
	}

	if len(retriever.ClientKeyPath) > 0 {
		clientKey, err = ioutil.ReadFile(retriever.ClientKeyPath)
		if err != nil {
			retriever.ZapLoggerEntry.GetLogger().Warn("failed to read client key",
				zap.Error(err),
				zap.String("path", retriever.ClientKeyPath))
			return nil
		}
	}

	return &CertStore{
		ServerCert: serverCert,
		ServerKey:  serverKey,
		ClientCert: clientCert,
		ClientKey:  clientKey,
	}
}

// Get provider of retriever.
func (retriever *CertRetrieverLocalFs) GetProvider() string {
	return retriever.Provider
}

// Return server cert path.
func (retriever *CertRetrieverLocalFs) GetServerCertPath() string {
	return retriever.ServerCertPath
}

// Return server key path.
func (retriever *CertRetrieverLocalFs) GetServerKeyPath() string {
	return retriever.ServerKeyPath
}

// Return client cert path.
func (retriever *CertRetrieverLocalFs) GetClientCertPath() string {
	return retriever.ClientCertPath
}

// Return client key path.
func (retriever *CertRetrieverLocalFs) GetClientKeyPath() string {
	return retriever.ClientKeyPath
}

// Return endpoint.
func (retriever *CertRetrieverLocalFs) GetEndpoint() string {
	return "local"
}

// Return locale.
func (retriever *CertRetrieverLocalFs) GetLocale() string {
	return retriever.Locale
}

// **********************************
// ************ remoteFs ************
// **********************************

// 1: Provider: Provider of retriever, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: Endpoint: Endpoint of RemoteFileStore server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: BasicAuth: Basic auth for RemoteFileStore server, like <user:pass>.
// 5: ServerCertPath: Path of server cert in remoteFs server.
// 6: ServerKeyPath: Path of server key in remoteFs server.
// 7: ClientCertPath: Path of client cert in remoteFs server.
// 8: ClientKeyPath: Path of client cert in remoteFs server.
type CertRetrieverRemoteFs struct {
	Provider         string            `yaml:"provider" json:"provider"`
	ZapLoggerEntry   *ZapLoggerEntry   `json:"-" yaml:"-"`
	EventLoggerEntry *EventLoggerEntry `json:"-" yaml:"-"`
	Locale           string            `yaml:"locale" json:"locale"`
	Endpoint         string            `yaml:"endpoint" json:"endpoint"`
	BasicAuth        string            `json:"-" yaml:"-"`
	ServerCertPath   string            `yaml:"serverCertPath" json:"serverCertPath"`
	ServerKeyPath    string            `yaml:"serverKeyPath" json:"serverKeyPath"`
	ClientCertPath   string            `yaml:"clientCertPath" json:"clientCertPath"`
	ClientKeyPath    string            `yaml:"clientKeyPath" json:"clientKeyPath"`
}

// Call remote file store and retrieve values based on keys.
func (retriever *CertRetrieverRemoteFs) Retrieve(context.Context) *CertStore {
	client := &http.Client{
		Timeout: DefaultTimeout,
	}

	return &CertStore{
		ServerCert: retriever.getValueFromRemoteFs(client, retriever.ServerCertPath),
		ServerKey:  retriever.getValueFromRemoteFs(client, retriever.ServerKeyPath),
		ClientCert: retriever.getValueFromRemoteFs(client, retriever.ClientCertPath),
		ClientKey:  retriever.getValueFromRemoteFs(client, retriever.ClientKeyPath),
	}
}

// Inner utility function.
func (retriever *CertRetrieverRemoteFs) getValueFromRemoteFs(client *http.Client, certPath string) []byte {
	if !strings.HasPrefix(retriever.Endpoint, "http://") {
		retriever.Endpoint = "http://" + retriever.Endpoint
	}

	if !strings.HasPrefix(certPath, "/") {
		certPath = "/" + certPath
	}

	req, err := http.NewRequest("GET", retriever.Endpoint+certPath, nil)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed create http request",
			zap.String("endpoint", retriever.Endpoint),
			zap.String("locale", retriever.Locale),
			zap.String("certPath", certPath),
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
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to get cert from remote file store",
			zap.String("endpoint", retriever.Endpoint),
			zap.String("locale", retriever.Locale),
			zap.String("certPath", certPath),
			zap.Error(err))
		return nil
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		retriever.ZapLoggerEntry.GetLogger().Warn("failed to read cert from remote file store",
			zap.String("endpoint", retriever.Endpoint),
			zap.String("locale", retriever.Locale),
			zap.String("certPath", certPath),
			zap.Error(err))
		return nil
	}

	return res
}

// Get provider of retriever.
func (retriever *CertRetrieverRemoteFs) GetProvider() string {
	return retriever.Provider
}

// Return server cert path.
func (retriever *CertRetrieverRemoteFs) GetServerCertPath() string {
	return retriever.ServerCertPath
}

// Return server key path.
func (retriever *CertRetrieverRemoteFs) GetServerKeyPath() string {
	return retriever.ServerKeyPath
}

// Return client cert path.
func (retriever *CertRetrieverRemoteFs) GetClientCertPath() string {
	return retriever.ClientCertPath
}

// Return client key path.
func (retriever *CertRetrieverRemoteFs) GetClientKeyPath() string {
	return retriever.ClientKeyPath
}

// Return endpoint.
func (retriever *CertRetrieverRemoteFs) GetEndpoint() string {
	return retriever.Endpoint
}

// Return locale.
func (retriever *CertRetrieverRemoteFs) GetLocale() string {
	return retriever.Locale
}
