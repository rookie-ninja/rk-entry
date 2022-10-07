package rk

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"github.com/rookie-ninja/rk-entry/v3/util"
	"gopkg.in/yaml.v3"
	"sync"
)

const CertKind = "cert"

type CertConfig struct {
	EntryConfigHeader `yaml:",inline"`
	Entry             struct {
		//Api struct {
		//	Enabled   bool   `yaml:"enabled"`
		//	UrlPrefix string `yaml:"urlPrefix"`
		//} `yaml:"api"`
		CA struct {
			LocalPath string `yaml:"localPath"`
			EmbedPath string `yaml:"embedPath"`
			PemBase64 string `yaml:"pemBase64"`
		} `yaml:"ca"`
		Cert struct {
			LocalPath string `yaml:"localPath"`
			EmbedPath string `yaml:"embedPath"`
			PemBase64 string `yaml:"pemBase64"`
		} `yaml:"cert"`
		Key struct {
			LocalPath string `yaml:"localPath"`
			EmbedPath string `yaml:"embedPath"`
			PemBase64 string `yaml:"pemBase64"`
		} `yaml:"key"`
	} `yaml:"entry"`
}

func (c *CertConfig) JSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

func (c *CertConfig) YAML() string {
	b, _ := yaml.Marshal(c)
	return string(b)
}

func (c *CertConfig) Header() *EntryConfigHeader {
	return &c.EntryConfigHeader
}

func (c *CertConfig) Register() (Entry, error) {
	if !c.Metadata.Enabled {
		return nil, nil
	}

	if !rku.IsValidDomain(c.Metadata.Domain) {
		return nil, nil
	}

	entry := &CertEntry{
		config: c,
		once:   sync.Once{},
	}
	entry.embedFS = Registry.EntryFS(entry.Kind(), entry.Name())

	Registry.AddEntry(entry)

	return entry, nil
}

type CertEntry struct {
	config  *CertConfig
	embedFS *embed.FS
	RootCA  *x509.Certificate `json:"-" json:"-"`
	Cert    *tls.Certificate  `json:"-" yaml:"-"`
	once    sync.Once
}

func (c *CertEntry) Category() string {
	return CategoryIndependent
}

func (c *CertEntry) Kind() string {
	return c.config.Kind
}

func (c *CertEntry) Name() string {
	return c.config.Metadata.Name
}

func (c *CertEntry) Config() EntryConfig {
	return c.config
}

func (c *CertEntry) Bootstrap(ctx context.Context) {
	c.once.Do(func() {
		certPem := make([]byte, 0)
		keyPem := make([]byte, 0)

		// 1: cert pem
		if len(c.config.Entry.Cert.PemBase64) > 0 {
			if _, err := base64.StdEncoding.Decode(certPem, []byte(c.config.Entry.Cert.PemBase64)); err != nil {
				rku.ShutdownWithError(err)
			}
		} else if len(c.config.Entry.Cert.EmbedPath) > 0 && c.embedFS != nil {
			v, err := c.embedFS.ReadFile(c.config.Entry.Cert.EmbedPath)
			if err != nil {
				rku.ShutdownWithError(err)
			}
			certPem = v
		} else if len(c.config.Entry.Cert.LocalPath) > 0 {
			certPem = rku.ReadFileFromLocal(c.config.Entry.Cert.LocalPath, true)
		}

		// 2: key pem
		if len(c.config.Entry.Key.PemBase64) > 0 {
			if _, err := base64.StdEncoding.Decode(certPem, []byte(c.config.Entry.Key.PemBase64)); err != nil {
				rku.ShutdownWithError(err)
			}
		} else if len(c.config.Entry.Key.EmbedPath) > 0 && c.embedFS != nil {
			if v, err := c.embedFS.ReadFile(c.config.Entry.Key.EmbedPath); err != nil {
				rku.ShutdownWithError(err)
			} else {
				certPem = v
			}
		} else if len(c.config.Entry.Key.LocalPath) > 0 {
			certPem = rku.ReadFileFromLocal(c.config.Entry.Key.LocalPath, true)
		}

		if v, err := tls.X509KeyPair(certPem, keyPem); err != nil {
			rku.ShutdownWithError(err)
		} else {
			c.Cert = &v
		}

		// 2: root pem
		caPem := make([]byte, 0)
		if len(c.config.Entry.CA.PemBase64) > 0 {
			if _, err := base64.StdEncoding.Decode(caPem, []byte(c.config.Entry.CA.PemBase64)); err != nil {
				rku.ShutdownWithError(err)
			}
		} else if len(c.config.Entry.CA.EmbedPath) > 0 && c.embedFS != nil {
			if v, err := c.embedFS.ReadFile(c.config.Entry.CA.EmbedPath); err != nil {
				rku.ShutdownWithError(err)
			} else {
				caPem = v
			}
		} else if len(c.config.Entry.CA.LocalPath) > 0 {
			caPem = rku.ReadFileFromLocal(c.config.Entry.CA.LocalPath, true)
		}

		block, _ := pem.Decode(caPem)
		if block == nil || block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			return
		}

		if v, err := x509.ParseCertificate(block.Bytes); err != nil {
			rku.ShutdownWithError(err)
		} else {
			c.RootCA = v
		}
	})
}

func (c *CertEntry) Interrupt(ctx context.Context) {}

func (c *CertEntry) Monitor() *Monitor {
	return nil
}

func (c *CertEntry) FS() *embed.FS {
	return c.embedFS
}

func (c *CertEntry) Apis() []*BuiltinApi {
	res := make([]*BuiltinApi, 0)
	return res
}
