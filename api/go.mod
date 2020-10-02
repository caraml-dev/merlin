module github.com/gojek/merlin

go 1.13

require (
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20200311173242-aae36546e51e
	github.com/antihax/optional v1.0.0
	github.com/emicklei/go-restful v2.10.0+incompatible // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/go-openapi/spec v0.19.9 // indirect
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator v9.30.0+incompatible
	github.com/gogo/protobuf v1.3.1
	github.com/gojek/merlin-pyspark-app v0.0.3
	github.com/gojek/mlp v0.0.0-20201002030420-4e35e69a9ab8
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/google/go-containerregistry v0.0.0-20191009212737-d753c5604768
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/schema v1.1.0
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/vault/api v1.0.4
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jinzhu/copier v0.0.0-20190924061706-b57f9002281a
	github.com/jinzhu/gorm v1.9.11
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kubeflow/kfserving v0.1.4-0.20191009025351-282def3efc0b
	github.com/mattbaird/jsonpatch v0.0.0-20171005235357-81af80346b1a // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	github.com/prometheus/common v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.4.0
	github.com/xanzy/go-gitlab v0.31.0
	go.uber.org/zap v1.13.0
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	gopkg.in/yaml.v2 v2.2.7
	istio.io/api v0.0.0-20191024002041-d00922a1ff07
	k8s.io/api v0.0.0-20181126151915-b503174bad59
	k8s.io/apiextensions-apiserver v0.0.0-20181126155829-0cd23ebeb688 // indirect
	k8s.io/apimachinery v0.0.0-20181126123746-eddba98df674
	k8s.io/client-go v0.0.0-20181126152608-d082d5923d3c
	k8s.io/klog v1.0.0 // indirect
	knative.dev/pkg v0.0.0-20191009183213-9c320664c8c4
	knative.dev/serving v0.8.0 // indirect
	sigs.k8s.io/controller-runtime v0.1.9 // indirect
	sigs.k8s.io/testing_frameworks v0.1.2 // indirect
)

replace github.com/gojek/merlin-pyspark-app => ../python/batch-predictor
