package cluster

import (
	"fmt"

	"github.com/gojek/merlin/config"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	sigyaml "sigs.k8s.io/yaml"
)

const (
	K8sUser = "user"
)

type CredsManager interface {
	GenerateConfig() (*rest.Config, error)
	GetClusterName() string
}

type K8sCredsManager struct {
	K8sConfig *config.K8sConfig
}

func NewK8sCredsManager(k *config.K8sConfig) *K8sCredsManager {
	return &K8sCredsManager{K8sConfig: k}
}

func (k *K8sCredsManager) GenerateConfig() (*rest.Config, error) {
	restConf, err := generateKubeConfig(k.K8sConfig)
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.RESTConfigFromKubeConfig(restConf)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (k *K8sCredsManager) GetClusterName() string {
	return k.K8sConfig.Name
}

func generateKubeConfig(c *config.K8sConfig) ([]byte, error) {
	kubeconfig := clientcmdapiv1.Config{
		Clusters: []clientcmdapiv1.NamedCluster{
			{
				Name:    c.Name,
				Cluster: *c.Cluster,
			},
		},
		AuthInfos: []clientcmdapiv1.NamedAuthInfo{
			{
				Name:     K8sUser,
				AuthInfo: *c.AuthInfo,
			},
		},
		Contexts: []clientcmdapiv1.NamedContext{
			{
				Name: fmt.Sprintf("%s-%s", c.Name, K8sUser),
				Context: clientcmdapiv1.Context{
					Cluster:  c.Name,
					AuthInfo: K8sUser,
				},
			},
		},
		CurrentContext: fmt.Sprintf("%s-%s", c.Name, K8sUser),
	}
	r, err := sigyaml.Marshal(kubeconfig)
	if err != nil {
		return nil, err
	}
	return r, nil
}
