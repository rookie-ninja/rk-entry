package rkentry

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
