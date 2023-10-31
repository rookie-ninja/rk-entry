// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterConfigEntry(t *testing.T) {
	defer assertNotPanic(t)

	// register without file path and content
	entries := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
			},
		},
	})
	assert.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.NotEmpty(t, entries[0].GetDescription())
	assert.NotNil(t, entries[0].Viper)
	assert.Empty(t, entries[0].Viper.AllKeys())

	// register with content
	entries = RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
				Content: map[string]interface{}{
					"content-key": "content-value",
				},
			},
		},
	})
	assert.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.NotEmpty(t, entries[0].GetDescription())
	assert.NotNil(t, entries[0].Viper)
	assert.Equal(t, "content-value", entries[0].GetString("content-key"))

	// register with file
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := filepath.ToSlash(filepath.Join(t.TempDir(), "ut-viper.yaml"))
	assert.Nil(t, os.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))
	entries = RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
				Path:        tempDir,
				Content: map[string]interface{}{
					"content-key": "content-value",
				},
			},
		},
	})
	assert.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.NotEmpty(t, entries[0].GetDescription())
	assert.Equal(t, "content-value", entries[0].GetString("content-key"))
	assert.Equal(t, "value", entries[0].GetString("key"))
}

func TestRegisterConfigEntry_WithNonExistPath(t *testing.T) {
	defer assertNotPanic(t)

	entries := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
				Path:        "/non-exist-ut-file",
			},
		},
	})
	assert.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.NotEmpty(t, entries[0].GetDescription())
	assert.Empty(t, entries[0].AllKeys())
}

func TestRegisterConfigEntry_WithDomainAndFileNotExist(t *testing.T) {
	defer assertNotPanic(t)
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := filepath.ToSlash(filepath.Join(t.TempDir(), "ut-viper.yaml"))
	assert.Nil(t, os.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	entries := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
				Path:        tempDir,
			},
		},
	})
	assert.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.NotEmpty(t, entries[0].GetDescription())
	assert.NotEmpty(t, entries[0].String())
	assert.Equal(t, "value", entries[0].GetString("key"))

	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterConfigEntry_WithDomainAndFileExist(t *testing.T) {
	defer assertNotPanic(t)
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := filepath.ToSlash(filepath.Join(t.TempDir(), "ut-viper.yaml"))
	assert.Nil(t, os.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	entries := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
				Domain:      "prod",
				Path:        tempDir,
			},
		},
	})
	assert.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.NotEmpty(t, entries[0].GetDescription())
	assert.Equal(t, "value", entries[0].GetString("key"))

	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterConfigEntry_WithDomainAndBothFileExist(t *testing.T) {
	defer assertNotPanic(t)

	// create default viper config file named as ut-viper.yaml
	viperConfigBeta := `
---
key: beta
`
	// create viper config file in ut temp dir
	tempDirBeta := filepath.ToSlash(filepath.Join(t.TempDir(), "ut-viper-beta.yaml"))
	assert.Nil(t, os.WriteFile(tempDirBeta, []byte(viperConfigBeta), os.ModePerm))

	// create prod viper config file named as ut-viper-prod.yaml
	viperConfigProd := `
---
key: prod
`
	// create viper config file in ut temp dir
	tempDirProd := filepath.ToSlash(filepath.Join(filepath.Dir(tempDirBeta), "ut-viper-prod.yaml"))
	assert.Nil(t, os.WriteFile(tempDirProd, []byte(viperConfigProd), os.ModePerm))

	// set domain to prod
	assert.Nil(t, os.Setenv("DOMAIN", "prod"))

	entries := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
				Domain:      "beta",
				Path:        tempDirBeta,
			},
			{
				Name:        "ut-config",
				Description: "desc",
				Domain:      "prod",
				Path:        tempDirProd,
			},
		},
	})
	assert.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.NotEmpty(t, entries[0].GetDescription())
	assert.Equal(t, "prod", entries[0].GetString("key"))

	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestRegisterConfigEntriesWithConfig_WithoutDomainAndBothFileExist(t *testing.T) {
	defer assertNotPanic(t)

	// create default viper config file named as ut-viper.yaml
	viperConfig := `
---
key: value
`
	// create viper config file in ut temp dir
	tempDir := filepath.ToSlash(filepath.Join(t.TempDir(), "ut-viper.yaml"))
	assert.Nil(t, os.WriteFile(tempDir, []byte(viperConfig), os.ModePerm))

	// create prod viper config file named as ut-viper-prod.yaml
	viperConfigProd := `
---
key: prod
`
	// create viper config file in ut temp dir
	tempDirProd := filepath.ToSlash(filepath.Join(filepath.Dir(tempDir), "ut-viper-prod.yaml"))
	assert.Nil(t, os.WriteFile(tempDirProd, []byte(viperConfigProd), os.ModePerm))

	entries := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
				Path:        tempDir,
			},
		},
	})
	assert.NotEmpty(t, entries)
	assert.NotEmpty(t, entries[0].GetName())
	assert.NotEmpty(t, entries[0].GetType())
	assert.NotEmpty(t, entries[0].GetDescription())
	assert.Equal(t, "value", entries[0].GetString("key"))

	// unset domain
	assert.Nil(t, os.Setenv("DOMAIN", ""))
}

func TestConfigEntry_UnmarshalJSON(t *testing.T) {
	entry := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
			},
		},
	})
	assert.Nil(t, entry[0].UnmarshalJSON(nil))
}

func TestConfigEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
			},
		},
	})
	entry[0].Bootstrap(context.Background())
}

func TestConfigEntry_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterConfigEntry(&BootConfig{
		Config: []*BootConfigE{
			{
				Name:        "ut-config",
				Description: "desc",
			},
		},
	})
	entry[0].Interrupt(context.Background())
}
