// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithZapLoggerEntryCred_WithNilLogger(t *testing.T) {
	entry := &CredEntry{
		ZapLoggerEntry: NoopZapLoggerEntry(),
	}

	opt := WithZapLoggerEntryCred(nil)
	opt(entry)

	assert.NotNil(t, entry.ZapLoggerEntry)
}

func TestWithZapLoggerEntryCred_HappyCase(t *testing.T) {
	entry := &CredEntry{}

	loggerEntry := NoopZapLoggerEntry()
	opt := WithZapLoggerEntryCred(loggerEntry)
	opt(entry)

	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestWithEventLoggerEntryCred_WithNilLogger(t *testing.T) {
	entry := &CredEntry{
		EventLoggerEntry: NoopEventLoggerEntry(),
	}

	opt := WithEventLoggerEntryCred(nil)
	opt(entry)

	assert.NotNil(t, entry.EventLoggerEntry)
}

func TestWithEventLoggerEntryCred_HappyCase(t *testing.T) {
	entry := &CredEntry{}

	loggerEntry := NoopEventLoggerEntry()
	opt := WithEventLoggerEntryCred(loggerEntry)
	opt(entry)

	assert.Equal(t, loggerEntry, entry.EventLoggerEntry)
}

func TestWithRetrieverCred_WithNilRetriever(t *testing.T) {
	entry := &CredEntry{}

	opt := WithRetrieverCred(nil)
	opt(entry)

	assert.Nil(t, entry.Retriever)
}

func TestWithRetrieverCred_HappyCase(t *testing.T) {
	entry := &CredEntry{}

	retriever := &CredRetrieverLocalFs{}

	opt := WithRetrieverCred(retriever)
	opt(entry)

	assert.Equal(t, retriever, entry.Retriever)
}

func TestRegisterCredEntriesFromConfig_HappyCase(t *testing.T) {
	bootConfig := `
cred:
  - name: "ut-localFs"
    provider: "localFs"
    locale: "*::*::*::*"
    paths: ["cred.yaml"]
  - name: "ut-etcd"
    provider: "etcd"
    locale: "*::*::*::*"
    endpoint: "localhost:2379"
    basicAuth: "root:etcd"
    paths: ["cred.yaml"]
  - name: "ut-consul"
    provider: "consul"
    locale: "*::*::*::*"
    endpoint: "localhost:8500"
    basicAuth: "root:consul"
    datacenter: "rk"
    token: "token"
    paths: ["cred.yaml"]
  - name: "ut-remoteFs"
    provider: "remoteFs"
    locale: "*::*::*::*"
    endpoint: "localhost:8080"
    basicAuth: "root:remote"
    paths: ["cred.yaml"]
`

	bootPath := createFileAtTestTempDir(t, bootConfig)

	assert.Equal(t, len(RegisterCredEntriesFromConfig(bootPath)), 4)
	entries := GlobalAppCtx.ListCredEntries()
	assert.Equal(t, len(entries), 4)

	// localFS entry
	localFSEntry := entries["ut-localFs"]
	assert.NotNil(t, localFSEntry)
	assert.Equal(t, "ut-localFs", localFSEntry.GetName())
	assert.Equal(t, "localFs", localFSEntry.Retriever.GetProvider())
	assert.Equal(t, "cred.yaml", localFSEntry.Retriever.ListPaths()[0])
	assert.Equal(t, "*::*::*::*", localFSEntry.Retriever.GetLocale())

	// etcd entry
	etcdEntry := entries["ut-etcd"]
	assert.NotNil(t, etcdEntry)
	assert.Equal(t, "ut-etcd", etcdEntry.GetName())
	assert.Equal(t, "etcd", etcdEntry.Retriever.GetProvider())
	assert.Equal(t, "cred.yaml", localFSEntry.Retriever.ListPaths()[0])
	assert.Equal(t, "*::*::*::*", etcdEntry.Retriever.GetLocale())
	assert.Equal(t, "localhost:2379", etcdEntry.Retriever.GetEndpoint())
	assert.Equal(t, "root:etcd", etcdEntry.Retriever.(*CredRetrieverEtcd).BasicAuth)

	// consul entry
	consulEntry := entries["ut-consul"]
	assert.NotNil(t, consulEntry)
	assert.Equal(t, "ut-consul", consulEntry.GetName())
	assert.Equal(t, "consul", consulEntry.Retriever.GetProvider())
	assert.Equal(t, "cred.yaml", localFSEntry.Retriever.ListPaths()[0])
	assert.Equal(t, "*::*::*::*", consulEntry.Retriever.GetLocale())
	assert.Equal(t, "localhost:8500", consulEntry.Retriever.GetEndpoint())
	assert.Equal(t, "root:consul", consulEntry.Retriever.(*CredRetrieverConsul).BasicAuth)
	assert.Equal(t, "token", consulEntry.Retriever.(*CredRetrieverConsul).Token)

	// remote file store entry
	remoteFsEntry := entries["ut-remoteFs"]
	assert.NotNil(t, remoteFsEntry)
	assert.Equal(t, "ut-remoteFs", remoteFsEntry.GetName())
	assert.Equal(t, "remoteFs", remoteFsEntry.Retriever.GetProvider())
	assert.Equal(t, "cred.yaml", localFSEntry.Retriever.ListPaths()[0])
	assert.Equal(t, "*::*::*::*", remoteFsEntry.Retriever.GetLocale())
	assert.Equal(t, "localhost:8080", remoteFsEntry.Retriever.GetEndpoint())
	assert.Equal(t, "root:remote", remoteFsEntry.Retriever.(*CredRetrieverRemoteFs).BasicAuth)

	GlobalAppCtx.clearCredEntries()
}

func TestRegisterCredEntry_WithoutOptions(t *testing.T) {
	entry := RegisterCredEntry()

	assert.NotNil(t, entry)
	entries := GlobalAppCtx.ListCredEntries()

	assert.Equal(t, 1, len(entries))

	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.Equal(t, CredEntryName, entry.EntryName)
	assert.Equal(t, CredEntryType, entry.EntryType)
	assert.Nil(t, entry.Retriever)
	assert.NotNil(t, entry.Store)

	GlobalAppCtx.clearCredEntries()
}

func TestRegisterCredEntry_HappyCase(t *testing.T) {
	zapLoggerEntry := NoopZapLoggerEntry()
	eventLoggerEntry := NoopEventLoggerEntry()
	retriever := &CredRetrieverLocalFs{}

	entry := RegisterCredEntry(
		WithNameCred("ut"),
		WithZapLoggerEntryCred(zapLoggerEntry),
		WithEventLoggerEntryCred(eventLoggerEntry),
		WithRetrieverCred(retriever))

	assert.NotNil(t, entry)

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
	assert.Equal(t, "ut", entry.EntryName)
	assert.Equal(t, CredEntryType, entry.EntryType)
	assert.Equal(t, retriever, entry.Retriever)
	assert.NotNil(t, entry.Store)

	GlobalAppCtx.clearCredEntries()
}

func TestCredEntry_Bootstrap_HappyCase(t *testing.T) {
	retriever := &FakeRetriever{}

	entry := RegisterCredEntry(WithRetrieverCred(retriever))

	entry.Bootstrap(context.Background())

	assert.Equal(t, 1, retriever.called)

	GlobalAppCtx.clearCredEntries()
}

func TestCredEntry_Interrupt_HappyCase(t *testing.T) {
	assertNotPanic(t)
	entry := RegisterCredEntry()

	entry.Interrupt(context.Background())
}

func TestCredEntry_String_HappyCase(t *testing.T) {
	assertNotPanic(t)
	entry := RegisterCredEntry()

	assert.NotEmpty(t, entry.String())

	GlobalAppCtx.clearCredEntries()
}

func TestCredEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterCredEntry()
	assert.Equal(t, CredEntryName, entry.GetName())

	GlobalAppCtx.clearCredEntries()
}

func TestCredEntry_GetType_HappyCase(t *testing.T) {
	entry := RegisterCredEntry()
	assert.Equal(t, CredEntryType, entry.GetType())

	GlobalAppCtx.clearCredEntries()
}
