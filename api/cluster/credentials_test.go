package cluster

import (
	"testing"

	"github.com/gojek/merlin/config"
	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/rest"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name      string
		wantError bool
		input     config.K8sConfig
		output    rest.Config
	}{
		{
			name:      "Test basic server, ca cert",
			wantError: false,
			input: config.K8sConfig{
				Name: "dummy-cluster",
				Cluster: &clientcmdapiv1.Cluster{
					Server:                "https://123.456.789",
					InsecureSkipTLSVerify: true,
				},
				AuthInfo: &clientcmdapiv1.AuthInfo{
					ClientCertificateData: []byte(`ABCDEF`),
					ClientKeyData:         []byte(`12345`),
				},
			},
			output: rest.Config{
				Host: "https://123.456.789",
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: true,
					CertData: []byte(`ABCDEF`),
					KeyData:  []byte(`12345`),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			credsM := NewK8sCredsManager(&tt.input)
			res, err := credsM.GenerateConfig()
			if err != nil && !tt.wantError {
				t.Errorf("Error not expected but occured: %s", err.Error())
			}
			if diff := cmp.Diff(*res, tt.output); diff != "" {
				t.Errorf("diff is not empty %s", diff)
			}
		})
	}
}

func TestgenerateKubeConfig(t *testing.T) {

	tests := []struct {
		name      string
		wantError bool
		input     config.K8sConfig
		output    *clientcmdapiv1.Config
	}{
		{
			name:      "Test basic server, ca cert",
			wantError: false,
			input: config.K8sConfig{
				Name: "dummy-cluster",
				Cluster: &clientcmdapiv1.Cluster{
					Server:                "https://123.456.789",
					InsecureSkipTLSVerify: true,
				},
				AuthInfo: &clientcmdapiv1.AuthInfo{
					ClientCertificateData: []byte(`ABCDEF`),
					ClientKeyData:         []byte(`12345`),
				},
			},
			output: &clientcmdapiv1.Config{
				Clusters: []clientcmdapiv1.NamedCluster{
					{
						Name: "dummy-cluster",
						Cluster: clientcmdapiv1.Cluster{
							Server:                "https://123.456.789",
							InsecureSkipTLSVerify: true,
						},
					},
				},
				AuthInfos: []clientcmdapiv1.NamedAuthInfo{
					{
						Name: K8sUser,
						AuthInfo: clientcmdapiv1.AuthInfo{
							ClientCertificateData: []byte(`ABCDEF`),
							ClientKeyData:         []byte(`12345`),
						},
					},
				},
				Contexts: []clientcmdapiv1.NamedContext{
					{
						Name: "dummy-cluster-user",
						Context: clientcmdapiv1.Context{
							Cluster:  "dummy-cluster",
							AuthInfo: K8sUser,
						},
					},
				},
				CurrentContext: "dummy-cluster-user",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := generateKubeConfig(&tt.input)
			if diff := cmp.Diff(*res, tt.output); diff != "" {
				t.Errorf("diff is not empty %s", diff)
			}
		})
	}
}
