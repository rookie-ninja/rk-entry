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
	CertEntryName  = "rk-cert-entry"
	CertEntryType  = "rk-cert-entry"
	DefaultTimeout = 3 * time.Second
)

// Bootstrap config of application's basic information.
// ETCD:
// 1: Cert.ETCD.Name: Name of section, required.
// 2: Cert.ETCD.Locale: <realm>::<region>::<az>::<domain>
// 3: Cert.ETCD.Endpoint: Endpoint of ETCD server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: Cert.ETCD.BasicAuth: Basic auth for ETCD server, like <user:pass>.
// 5: Cert.ETCD.ServerCertPath: Key of server cert in ETCD server.
// 6: Cert.ETCD.ServerKeyPath: Key of server key in ETCD server.
// 7: Cert.ETCD.ClientCertPath: Key of client cert in ETCD server.
// 8: Cert.ETCD.ClientKeyPath: Key of client cert in ETCD server.
//
// Local FS
// 1: Cert.Local.Name: Name of section, required.
// 2: Cert.Local.Locale: <realm>::<region>::<az>::<domain>
// 3: Cert.Local.ServerCertPath: Key of server cert in local fs.
// 4: Cert.Local.ServerKeyPath: Key of server key in local fs.
// 5: Cert.Local.ClientCertPath: Key of client cert in local fs.
// 6: Cert.Local.ClientKeyPath: Key of client cert in local fs.
//
// Consul
// 1: Cert.Consul.Name: Name of section, required.
// 2: Cert.Consul.Locale: <realm>::<region>::<az>::<domain>
// 3: Cert.Consul.Endpoint: Endpoint of Consul server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: Cert.Consul.Datacenter: Consul datacenter.
// 5: Cert.Consul.Token: Token for access Consul.
// 6: Cert.Consul.BasicAuth: Basic auth for Consul server, like <user:pass>.
// 7: Cert.Consul.ServerCertPath: Key of server cert in Consul server.
// 8: Cert.Consul.ServerKeyPath: Key of server key in Consul server.
// 9: Cert.Consul.ClientCertPath: Key of client cert in Consul server.
// 10: Cert.Consul.ClientKeyPath: Key of client cert in Consul server.
//
// Remote File Store:
// 1: Cert.RemoteFileStore.Name: Name of section, required.
// 2: Cert.RemoteFileStore.Locale: <realm>::<region>::<az>::<domain>
// 3: Cert.RemoteFileStore.Endpoint: Endpoint of RemoteFileStore server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: Cert.RemoteFileStore.BasicAuth: Basic auth for RemoteFileStore server, like <user:pass>.
// 5: Cert.RemoteFileStore.ServerCertPath: Key of server cert in RemoteFileStore server.
// 6: Cert.RemoteFileStore.ServerKeyPath: Key of server key in RemoteFileStore server.
// 7: Cert.RemoteFileStore.ClientCertPath: Key of client cert in RemoteFileStore server.
// 8: Cert.RemoteFileStore.ClientKeyPath: Key of client cert in RemoteFileStore server.
//
// Logger:
// 1: Cert.Logger.ZapLogger.Ref: Name of zap logger entry defined in ZapLoggerEntry.
// 2: Cert.Logger.EventLogger.Ref: Name of event logger entry defined in EventLoggerEntry.
type BootConfigCert struct {
	Cert struct {
		Consul []struct {
			Name           string `yaml:"name"`
			Locale         string `yaml:"locale"`
			Endpoint       string `yaml:"endpoint"`
			Datacenter     string `yaml:"datacenter"`
			Token          string `yaml:"token"`
			BasicAuth      string `yaml:"basicAuth"`
			ServerCertPath string `yaml:"serverCertPath"`
			ServerKeyPath  string `yaml:"serverKeyPath"`
			ClientCertPath string `yaml:"clientCertPath"`
			ClientKeyPath  string `yaml:"clientKeyPath"`
		} `yaml:"consul"`
		ETCD []struct {
			Name           string `yaml:"name"`
			Locale         string `yaml:"locale"`
			Endpoint       string `yaml:"endpoint"`
			BasicAuth      string `yaml:"basicAuth"`
			ServerCertPath string `yaml:"serverCertPath"`
			ServerKeyPath  string `yaml:"serverKeyPath"`
			ClientCertPath string `yaml:"clientCertPath"`
			ClientKeyPath  string `yaml:"clientKeyPath"`
		} `yaml:"etcd"`
		Local []struct {
			Name           string `yaml:"name"`
			Locale         string `yaml:"locale"`
			ServerCertPath string `yaml:"serverCertPath"`
			ServerKeyPath  string `yaml:"serverKeyPath"`
			ClientCertPath string `yaml:"clientCertPath"`
			ClientKeyPath  string `yaml:"clientKeyPath"`
		} `yaml:"local"`
		RemoteFileStore []struct {
			Name           string `yaml:"name"`
			Locale         string `yaml:"locale"`
			Endpoint       string `yaml:"endpoint"`
			BasicAuth      string `yaml:"basicAuth"`
			ServerCertPath string `yaml:"serverCertPath"`
			ServerKeyPath  string `yaml:"serverKeyPath"`
			ClientCertPath string `yaml:"clientCertPath"`
			ClientKeyPath  string `yaml:"clientKeyPath"`
		} `yaml:"remoteFileStore"`
		Logger struct {
			ZapLogger struct {
				Ref string `yaml:"ref"`
			} `yaml:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref"`
			} `yaml:"eventLogger"`
		} `yaml:"logger"`
	} `yaml:"cert"`
}

// Stores certificate as byte array.
// ServerCert: Server certificate.
// ServerKey: Private key of server certificate.
// ClientCert: Client certificate.
// ClientKey: Private key of client certificate.
type CertStore struct {
	ServerCert []byte // Server certificate
	ServerKey  []byte // Server key
	ClientCert []byte // Client certificate, useful while client authentication was enabled from server
	ClientKey  []byte // Client key (private), useful while client authentication was enabled from server
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

// Interface for retrieving certificates.
type CertRetriever interface {
	// Read certificate files into byte array and store it into CertStore.
	Retrieve(context.Context) *CertStore

	// Return name of the retriever.
	GetName() string
}

// CertEntry contains bellow fields.
// 1: entryName: Name of entry.
// 2: entryType: Type of entry which is CertEntry.
// 3: ZapLoggerEntry: ZapLoggerEntry was initialized at the beginning.
// 4: EventLoggerEntry: EventLoggerEntry was initialized at the beginning.
// 5: Stores: Map of certificate store.
// 6: Retrievers: Map of certificate retriever.
type CertEntry struct {
	entryName        string
	entryType        string
	ZapLoggerEntry   *ZapLoggerEntry
	EventLoggerEntry *EventLoggerEntry
	Stores           map[string]*CertStore
	Retrievers       map[string]CertRetriever
}

// CertEntryOption Option which used while registering entry from codes.
type CertEntryOption func(entry *CertEntry)

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
func WithCertRetrieverCert(retrievers ...CertRetriever) CertEntryOption {
	return func(entry *CertEntry) {
		for i := range retrievers {
			retriever := retrievers[i]
			if retriever != nil {
				entry.Retrievers[retriever.GetName()] = retriever
			}
		}
	}
}

// Implements rkentry.EntryRegFunc which generate RKEntry based on boot configuration file.
// Currently, only YAML file is supported.
// File path could be either relative or absolute.
func RegisterCertEntriesFromConfig(configFilePath string) map[string]Entry {
	config := &BootConfigCert{}

	rkcommon.UnmarshalBootConfig(configFilePath, config)

	res := make(map[string]Entry)

	zapLoggerEntry := GlobalAppCtx.GetZapLoggerEntry(config.Cert.Logger.ZapLogger.Ref)
	if zapLoggerEntry == nil {
		zapLoggerEntry = GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	eventLoggerEntry := GlobalAppCtx.GetEventLoggerEntry(config.Cert.Logger.EventLogger.Ref)
	if eventLoggerEntry == nil {
		eventLoggerEntry = GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	retrievers := make([]CertRetriever, 0)

	// deal with etcd
	for i := range config.Cert.ETCD {
		element := config.Cert.ETCD[i]
		if len(element.Name) < 1 || !rkcommon.MatchLocaleWithEnv(element.Locale) {
			continue
		}

		retriever := &CertRetrieverETCD{
			Name:             element.Name,
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

		retrievers = append(retrievers, retriever)
	}

	// deal with local
	for i := range config.Cert.Local {
		element := config.Cert.Local[i]

		if len(element.Name) < 1 || !rkcommon.MatchLocaleWithEnv(element.Locale) {
			continue
		}

		retriever := &CertRetrieverLocal{
			Name:             element.Name,
			Locale:           element.Locale,
			ZapLoggerEntry:   zapLoggerEntry,
			EventLoggerEntry: eventLoggerEntry,
			ServerCertPath:   element.ServerCertPath,
			ServerKeyPath:    element.ServerKeyPath,
			ClientCertPath:   element.ClientCertPath,
			ClientKeyPath:    element.ClientKeyPath,
		}

		retrievers = append(retrievers, retriever)
	}

	// deal with consul
	for i := range config.Cert.Consul {
		element := config.Cert.Consul[i]
		if len(element.Name) < 1 || !rkcommon.MatchLocaleWithEnv(element.Locale) {
			continue
		}

		retriever := &CertRetrieverConsul{
			Name:             element.Name,
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

		retrievers = append(retrievers, retriever)
	}

	// deal with remote file store
	for i := range config.Cert.RemoteFileStore {
		element := config.Cert.RemoteFileStore[i]
		if len(element.Name) < 1 || !rkcommon.MatchLocaleWithEnv(element.Locale) {
			continue
		}

		retriever := &CertRetrieverRemoteFileStore{
			Name:             element.Name,
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

		retrievers = append(retrievers, retriever)
	}

	entry := RegisterCertEntry(
		WithZapLoggerEntryCert(zapLoggerEntry),
		WithEventLoggerEntryCert(eventLoggerEntry),
		WithCertRetrieverCert(retrievers...))

	res[entry.GetName()] = entry

	return res
}

// Create cert entry with options.
func RegisterCertEntry(opts ...CertEntryOption) *CertEntry {
	entry := &CertEntry{
		EventLoggerEntry: GlobalAppCtx.GetEventLoggerEntryDefault(),
		ZapLoggerEntry:   GlobalAppCtx.GetZapLoggerEntryDefault(),
		entryName:        CertEntryName,
		entryType:        CertEntryType,
		Stores:           make(map[string]*CertStore),
		Retrievers:       make(map[string]CertRetriever),
	}

	for i := range opts {
		opts[i](entry)
	}

	GlobalAppCtx.addCertEntry(entry)

	return entry
}

// Iterate retrievers and call Retrieve() for each of them.
func (entry *CertEntry) Bootstrap(ctx context.Context) {
	for _, v := range entry.Retrievers {
		store := v.Retrieve(ctx)

		entry.Stores[v.GetName()] = store
	}
}

// Interrupt entry.
func (entry *CertEntry) Interrupt(context.Context) {
	// no op
}

// Return string of entry.
func (entry *CertEntry) String() string {
	m := map[string]interface{}{
		"entry_name": entry.entryName,
		"entry_type": entry.entryType,
	}

	retrievers := make([]string, 0)
	for k := range entry.Retrievers {
		retrievers = append(retrievers, k)
	}

	m["retrievers"] = retrievers

	for k, v := range entry.Stores {
		if v != nil {
			if len(v.ServerCert) > 0 {
				m[k+"_server_cert_exist"] = true
			} else {
				m[k+"_server_cert_exist"] = false
			}

			if len(v.ServerKey) > 0 {
				m[k+"_server_key_exist"] = true
			} else {
				m[k+"_server_key_exist"] = false
			}

			if len(v.ClientCert) > 0 {
				m[k+"_client_cert_exist"] = true
			} else {
				m[k+"_client_cert_exist"] = false
			}

			if len(v.ClientKey) > 0 {
				m[k+"_client_key_exist"] = true
			} else {
				m[k+"_client_key_exist"] = false
			}
		}
	}

	bytes, _ := json.Marshal(m)

	return string(bytes)
}

// Get name of entry.
func (entry *CertEntry) GetName() string {
	return entry.entryName
}

// Get type of entry.
func (entry *CertEntry) GetType() string {
	return entry.entryType
}

// ******************************
// ************ ETCD ************
// ******************************

// 1: Name: Name of section, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: Endpoint: Endpoint of ETCD server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: BasicAuth: Basic auth for ETCD server, like <user:pass>.
// 5: ServerCertPath: Key of server cert in ETCD server.
// 6: ServerKeyPath: Key of server key in ETCD server.
// 7: ClientCertPath: Key of client cert in ETCD server.
// 8: ClientKeyPath: Key of client cert in ETCD server.
type CertRetrieverETCD struct {
	Name             string
	Locale           string
	ZapLoggerEntry   *ZapLoggerEntry
	EventLoggerEntry *EventLoggerEntry
	Endpoint         string
	BasicAuth        string
	ServerCertPath   string
	ServerKeyPath    string
	ClientCertPath   string
	ClientKeyPath    string
}

// Call ETCD server and retrieve values based on keys.
func (retriever *CertRetrieverETCD) Retrieve(context.Context) *CertStore {
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
		ServerCert: retriever.getValueFromETCD(client, retriever.ServerCertPath),
		ServerKey:  retriever.getValueFromETCD(client, retriever.ServerKeyPath),
		ClientCert: retriever.getValueFromETCD(client, retriever.ClientCertPath),
		ClientKey:  retriever.getValueFromETCD(client, retriever.ClientKeyPath),
	}
}

// Inner utility function.
func (retriever *CertRetrieverETCD) getValueFromETCD(client *clientv3.Client, key string) []byte {
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

// Get name of retriever.
func (retriever *CertRetrieverETCD) GetName() string {
	return retriever.Name
}

// ********************************
// ************ Consul ************
// ********************************

// 1: Name: Name of section, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: Endpoint: Endpoint of Consul server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: Datacenter: Consul datacenter.
// 5: Token: Token for access Consul.
// 4: BasicAuth: Basic auth for Consul server, like <user:pass>.
// 5: ServerCertPath: Key of server cert in Consul server.
// 6: ServerKeyPath: Key of server key in Consul server.
// 7: ClientCertPath: Key of client cert in Consul server.
// 8: ClientKeyPath: Key of client cert in Consul server.
type CertRetrieverConsul struct {
	Name             string
	ZapLoggerEntry   *ZapLoggerEntry
	EventLoggerEntry *EventLoggerEntry
	Locale           string
	Endpoint         string
	Datacenter       string
	Token            string
	BasicAuth        string
	ServerCertPath   string
	ServerKeyPath    string
	ClientCertPath   string
	ClientKeyPath    string
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

// Get name of retriever.
func (retriever *CertRetrieverConsul) GetName() string {
	return retriever.Name
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

// *******************************
// ************ Local ************
// *******************************

// 1: Name: Name of section, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: ServerCertPath: Key of server cert in local fs.
// 4: ServerKeyPath: Key of server key in local fs.
// 5: ClientCertPath: Key of client cert in local fs.
// 6: ClientKeyPath: Key of client cert in local fs.
type CertRetrieverLocal struct {
	Name             string
	Locale           string
	ZapLoggerEntry   *ZapLoggerEntry
	EventLoggerEntry *EventLoggerEntry
	ServerCertPath   string
	ServerKeyPath    string
	ClientCertPath   string
	ClientKeyPath    string
}

// Read files from local file system and retrieve values based on keys.
func (retriever *CertRetrieverLocal) Retrieve(context.Context) *CertStore {
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

// Get name of retriever.
func (retriever *CertRetrieverLocal) GetName() string {
	return retriever.Name
}

// *******************************************
// ************ Remote File Store ************
// *******************************************

// 1: Name: Name of section, required.
// 2: Locale: <realm>::<region>::<az>::<domain>
// 3: Endpoint: Endpoint of RemoteFileStore server, http://x.x.x.x or x.x.x.x both acceptable.
// 4: BasicAuth: Basic auth for RemoteFileStore server, like <user:pass>.
// 5: ServerCertPath: Key of server cert in RemoteFileStore server.
// 6: ServerKeyPath: Key of server key in RemoteFileStore server.
// 7: ClientCertPath: Key of client cert in RemoteFileStore server.
// 8: ClientKeyPath: Key of client cert in RemoteFileStore server.
type CertRetrieverRemoteFileStore struct {
	Name             string
	ZapLoggerEntry   *ZapLoggerEntry
	EventLoggerEntry *EventLoggerEntry
	Locale           string
	Endpoint         string
	BasicAuth        string
	ServerCertPath   string
	ServerKeyPath    string
	ClientCertPath   string
	ClientKeyPath    string
}

// Call remote file store and retrieve values based on keys.
func (retriever *CertRetrieverRemoteFileStore) Retrieve(context.Context) *CertStore {
	client := &http.Client{
		Timeout: DefaultTimeout,
	}

	return &CertStore{
		ServerCert: retriever.getValueFromRemoteFileStore(client, retriever.ServerCertPath),
		ServerKey:  retriever.getValueFromRemoteFileStore(client, retriever.ServerKeyPath),
		ClientCert: retriever.getValueFromRemoteFileStore(client, retriever.ClientCertPath),
		ClientKey:  retriever.getValueFromRemoteFileStore(client, retriever.ClientKeyPath),
	}
}

// Inner utility function.
func (retriever *CertRetrieverRemoteFileStore) getValueFromRemoteFileStore(client *http.Client, certPath string) []byte {
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

// Get name of retriever.
func (retriever *CertRetrieverRemoteFileStore) GetName() string {
	return retriever.Name
}
