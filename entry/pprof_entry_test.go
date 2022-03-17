package rkentry

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegisterPProfEntry(t *testing.T) {
	entry := RegisterPProfEntry(&BootPProf{
		Enabled: true,
		Path:    "ut-path",
	})

	assert.Equal(t, "/ut-path/", entry.Path)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
}

func TestPProfEntry_Bootstrap_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterPProfEntry(&BootPProf{
		Enabled: true,
		Path:    "ut-path",
	})

	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestPProfEntry_UnmarshalJSON(t *testing.T) {
	entry := RegisterPProfEntry(&BootPProf{
		Enabled: true,
		Path:    "ut-path",
	})
	assert.Nil(t, entry.UnmarshalJSON(nil))
}
