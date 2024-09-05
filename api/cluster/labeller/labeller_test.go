package labeller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitKubernetesLabeller(t *testing.T) {
	defaultNsPrefix := "caraml.dev/"
	err := InitKubernetesLabeller("gojek.com/", "caraml.dev/", "dev")
	assert.NoError(t, err)

	defer func() {
		_ = InitKubernetesLabeller("", "", "")
	}()

	tests := []struct {
		prefix  string
		wantErr bool
	}{
		{
			"gojek.com/",
			false,
		},
		{
			"model.caraml.dev/",
			false,
		},
		{
			"goto/gojek",
			true,
		},
		{
			"gojek",
			true,
		},
		{
			"gojek.com/caraml",
			true,
		},
		{
			"gojek//",
			true,
		},
		{
			"gojek.com//",
			true,
		},
		{
			"//gojek.com",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			if err := InitKubernetesLabeller(tt.prefix, defaultNsPrefix, "dev"); (err != nil) != tt.wantErr {
				t.Errorf("InitKubernetesLabeller() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
