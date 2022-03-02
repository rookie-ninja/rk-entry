// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkentry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/json"
	"encoding/pem"
)

const CertEntryType = "CertEntry"

// RegisterCertEntry create cert entry with options.
func RegisterCertEntry(boot *BootCert) []*CertEntry {
	res := make([]*CertEntry, 0)

	for _, cert := range boot.Cert {
		if len(cert.Locale) < 1 {
			cert.Locale = "*::*::*::*"
		}

		if len(cert.Name) < 1 || !IsLocaleValid(cert.Locale) {
			continue
		}

		entry := &CertEntry{
			entryName:        cert.Name,
			entryType:        CertEntryType,
			entryDescription: cert.Description,
			rootPemPath:      cert.RootPemPath,
			keyPemPath:       cert.KeyPemPath,
			certPemPath:      cert.CertPemPath,
		}

		GlobalAppCtx.AddEntry(entry)
		res = append(res, entry)
	}

	return res
}

// registerCertEntry register function
func registerCertEntry(raw []byte) map[string]Entry {
	boot := &BootCert{}
	UnmarshalBoot(raw, boot)

	res := map[string]Entry{}

	entries := RegisterCertEntry(boot)
	for i := range entries {
		entry := entries[i]
		res[entry.GetName()] = entry
	}

	return res
}

// BootCert is bootstrap config of CertEntry.
type BootCert struct {
	Cert []*BootCertE `yaml:"cert" json:"cert"`
}

type BootCertE struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Locale      string `yaml:"locale" json:"locale"`
	RootPemPath string `yaml:"rootPemPath" json:"rootPemPath"`
	CertPemPath string `yaml:"certPemPath" json:"certPemPath"`
	KeyPemPath  string `yaml:"keyPemPath" json:"keyPemPath"`
}

// CertEntry contains bellow fields.
type CertEntry struct {
	entryName        string            `json:"-" yaml:"-"`
	entryType        string            `json:"-" yaml:"-"`
	entryDescription string            `json:"-" yaml:"-"`
	rootPemPath      string            `json:"-" yaml:"-"`
	keyPemPath       string            `json:"-" yaml:"-"`
	certPemPath      string            `json:"-" yaml:"-"`
	embedFS          *embed.FS         `json:"-" yaml:"-"`
	RootCA           *x509.Certificate `json:"-" json:"-"`
	Certificate      *tls.Certificate  `json:"-" yaml:"-"`
}

func (entry *CertEntry) SetEmbedFS(fs *embed.FS) {
	entry.embedFS = fs
}

// Bootstrap iterate retrievers and call Retrieve() for each of them.
func (entry *CertEntry) Bootstrap(ctx context.Context) {
	// server cert path
	if len(entry.keyPemPath) > 0 && len(entry.certPemPath) > 0 {
		cert, err := tls.X509KeyPair(
			readFile(entry.certPemPath, entry.embedFS),
			readFile(entry.keyPemPath, entry.embedFS))
		if err != nil {
			ShutdownWithError(err)
		}

		entry.Certificate = &cert
	}

	if len(entry.rootPemPath) > 0 {
		block, _ := pem.Decode(readFile(entry.rootPemPath, entry.embedFS))
		if block == nil || block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			return
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			ShutdownWithError(err)
		}

		entry.RootCA = cert
	}
}

// Interrupt entry.
func (entry *CertEntry) Interrupt(context.Context) {}

// String return string of entry.
func (entry *CertEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// MarshalJSON marshal entry
func (entry *CertEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"name":        entry.entryName,
		"type":        entry.entryType,
		"description": entry.entryDescription,
		"rootPemPath": entry.rootPemPath,
		"keyPemPath":  entry.keyPemPath,
		"certPemPath": entry.certPemPath,
	}

	return json.Marshal(&m)
}

// UnmarshalJSON unmarshal entry
func (entry *CertEntry) UnmarshalJSON([]byte) error {
	return nil
}

// GetName return name of entry.
func (entry *CertEntry) GetName() string {
	return entry.entryName
}

// GetType return type of entry.
func (entry *CertEntry) GetType() string {
	return entry.entryType
}

// GetDescription return description of entry
func (entry *CertEntry) GetDescription() string {
	return entry.entryDescription
}
