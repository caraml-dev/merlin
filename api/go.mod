module github.com/gojek/merlin

go 1.13

require (
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20200424034326-7cd886c7ec44
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/antihax/optional v1.0.0
	github.com/antonmedv/expr v1.8.9
	github.com/buger/jsonparser v1.1.1
	github.com/cespare/xxhash v1.1.0
	github.com/coocood/freecache v1.1.1
	github.com/fatih/color v1.7.0
	github.com/feast-dev/feast/sdk/go v0.9.2
	github.com/ghodss/yaml v1.0.0
	github.com/go-gota/gota v0.10.1
	github.com/go-openapi/spec v0.19.9 // indirect
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator v9.30.0+incompatible
	github.com/gogo/protobuf v1.3.1
	github.com/gojek/heimdall/v7 v7.0.2
	github.com/gojek/merlin-pyspark-app v0.0.3
	github.com/gojek/mlp v0.0.0-20201002030420-4e35e69a9ab8
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/golang/protobuf v1.4.2
	github.com/google/go-containerregistry v0.0.0-20191009212737-d753c5604768
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/schema v1.1.0
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/vault/api v1.0.4
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/jinzhu/gorm v1.9.11
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kubeflow/kfserving v0.4.1
	github.com/mmcloughlin/geohash v0.10.0
	github.com/oliveagle/jsonpath v0.0.0-20180606110733-2e52cf6e6852
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pilagod/gorm-cursor-paginator v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	github.com/xanzy/go-gitlab v0.31.0
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/grpc v1.31.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.2.8
	istio.io/api v0.0.0-20191115173247-e1a1952e5b81
	istio.io/client-go v0.0.0-20191120150049-26c62a04cdbc
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v0.18.3
	knative.dev/pkg v0.0.0-20191217184203-cf220a867b3d
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/gojek/merlin-pyspark-app => ../python/batch-predictor
	k8s.io/api => k8s.io/api v0.16.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.6
	k8s.io/apiserver => k8s.io/apiserver v0.16.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.6
	k8s.io/client-go => k8s.io/client-go v0.16.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.6
	k8s.io/code-generator => k8s.io/code-generator v0.16.6
	k8s.io/component-base => k8s.io/component-base v0.16.6
	k8s.io/cri-api => k8s.io/cri-api v0.16.6
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.6
	k8s.io/kubectl => k8s.io/kubectl v0.16.6
	k8s.io/kubelet => k8s.io/kubelet v0.16.6
	k8s.io/kubernetes => k8s.io/kubernetes v1.16.6
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.6
	k8s.io/metrics => k8s.io/metrics v0.16.6
	k8s.io/node-api => k8s.io/node-api v0.16.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.6
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.16.6
	k8s.io/sample-controller => k8s.io/sample-controller v0.16.6
)
