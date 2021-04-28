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

func TestWithCertRetrieverCert_WithoutRetrievers(t *testing.T) {
	entry := &CertEntry{
		Retrievers: map[string]CertRetriever{},
	}

	opt := WithCertRetrieverCert()
	opt(entry)

	assert.Equal(t, 0, len(entry.Retrievers))
}

func TestWithCertRetrieverCert_HappyCase(t *testing.T) {
	entry := &CertEntry{
		Retrievers: map[string]CertRetriever{},
	}

	retrieverA := &CertRetrieverLocal{
		Name: "A",
	}

	retrieverB := &CertRetrieverLocal{
		Name: "B",
	}

	opt := WithCertRetrieverCert(retrieverA, retrieverB)
	opt(entry)

	assert.Equal(t, 2, len(entry.Retrievers))
}

func TestRegisterCertEntriesFromConfig_HappyCase(t *testing.T) {
	bootConfig := `
cert:
  local:
    - name: "ut-local"
      locale: "*::*::*::*"
      serverCertPath: "server.pem"
      serverKeyPath: "server-key.pem"
      clientCertPath: "client.pem"
      clientKeyPath: "client-key.pem"
  etcd:
    - name: "ut-etcd"
      locale: "*::*::*::*"
      endpoint: "localhost:2379"
      basicAuth: "root:etcd"
      serverCertPath: "server.pem"
      serverKeyPath: "server-key.pem"
      clientCertPath: "client.pem"
      clientKeyPath: "client-key.pem"
  consul:
    - name: "ut-consul"
      locale: "*::*::*::*"
      endpoint: "localhost:8500"
      basicAuth: "root:consul"
      datacenter: "rk"
      token: "token"
      serverCertPath: "server.pem"
      serverKeyPath: "server-key.pem"
      clientCertPath: "client.pem"
      clientKeyPath: "client-key.pem"
  remoteFileStore:
    - name: "ut-remote-file-store"
      locale: "*::*::*::*"
      endpoint: "localhost:8080"
      basicAuth: "root:remote"
      serverCertPath: "server.pem"
      serverKeyPath: "server-key.pem"
      clientCertPath: "client.pem"
      clientKeyPath: "client-key.pem"
`

	bootPath := createFileAtTestTempDir(t, bootConfig)
	entries := RegisterCertEntriesFromConfig(bootPath)

	assert.Equal(t, len(entries), 1)

	raw := entries[CertEntryName]
	assert.NotNil(t, raw)

	entry := raw.(*CertEntry)

	// local
	assert.NotNil(t, entry.Retrievers["ut-local"])
	assert.Equal(t, "server.pem", entry.Retrievers["ut-local"].(*CertRetrieverLocal).ServerCertPath)
	assert.Equal(t, "server-key.pem", entry.Retrievers["ut-local"].(*CertRetrieverLocal).ServerKeyPath)
	assert.Equal(t, "client.pem", entry.Retrievers["ut-local"].(*CertRetrieverLocal).ClientCertPath)
	assert.Equal(t, "client-key.pem", entry.Retrievers["ut-local"].(*CertRetrieverLocal).ClientKeyPath)
	assert.Equal(t, "*::*::*::*", entry.Retrievers["ut-local"].(*CertRetrieverLocal).Locale)

	// etcd
	assert.NotNil(t, entry.Retrievers["ut-etcd"])
	assert.Equal(t, "server.pem", entry.Retrievers["ut-etcd"].(*CertRetrieverETCD).ServerCertPath)
	assert.Equal(t, "server-key.pem", entry.Retrievers["ut-etcd"].(*CertRetrieverETCD).ServerKeyPath)
	assert.Equal(t, "client.pem", entry.Retrievers["ut-etcd"].(*CertRetrieverETCD).ClientCertPath)
	assert.Equal(t, "client-key.pem", entry.Retrievers["ut-etcd"].(*CertRetrieverETCD).ClientKeyPath)
	assert.Equal(t, "*::*::*::*", entry.Retrievers["ut-etcd"].(*CertRetrieverETCD).Locale)
	assert.Equal(t, "root:etcd", entry.Retrievers["ut-etcd"].(*CertRetrieverETCD).BasicAuth)
	assert.Equal(t, "localhost:2379", entry.Retrievers["ut-etcd"].(*CertRetrieverETCD).Endpoint)

	// consul
	assert.NotNil(t, entry.Retrievers["ut-consul"])
	assert.Equal(t, "server.pem", entry.Retrievers["ut-consul"].(*CertRetrieverConsul).ServerCertPath)
	assert.Equal(t, "server-key.pem", entry.Retrievers["ut-consul"].(*CertRetrieverConsul).ServerKeyPath)
	assert.Equal(t, "client.pem", entry.Retrievers["ut-consul"].(*CertRetrieverConsul).ClientCertPath)
	assert.Equal(t, "client-key.pem", entry.Retrievers["ut-consul"].(*CertRetrieverConsul).ClientKeyPath)
	assert.Equal(t, "*::*::*::*", entry.Retrievers["ut-consul"].(*CertRetrieverConsul).Locale)
	assert.Equal(t, "root:consul", entry.Retrievers["ut-consul"].(*CertRetrieverConsul).BasicAuth)
	assert.Equal(t, "localhost:8500", entry.Retrievers["ut-consul"].(*CertRetrieverConsul).Endpoint)
	assert.Equal(t, "token", entry.Retrievers["ut-consul"].(*CertRetrieverConsul).Token)

	// remote file store
	assert.NotNil(t, entry.Retrievers["ut-remote-file-store"])
	assert.Equal(t, "server.pem", entry.Retrievers["ut-remote-file-store"].(*CertRetrieverRemoteFileStore).ServerCertPath)
	assert.Equal(t, "server-key.pem", entry.Retrievers["ut-remote-file-store"].(*CertRetrieverRemoteFileStore).ServerKeyPath)
	assert.Equal(t, "client.pem", entry.Retrievers["ut-remote-file-store"].(*CertRetrieverRemoteFileStore).ClientCertPath)
	assert.Equal(t, "client-key.pem", entry.Retrievers["ut-remote-file-store"].(*CertRetrieverRemoteFileStore).ClientKeyPath)
	assert.Equal(t, "*::*::*::*", entry.Retrievers["ut-remote-file-store"].(*CertRetrieverRemoteFileStore).Locale)
	assert.Equal(t, "root:remote", entry.Retrievers["ut-remote-file-store"].(*CertRetrieverRemoteFileStore).BasicAuth)
	assert.Equal(t, "localhost:8080", entry.Retrievers["ut-remote-file-store"].(*CertRetrieverRemoteFileStore).Endpoint)
}

func TestRegisterCertEntry_WithoutOptions(t *testing.T) {
	entry := RegisterCertEntry()

	assert.NotNil(t, entry)
	assert.NotNil(t, GlobalAppCtx.GetCertEntry())

	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.Equal(t, CertEntryName, entry.entryName)
	assert.Equal(t, CertEntryType, entry.entryType)
	assert.Equal(t, len(entry.Stores), 0)
	assert.Equal(t, len(entry.Retrievers), 0)
}

func TestRegisterCertEntry_HappyCase(t *testing.T) {
	zapLoggerEntry := NoopZapLoggerEntry()
	eventLoggerEntry := NoopEventLoggerEntry()
	retriever := &CertRetrieverLocal{
		Name: "ut",
	}

	entry := RegisterCertEntry(
		WithZapLoggerEntryCert(zapLoggerEntry),
		WithEventLoggerEntryCert(eventLoggerEntry),
		WithCertRetrieverCert(retriever))

	assert.NotNil(t, entry)
	assert.NotNil(t, GlobalAppCtx.GetCertEntry())

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
	assert.Equal(t, CertEntryName, entry.entryName)
	assert.Equal(t, CertEntryType, entry.entryType)
	assert.Equal(t, len(entry.Stores), 0)
	assert.Equal(t, len(entry.Retrievers), 1)

	delete(GlobalAppCtx.BasicEntries, "ut")
}

func TestCertEntry_Bootstrap_HappyCase(t *testing.T) {
	retriever := &FakeRetriever{}

	entry := RegisterCertEntry(WithCertRetrieverCert(retriever))

	entry.Bootstrap(context.Background())

	assert.Equal(t, 1, retriever.called)
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
}

func TestCertEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterCertEntry()
	assert.Equal(t, CertEntryName, entry.GetName())
}

func TestCertEntry_GetType_HappyCase(t *testing.T) {
	entry := RegisterCertEntry()
	assert.Equal(t, CertEntryType, entry.GetType())
}

type FakeRetriever struct {
	called int
}

func (retriever *FakeRetriever) Retrieve(context.Context) *CertStore {
	retriever.called++
	return nil
}

func (retriever *FakeRetriever) GetName() string {
	return "fake-retriever"
}

func (retriever *FakeRetriever) GetType() string {
	return "fake-retriever"
}
