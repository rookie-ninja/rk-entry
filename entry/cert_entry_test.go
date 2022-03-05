package rkentry

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"testing"
	"time"
)

func TestRegisterCertEntry(t *testing.T) {
	// without name
	assert.Empty(t, RegisterCertEntry(&BootCert{
		Cert: []*BootCertE{
			{
				Name: "",
			},
		},
	}))

	// happy case
	entries := RegisterCertEntry(&BootCert{
		Cert: []*BootCertE{
			{
				Name: "ut-cert",
			},
		},
	})
	assert.Len(t, entries, 1)
}

func TestRegisterCertEntry_FromYAML(t *testing.T) {
	bootStr := `
---
cert:
  - name: ut-cert
    caPath: /ut-ca
    keyPemPath: /ut-key
    certPemPath: /ut-cert
`

	entries := RegisterCertEntryYAML([]byte(bootStr))
	assert.Len(t, entries, 1)

	entry := entries["ut-cert"].(*CertEntry)
	assert.Equal(t, "/ut-ca", entry.caPath)
	assert.Equal(t, "/ut-key", entry.keyPemPath)
	assert.Equal(t, "/ut-cert", entry.certPemPath)
	assert.Equal(t, "ut-cert", entry.GetName())
	assert.Equal(t, CertEntryType, entry.GetType())
	assert.Empty(t, entry.GetDescription())
	assert.Empty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
}

func TestCertEntry_Bootstrap(t *testing.T) {
	//defer assertNotPanic(t)

	// create root & key & cert pem
	caPem, _ := generateCerts(t)
	certPem, keyPem := generateCerts(t)

	caDir := path.Join(t.TempDir(), "ca.pem")
	certPemDir := path.Join(t.TempDir(), "cert.pem")
	keyPemDir := path.Join(t.TempDir(), "key.pem")

	assert.Nil(t, ioutil.WriteFile(caDir, caPem, os.ModePerm))
	assert.Nil(t, ioutil.WriteFile(certPemDir, certPem, os.ModePerm))
	assert.Nil(t, ioutil.WriteFile(keyPemDir, keyPem, os.ModePerm))

	entries := RegisterCertEntry(&BootCert{
		Cert: []*BootCertE{
			{
				Name:        "ut-cert",
				KeyPemPath:  keyPemDir,
				CertPemPath: certPemDir,
				CAPath:      caDir,
			},
		},
	})
	assert.Len(t, entries, 1)

	entry := entries[0]
	entry.Bootstrap(context.TODO())

	assert.NotNil(t, entry.Certificate)
	assert.NotNil(t, entry.RootCA)

	entry.Interrupt(context.TODO())
}

func TestCertEntry_UnmarshalJSON(t *testing.T) {
	entries := RegisterCertEntry(&BootCert{
		Cert: []*BootCertE{
			{
				Name: "cert",
			},
		},
	})
	assert.Nil(t, entries[0].UnmarshalJSON(nil))
}

func generateCerts(t *testing.T) ([]byte, []byte) {
	// Create certs and return as []byte
	ca := &x509.Certificate{
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		SerialNumber:          big.NewInt(42),
		NotAfter:              time.Now().Add(2 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Create a Private Key
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	assert.Nil(t, err)

	// Use CA Cert to sign a CSR and create a Public Cert
	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)
	assert.Nil(t, err)

	// Convert keys into pem.Block
	c := &pem.Block{Type: "CERTIFICATE", Bytes: cert}
	k := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}

	return pem.EncodeToMemory(c), pem.EncodeToMemory(k)
}
