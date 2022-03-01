module github.com/gojek/merlin

go 1.13

require (
	cloud.google.com/go/bigtable v1.11.0
	cloud.google.com/go/compute v1.5.0 // indirect
	cloud.google.com/go/iam v0.2.0 // indirect
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20220214044918-55732a6a392c
	github.com/HdrHistogram/hdrhistogram-go v1.0.1 // indirect
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/antihax/optional v1.0.0
	github.com/antonmedv/expr v1.8.9
	github.com/buger/jsonparser v1.1.1
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/cespare/xxhash v1.1.0
	github.com/coocood/freecache v1.1.1
	github.com/fatih/color v1.7.0
	github.com/feast-dev/feast/sdk/go v0.9.2
	github.com/ghodss/yaml v1.0.0
	github.com/go-gota/gota v0.10.1
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-openapi/spec v0.19.9 // indirect
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator v9.30.0+incompatible
	github.com/go-redis/redis/v8 v8.11.4
	github.com/gogo/protobuf v1.3.2
	github.com/gojek/heimdall/v7 v7.0.2
	github.com/gojek/merlin-pyspark-app v0.0.3
	github.com/gojek/mlp v0.0.0-20201002030420-4e35e69a9ab8
	github.com/gojekfarm/jsonpath v0.1.0
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/golang/protobuf v1.5.2
	github.com/google/go-containerregistry v0.7.1-0.20211118220127-abdc633f8305
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.1.0
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/vault/api v1.0.4
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/jinzhu/gorm v1.9.11
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kserve/kserve v0.8.0
	github.com/linkedin/goavro/v2 v2.10.1
	github.com/mmcloughlin/geohash v0.10.0
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/opentracing-contrib/go-stdlib v1.0.0
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pilagod/gorm-cursor-paginator v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.12.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.32.1
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/robfig/cron v1.2.0
	github.com/rs/cors v1.7.0
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.7.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible
	github.com/xanzy/go-gitlab v0.31.0
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	google.golang.org/api v0.70.0
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	istio.io/api v0.0.0-20200715212100-dbf5277541ef
	istio.io/client-go v0.0.0-20201005161859-d8818315d678
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/klog/v2 v2.40.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	knative.dev/pkg v0.0.0-20211206113427-18589ac7627e
	sigs.k8s.io/controller-runtime v0.11.0
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/gojek/merlin-pyspark-app => ../python/batch-predictor
	github.com/googleapis/gnostic => github.com/google/gnostic v0.5.5
	github.com/prometheus/tsdb => github.com/prometheus/tsdb v0.3.1
	k8s.io/api => k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.3
	k8s.io/apiserver => k8s.io/apiserver v0.21.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.3
	k8s.io/client-go => k8s.io/client-go v0.21.3
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.3
	k8s.io/code-generator => k8s.io/code-generator v0.21.3
	k8s.io/component-base => k8s.io/component-base v0.21.3
	k8s.io/cri-api => k8s.io/cri-api v0.21.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.3
	k8s.io/kubectl => k8s.io/kubectl v0.21.3
	k8s.io/kubelet => k8s.io/kubelet v0.21.3
	k8s.io/kubernetes => k8s.io/kubernetes v1.16.6
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.3
	k8s.io/metrics => k8s.io/metrics v0.21.3
	k8s.io/node-api => k8s.io/node-api v0.21.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.3
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.21.3
	k8s.io/sample-controller => k8s.io/sample-controller v0.21.3
)
