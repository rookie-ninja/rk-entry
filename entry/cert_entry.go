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
	"github.com/rookie-ninja/rk-common/common"
)

const (
	CertEntryName        = "CertDefault"
	CertEntryType        = "CertEntry"
	CertEntryDescription = "Internal RK entry which retrieves certificates from localFs, remoteFs, etcd or consul."
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
	Retriever        Retriever         `json:"retriever" yaml:"retriever"`
	ServerKeyPath    string            `json:"serverKeyPath" yaml:"serverKeyPath"`
	ServerCertPath   string            `json:"serverCertPath" yaml:"serverCertPath"`
	ClientKeyPath    string            `json:"clientKeyPath" yaml:"clientKeyPath"`
	ClientCertPath   string            `json:"clientKeyPath" yaml:"clientKeyPath"`
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

// Provide Retriever.
func WithRetrieverCert(retriever Retriever) CertEntryOption {
	return func(entry *CertEntry) {
		entry.Retriever = retriever
	}
}

// Provide server key path.
func WithServerKeyPath(serverKeyPath string) CertEntryOption {
	return func(entry *CertEntry) {
		entry.ServerKeyPath = serverKeyPath
	}
}

// Provide server cert path.
func WithServerCertPath(serverCertPath string) CertEntryOption {
	return func(entry *CertEntry) {
		entry.ServerCertPath = serverCertPath
	}
}

// Provide client key path.
func WithClientKeyPath(clientKeyPath string) CertEntryOption {
	return func(entry *CertEntry) {
		entry.ClientKeyPath = clientKeyPath
	}
}

// Provide client cert path.
func WithClientCertPath(clientCertPath string) CertEntryOption {
	return func(entry *CertEntry) {
		entry.ClientCertPath = clientCertPath
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
				Paths: []string{
					element.ServerKeyPath,
					element.ServerCertPath,
					element.ClientKeyPath,
					element.ClientCertPath,
				},
			}
		case ProviderEtcd:
			retriever = &CredRetrieverEtcd{
				Provider:         element.Provider,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Locale:           element.Locale,
				Endpoint:         element.Endpoint,
				BasicAuth:        element.BasicAuth,
				Paths: []string{
					element.ServerKeyPath,
					element.ServerCertPath,
					element.ClientKeyPath,
					element.ClientCertPath,
				},
			}
		case ProviderLocalFs:
			retriever = &CredRetrieverLocalFs{
				Provider:         element.Provider,
				Locale:           element.Locale,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Paths: []string{
					element.ServerKeyPath,
					element.ServerCertPath,
					element.ClientKeyPath,
					element.ClientCertPath,
				},
			}
		case ProviderRemoteFs:
			retriever = &CredRetrieverRemoteFs{
				Provider:         element.Provider,
				ZapLoggerEntry:   zapLoggerEntry,
				EventLoggerEntry: eventLoggerEntry,
				Locale:           element.Locale,
				Endpoint:         element.Endpoint,
				BasicAuth:        element.BasicAuth,
				Paths: []string{
					element.ServerKeyPath,
					element.ServerCertPath,
					element.ClientKeyPath,
					element.ClientCertPath,
				},
			}
		}

		entry := RegisterCertEntry(
			WithNameCert(element.Name),
			WithDescriptionCert(element.Description),
			WithZapLoggerEntryCert(zapLoggerEntry),
			WithEventLoggerEntryCert(eventLoggerEntry),
			WithRetrieverCert(retriever),
			WithServerKeyPath(element.ServerKeyPath),
			WithServerCertPath(element.ServerCertPath),
			WithClientKeyPath(element.ClientKeyPath),
			WithClientCertPath(element.ClientCertPath))

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
		Store:            &CertStore{},
	}

	for i := range opts {
		opts[i](entry)
	}

	GlobalAppCtx.AddCertEntry(entry)

	return entry
}

// Iterate retrievers and call Retrieve() for each of them.
func (entry *CertEntry) Bootstrap(ctx context.Context) {
	credStore := entry.Retriever.Retrieve(ctx)

	entry.Store.ServerCert = credStore.GetCred(entry.ServerCertPath)
	entry.Store.ServerKey = credStore.GetCred(entry.ServerKeyPath)
	entry.Store.ClientCert = credStore.GetCred(entry.ClientCertPath)
	entry.Store.ClientKey = credStore.GetCred(entry.ClientKeyPath)
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
