package rkentry

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func TestRegisterPromEntry(t *testing.T) {
	boot := &BootProm{
		Enabled: true,
		Path:    "/ut",
	}
	boot.Pusher.Enabled = true
	entry := RegisterPromEntry(boot)

	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON(nil))
	assert.NotNil(t, entry.Registry)
	assert.NotNil(t, entry.Registerer)
	assert.NotNil(t, entry.Gatherer)
	assert.NotNil(t, entry.Pusher)
	entry.RegisterCollectors(collectors.NewBuildInfoCollector())

	// with registry
	boot = &BootProm{
		Enabled: true,
	}
	registry := prometheus.NewRegistry()
	entry = RegisterPromEntry(boot, WithRegistryPromEntry(registry))

	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.Equal(t, registry, entry.Registry)
	assert.Equal(t, registry, entry.Registerer)
	assert.Equal(t, registry, entry.Gatherer)
	assert.Nil(t, entry.Pusher)
}

func TestPromEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	// without pusher
	boot := &BootProm{
		Enabled: true,
	}
	boot.Pusher.Enabled = true
	entry := RegisterPromEntry(boot)
	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())

	// with pusher
	boot.Pusher.Enabled = true

	// assign certs
	caPem, _ := generateCerts(t)
	certPem, keyPem := generateCerts(t)
	caDir := path.Join(t.TempDir(), "ca.pem")
	certPemDir := path.Join(t.TempDir(), "cert.pem")
	keyPemDir := path.Join(t.TempDir(), "key.pem")
	assert.Nil(t, ioutil.WriteFile(caDir, caPem, os.ModePerm))
	assert.Nil(t, ioutil.WriteFile(certPemDir, certPem, os.ModePerm))
	assert.Nil(t, ioutil.WriteFile(keyPemDir, keyPem, os.ModePerm))
	RegisterCertEntry(&BootCert{
		Cert: []*BootCertE{
			{
				Name:        "ut-cert",
				KeyPemPath:  keyPemDir,
				CertPemPath: certPemDir,
				CAPath:      caDir,
			},
		},
	})

	boot.Pusher.CertEntry = "ut-cert"

	entry = RegisterPromEntry(boot)
	entry.Bootstrap(context.TODO())
	time.Sleep(3 * time.Second)

	entry.Interrupt(context.TODO())
}
