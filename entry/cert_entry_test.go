// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSeverCertString_WithNilServerCert(t *testing.T) {
	store := CertStore{
		ServerCert: nil,
	}

	assert.Empty(t, store.SeverCertString())
}

func TestSeverCertString_WithEmptyServerCert(t *testing.T) {
	store := CertStore{
		ServerCert: []byte{},
	}

	assert.Empty(t, store.SeverCertString())
}

func TestSeverCertString_WithInvalidServerCert(t *testing.T) {
	store := CertStore{
		ServerCert: []byte{0, 1},
	}

	assert.Empty(t, store.SeverCertString())
}

func TestSeverCertString_HappyCase(t *testing.T) {
	cert := `
-----BEGIN CERTIFICATE-----
MIIC/jCCAeagAwIBAgIUWVMP53O835+njsr23UZIX2KEXGYwDQYJKoZIhvcNAQEL
BQAwYDELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxCzAJBgNVBAcTAkJK
MQswCQYDVQQKEwJSSzEQMA4GA1UECxMHUksgRGVtbzETMBEGA1UEAxMKUksgRGVt
byBDQTAeFw0yMTA0MDcxMzAzMDBaFw0yNjA0MDYxMzAzMDBaMEIxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMQswCQYDVQQHEwJCSjEUMBIGA1UEAxMLZXhh
bXBsZS5uZXQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARf8p/nxvY1HHUkJXZk
fFQgDtQ2CK9DOAe6y3lE21HTJ/Vi4vHNqWko9koyYgKqgUXyiq5lGAswo68KvmD7
c2L4o4GYMIGVMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAM
BgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTv6dUlEI6NcQBzihnzKZrxKpbnTTAfBgNV
HSMEGDAWgBRgwpYKhgfeO3p2XuX0he35caeUgTAgBgNVHREEGTAXgglsb2NhbGhv
c3SHBH8AAAGHBAAAAAAwDQYJKoZIhvcNAQELBQADggEBAByqLc3QkaGNr+QqjFw7
znk9j0X4Ucm/1N6iGIp8fUi9t+mS1La6CB1ej+FoWkSYskzqBpdIkqzqZan1chyF
njhtMsWgZYW6srXNRgByA9XS2s28+xg9owcpceXa3wG4wbnTj1emcunzSrKVFjS1
IJUjl5HWCKibnVjgt4g0s9tc8KYpXkGYl23U4FUta/07YFmtW5SDF38NWrNOe5qV
EALMz1Ry0PMgY0SDtKhddDNnNS32fz40IP0wB7a31T24eZetZK/INaIi+5SM0iLx
kfqN71xKxAIIYmuI9YwWCFaZ2+qbLIiDTbR6gyuLIQ2AfwBLZ06g939ZfSqZuP8P
oxU=
-----END CERTIFICATE-----
`

	store := CertStore{
		ServerCert: []byte(cert),
	}

	assert.NotEmpty(t, store.SeverCertString())
}

func TestClientCertString_WithNilClientCert(t *testing.T) {
	store := CertStore{
		ClientCert: nil,
	}

	assert.Empty(t, store.ClientCertString())
}

func TestClientCertString_WithEmptyClientCert(t *testing.T) {
	store := CertStore{
		ClientCert: []byte{},
	}

	assert.Empty(t, store.ClientCertString())
}

func TestClientCertString_WithInvalidClientCert(t *testing.T) {
	store := CertStore{
		ClientCert: []byte{0, 1},
	}

	assert.Empty(t, store.ClientCertString())
}

func TestClientCertString_HappyCase(t *testing.T) {
	cert := `
-----BEGIN CERTIFICATE-----
MIIC/jCCAeagAwIBAgIUWVMP53O835+njsr23UZIX2KEXGYwDQYJKoZIhvcNAQEL
BQAwYDELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxCzAJBgNVBAcTAkJK
MQswCQYDVQQKEwJSSzEQMA4GA1UECxMHUksgRGVtbzETMBEGA1UEAxMKUksgRGVt
byBDQTAeFw0yMTA0MDcxMzAzMDBaFw0yNjA0MDYxMzAzMDBaMEIxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMQswCQYDVQQHEwJCSjEUMBIGA1UEAxMLZXhh
bXBsZS5uZXQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARf8p/nxvY1HHUkJXZk
fFQgDtQ2CK9DOAe6y3lE21HTJ/Vi4vHNqWko9koyYgKqgUXyiq5lGAswo68KvmD7
c2L4o4GYMIGVMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAM
BgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTv6dUlEI6NcQBzihnzKZrxKpbnTTAfBgNV
HSMEGDAWgBRgwpYKhgfeO3p2XuX0he35caeUgTAgBgNVHREEGTAXgglsb2NhbGhv
c3SHBH8AAAGHBAAAAAAwDQYJKoZIhvcNAQELBQADggEBAByqLc3QkaGNr+QqjFw7
znk9j0X4Ucm/1N6iGIp8fUi9t+mS1La6CB1ej+FoWkSYskzqBpdIkqzqZan1chyF
njhtMsWgZYW6srXNRgByA9XS2s28+xg9owcpceXa3wG4wbnTj1emcunzSrKVFjS1
IJUjl5HWCKibnVjgt4g0s9tc8KYpXkGYl23U4FUta/07YFmtW5SDF38NWrNOe5qV
EALMz1Ry0PMgY0SDtKhddDNnNS32fz40IP0wB7a31T24eZetZK/INaIi+5SM0iLx
kfqN71xKxAIIYmuI9YwWCFaZ2+qbLIiDTbR6gyuLIQ2AfwBLZ06g939ZfSqZuP8P
oxU=
-----END CERTIFICATE-----
`

	store := CertStore{
		ServerCert: []byte(cert),
	}

	assert.NotEmpty(t, store.SeverCertString())
}

func TestParseCert_WithNilInput(t *testing.T) {
	store := CertStore{}

	cert, err := store.parseCert(nil)
	assert.NotNil(t, err)
	assert.Empty(t, cert)
}

func TestParseCert_WithEmptyInput(t *testing.T) {
	store := CertStore{}

	cert, err := store.parseCert([]byte{})
	assert.NotNil(t, err)
	assert.Empty(t, cert)
}

func TestParseCert_WithInvalidCert(t *testing.T) {
	store := CertStore{}

	cert, err := store.parseCert([]byte{0, 1})
	assert.NotNil(t, err)
	assert.Empty(t, cert)
}

func TestParseCert_HappyCase(t *testing.T) {
	certStr := `
-----BEGIN CERTIFICATE-----
MIIC/jCCAeagAwIBAgIUWVMP53O835+njsr23UZIX2KEXGYwDQYJKoZIhvcNAQEL
BQAwYDELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxCzAJBgNVBAcTAkJK
MQswCQYDVQQKEwJSSzEQMA4GA1UECxMHUksgRGVtbzETMBEGA1UEAxMKUksgRGVt
byBDQTAeFw0yMTA0MDcxMzAzMDBaFw0yNjA0MDYxMzAzMDBaMEIxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMQswCQYDVQQHEwJCSjEUMBIGA1UEAxMLZXhh
bXBsZS5uZXQwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARf8p/nxvY1HHUkJXZk
fFQgDtQ2CK9DOAe6y3lE21HTJ/Vi4vHNqWko9koyYgKqgUXyiq5lGAswo68KvmD7
c2L4o4GYMIGVMA4GA1UdDwEB/wQEAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAM
BgNVHRMBAf8EAjAAMB0GA1UdDgQWBBTv6dUlEI6NcQBzihnzKZrxKpbnTTAfBgNV
HSMEGDAWgBRgwpYKhgfeO3p2XuX0he35caeUgTAgBgNVHREEGTAXgglsb2NhbGhv
c3SHBH8AAAGHBAAAAAAwDQYJKoZIhvcNAQELBQADggEBAByqLc3QkaGNr+QqjFw7
znk9j0X4Ucm/1N6iGIp8fUi9t+mS1La6CB1ej+FoWkSYskzqBpdIkqzqZan1chyF
njhtMsWgZYW6srXNRgByA9XS2s28+xg9owcpceXa3wG4wbnTj1emcunzSrKVFjS1
IJUjl5HWCKibnVjgt4g0s9tc8KYpXkGYl23U4FUta/07YFmtW5SDF38NWrNOe5qV
EALMz1Ry0PMgY0SDtKhddDNnNS32fz40IP0wB7a31T24eZetZK/INaIi+5SM0iLx
kfqN71xKxAIIYmuI9YwWCFaZ2+qbLIiDTbR6gyuLIQ2AfwBLZ06g939ZfSqZuP8P
oxU=
-----END CERTIFICATE-----
`

	store := CertStore{}

	cert, err := store.parseCert([]byte(certStr))

	assert.Nil(t, err)
	assert.NotNil(t, cert)
}

func TestWithZapLoggerEntryCert_WithNilLogger(t *testing.T) {
	entry := &CertEntry{
		ZapLoggerEntry: NoopZapLoggerEntry(),
	}

	opt := WithZapLoggerEntryCert(nil)
	opt(entry)

	assert.NotNil(t, entry.ZapLoggerEntry)
}

func TestWithZapLoggerEntryCert_HappyCase(t *testing.T) {
	entry := &CertEntry{}

	loggerEntry := NoopZapLoggerEntry()
	opt := WithZapLoggerEntryCert(loggerEntry)
	opt(entry)

	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestWithEventLoggerEntryCert_WithNilLogger(t *testing.T) {
	entry := &CertEntry{
		EventLoggerEntry: NoopEventLoggerEntry(),
	}

	opt := WithEventLoggerEntryCert(nil)
	opt(entry)

	assert.NotNil(t, entry.EventLoggerEntry)
}

func TestWithEventLoggerEntryCert_HappyCase(t *testing.T) {
	entry := &CertEntry{}

	loggerEntry := NoopEventLoggerEntry()
	opt := WithEventLoggerEntryCert(loggerEntry)
	opt(entry)

	assert.Equal(t, loggerEntry, entry.EventLoggerEntry)
}

func TestWithRetrieverCert_WithNilRetriever(t *testing.T) {
	entry := &CertEntry{}

	opt := WithRetrieverCert(nil)
	opt(entry)

	assert.Nil(t, entry.Retriever)
}

func TestWithRetrieverCert_HappyCase(t *testing.T) {
	entry := &CertEntry{}

	retriever := &CredRetrieverLocalFs{}

	opt := WithRetrieverCert(retriever)
	opt(entry)

	assert.Equal(t, retriever, entry.Retriever)
}

func TestRegisterCertEntriesFromConfig_HappyCase(t *testing.T) {
	bootConfig := `
cert:
  - name: "ut-localFs"
    provider: "localFs"
    locale: "*::*::*::*"
    serverCertPath: "server.pem"
    serverKeyPath: "server-key.pem"
    clientCertPath: "client.pem"
    clientKeyPath: "client-key.pem"
  - name: "ut-etcd"
    provider: "etcd"
    locale: "*::*::*::*"
    endpoint: "localhost:2379"
    basicAuth: "root:etcd"
    serverCertPath: "server.pem"
    serverKeyPath: "server-key.pem"
    clientCertPath: "client.pem"
    clientKeyPath: "client-key.pem"
  - name: "ut-consul"
    provider: "consul"
    locale: "*::*::*::*"
    endpoint: "localhost:8500"
    basicAuth: "root:consul"
    datacenter: "rk"
    token: "token"
    serverCertPath: "server.pem"
    serverKeyPath: "server-key.pem"
    clientCertPath: "client.pem"
    clientKeyPath: "client-key.pem"
  - name: "ut-remoteFs"
    provider: "remoteFs"
    locale: "*::*::*::*"
    endpoint: "localhost:8080"
    basicAuth: "root:remote"
    serverCertPath: "server.pem"
    serverKeyPath: "server-key.pem"
    clientCertPath: "client.pem"
    clientKeyPath: "client-key.pem"
`

	bootPath := createFileAtTestTempDir(t, bootConfig)

	assert.Equal(t, len(RegisterCertEntriesFromConfig(bootPath)), 4)
	entries := GlobalAppCtx.ListCertEntries()
	assert.Equal(t, len(entries), 4)

	// localFS entry
	localFSEntry := entries["ut-localFs"]
	assert.NotNil(t, localFSEntry)
	assert.Equal(t, "ut-localFs", localFSEntry.GetName())
	assert.Equal(t, "localFs", localFSEntry.Retriever.GetProvider())
	assert.Equal(t, "server.pem", localFSEntry.ServerCertPath)
	assert.Equal(t, "server-key.pem", localFSEntry.ServerKeyPath)
	assert.Equal(t, "client.pem", localFSEntry.ClientCertPath)
	assert.Equal(t, "client-key.pem", localFSEntry.ClientKeyPath)
	assert.Equal(t, "*::*::*::*", localFSEntry.Retriever.GetLocale())

	// etcd entry
	etcdEntry := entries["ut-etcd"]
	assert.NotNil(t, etcdEntry)
	assert.Equal(t, "ut-etcd", etcdEntry.GetName())
	assert.Equal(t, "etcd", etcdEntry.Retriever.GetProvider())
	assert.Equal(t, "server.pem", etcdEntry.ServerCertPath)
	assert.Equal(t, "server-key.pem", etcdEntry.ServerKeyPath)
	assert.Equal(t, "client.pem", etcdEntry.ClientCertPath)
	assert.Equal(t, "client-key.pem", etcdEntry.ClientKeyPath)
	assert.Equal(t, "*::*::*::*", etcdEntry.Retriever.GetLocale())
	assert.Equal(t, "localhost:2379", etcdEntry.Retriever.GetEndpoint())
	assert.Equal(t, "root:etcd", etcdEntry.Retriever.(*CredRetrieverEtcd).BasicAuth)

	// consul entry
	consulEntry := entries["ut-consul"]
	assert.NotNil(t, consulEntry)
	assert.Equal(t, "ut-consul", consulEntry.GetName())
	assert.Equal(t, "consul", consulEntry.Retriever.GetProvider())
	assert.Equal(t, "server.pem", consulEntry.ServerCertPath)
	assert.Equal(t, "server-key.pem", consulEntry.ServerKeyPath)
	assert.Equal(t, "client.pem", consulEntry.ClientCertPath)
	assert.Equal(t, "client-key.pem", consulEntry.ClientKeyPath)
	assert.Equal(t, "*::*::*::*", consulEntry.Retriever.GetLocale())
	assert.Equal(t, "localhost:8500", consulEntry.Retriever.GetEndpoint())
	assert.Equal(t, "root:consul", consulEntry.Retriever.(*CredRetrieverConsul).BasicAuth)
	assert.Equal(t, "token", consulEntry.Retriever.(*CredRetrieverConsul).Token)

	// remote file store entry
	remoteFsEntry := entries["ut-remoteFs"]
	assert.NotNil(t, remoteFsEntry)
	assert.Equal(t, "ut-remoteFs", remoteFsEntry.GetName())
	assert.Equal(t, "remoteFs", remoteFsEntry.Retriever.GetProvider())
	assert.Equal(t, "server.pem", remoteFsEntry.ServerCertPath)
	assert.Equal(t, "server-key.pem", remoteFsEntry.ServerKeyPath)
	assert.Equal(t, "client.pem", remoteFsEntry.ClientCertPath)
	assert.Equal(t, "client-key.pem", remoteFsEntry.ClientKeyPath)
	assert.Equal(t, "*::*::*::*", remoteFsEntry.Retriever.GetLocale())
	assert.Equal(t, "localhost:8080", remoteFsEntry.Retriever.GetEndpoint())
	assert.Equal(t, "root:remote", remoteFsEntry.Retriever.(*CredRetrieverRemoteFs).BasicAuth)

	GlobalAppCtx.clearCertEntries()
}

func TestRegisterCertEntry_WithoutOptions(t *testing.T) {
	entry := RegisterCertEntry()

	assert.NotNil(t, entry)
	entries := GlobalAppCtx.ListCertEntries()

	assert.Equal(t, len(entries), 1)

	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.Equal(t, CertEntryName, entry.EntryName)
	assert.Equal(t, CertEntryType, entry.EntryType)
	assert.Nil(t, entry.Retriever)
	assert.NotNil(t, entry.Store)

	GlobalAppCtx.clearCertEntries()
}

func TestRegisterCertEntry_HappyCase(t *testing.T) {
	zapLoggerEntry := NoopZapLoggerEntry()
	eventLoggerEntry := NoopEventLoggerEntry()
	retriever := &CredRetrieverLocalFs{}

	entry := RegisterCertEntry(
		WithNameCert("ut"),
		WithZapLoggerEntryCert(zapLoggerEntry),
		WithEventLoggerEntryCert(eventLoggerEntry),
		WithRetrieverCert(retriever))

	assert.NotNil(t, entry)

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
	assert.Equal(t, "ut", entry.EntryName)
	assert.Equal(t, CertEntryType, entry.EntryType)
	assert.Equal(t, retriever, entry.Retriever)
	assert.NotNil(t, entry.Store)

	GlobalAppCtx.clearCertEntries()
}

func TestCertEntry_Bootstrap_HappyCase(t *testing.T) {
	retriever := &FakeRetriever{}

	entry := RegisterCertEntry(WithRetrieverCert(retriever))

	entry.Bootstrap(context.Background())

	assert.Equal(t, 1, retriever.called)

	GlobalAppCtx.clearCertEntries()
}

func TestCertEntry_Interrupt_HappyCase(t *testing.T) {
	assertNotPanic(t)
	entry := RegisterCertEntry()

	entry.Interrupt(context.Background())
}

func TestCertEntry_String_HappyCase(t *testing.T) {
	assertNotPanic(t)
	entry := RegisterCertEntry()

	assert.NotEmpty(t, entry.String())

	GlobalAppCtx.clearCertEntries()
}

func TestCertEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterCertEntry()
	assert.Equal(t, CertEntryName, entry.GetName())

	GlobalAppCtx.clearCertEntries()
}

func TestCertEntry_GetType_HappyCase(t *testing.T) {
	entry := RegisterCertEntry()
	assert.Equal(t, CertEntryType, entry.GetType())

	GlobalAppCtx.clearCertEntries()
}

type FakeRetriever struct {
	called int
}

func (retriever *FakeRetriever) Retrieve(context.Context) *CredStore {
	retriever.called++
	return &CredStore{
		Cred: make(map[string][]byte, 0),
	}
}

func (retriever *FakeRetriever) GetName() string {
	return "fake-retriever"
}

func (retriever *FakeRetriever) GetProvider() string {
	return "fake-retriever"
}

func (retriever *FakeRetriever) GetEndpoint() string {
	return "fake-endpoint"
}

func (retriever *FakeRetriever) GetLocale() string {
	return "fake-locale"
}

func (retriever *FakeRetriever) ListPaths() []string {
	return []string{"fake-path"}
}
