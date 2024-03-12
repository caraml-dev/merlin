module github.com/caraml-dev/merlin

go 1.22

require (
	cloud.google.com/go/bigtable v1.11.0
	cloud.google.com/go/storage v1.39.0
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20221025152940-c261df66a006
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/antihax/optional v1.0.0
	github.com/antonmedv/expr v1.12.5
	github.com/bboughton/gcp-helpers v0.1.0
	github.com/buger/jsonparser v1.1.1
	github.com/caraml-dev/merlin-pyspark-app v0.0.3
	github.com/caraml-dev/mlp v1.12.0
	github.com/caraml-dev/protopath v0.1.0
	github.com/cenkalti/backoff/v4 v4.2.1
	github.com/caraml-dev/universal-prediction-interface v1.0.0
	github.com/cespare/xxhash v1.1.0
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2
	github.com/coocood/freecache v1.2.3
	github.com/fatih/color v1.15.0
	github.com/feast-dev/feast/sdk/go v0.9.4
	github.com/fraugster/parquet-go v0.10.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-coldbrew/grpcpool v0.0.0-20230414075243-75e5835d29e8
	github.com/go-gota/gota v0.0.0-00010101000000-000000000000
	github.com/go-playground/locales v0.14.0
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator/v10 v10.11.1
	github.com/go-redis/redis/v8 v8.11.4
	github.com/gogo/protobuf v1.3.2
	github.com/gojekfarm/jsonpath v0.1.1
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/golang/geo v0.0.0-20210211234256-740aa86cb551
	github.com/golang/protobuf v1.5.3
	github.com/google/go-cmp v0.6.0
	github.com/google/go-containerregistry v0.19.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.1.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jedib0t/go-pretty/v6 v6.5.3
	github.com/jinzhu/copier v0.3.5
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kserve/kserve v0.12.0
	github.com/linkedin/goavro/v2 v2.11.1
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mmcloughlin/geohash v0.10.0
	github.com/newrelic/newrelic-client-go/v2 v2.17.0
	github.com/opentracing-contrib/go-grpc v0.0.0-20200813121455-4a6760c71486
	github.com/opentracing/opentracing-go v1.2.0
	github.com/ory/viper v1.7.5
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pilagod/gorm-cursor-paginator/v2 v2.4.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.19.0
	github.com/prometheus/client_model v0.6.0
	github.com/prometheus/prometheus v0.50.1
	github.com/robfig/cron v1.2.0
	github.com/rs/cors v1.8.2
	github.com/soheilhy/cmux v0.1.5
	github.com/spaolacci/murmur3 v1.1.0
	github.com/stretchr/testify v1.8.4
	github.com/xanzy/go-gitlab v0.32.0
	go.opencensus.io v0.24.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.49.0
	go.opentelemetry.io/contrib/propagators/b3 v1.17.0
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/exporters/jaeger v1.16.0
	go.opentelemetry.io/otel/sdk v1.22.0
	go.opentelemetry.io/otel/trace v1.24.0
	go.uber.org/automaxprocs v1.5.3
	go.uber.org/zap v1.27.0
	golang.org/x/exp v0.0.0-20240119083558-1b970713d09a
	golang.org/x/text v0.14.0
	google.golang.org/api v0.169.0
	google.golang.org/grpc v1.62.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/postgres v1.4.4
	gorm.io/gorm v1.24.0
	istio.io/api v1.19.4
	istio.io/client-go v1.19.4
	k8s.io/api v0.29.2
	k8s.io/apimachinery v0.29.2
	k8s.io/client-go v0.29.2
	knative.dev/networking v0.0.0-20240116081125-ce0738abf051
	knative.dev/pkg v0.0.0-20240311075308-106b052bd1a4
	knative.dev/serving v0.40.1
	sigs.k8s.io/controller-runtime v0.17.2
	sigs.k8s.io/yaml v1.4.0
)

require (
	cloud.google.com/go v0.112.0 // indirect
	cloud.google.com/go/compute v1.24.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v1.1.6 // indirect
	cloud.google.com/go/longrunning v0.5.5 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.7.1-0.20200907061046-05415f1de66d // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.9.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.5.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.1 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.2.1 // indirect
	github.com/Microsoft/go-winio v0.6.1 // indirect
	github.com/VividCortex/ewma v1.2.0 // indirect
	github.com/alecthomas/units v0.0.0-20231202071711-9a357b53e9c9 // indirect
	github.com/antlr/antlr4/runtime/Go/antlr v0.0.0-20220804214150-8b0cc382067f // indirect
	github.com/apache/thrift v0.16.0 // indirect
	github.com/aws/aws-sdk-go v1.50.0 // indirect
	github.com/bboreham/go-loser v0.0.0-20230920113527-fcc2c21820a3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/certifi/gocertifi v0.0.0-20191021191039-0944d244cd40 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cheggaaa/pb/v3 v3.0.8 // indirect
	github.com/containerd/stargz-snapshotter/estargz v0.14.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/varint v1.0.0 // indirect
	github.com/dgraph-io/ristretto v0.0.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/cli v24.0.0+incompatible // indirect
	github.com/docker/distribution v2.8.2+incompatible // indirect
	github.com/docker/docker v25.0.0+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.7.0 // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/evanphx/json-patch v5.7.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.0 // indirect
	github.com/facette/natsort v0.0.0-20181210072756-2cd4dd1e2dcb // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.3.1 // indirect
	github.com/go-git/go-git/v5 v5.3.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.2 // indirect
	github.com/googleapis/google-cloud-go-testing v0.0.0-20210719221736-1c9a4c676720 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/grafana/regexp v0.0.0-20221122212121-6b5c0a4cb7fd // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.4 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.4 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kevinburke/ssh_config v1.1.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.10.2 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/newrelic/go-agent v3.19.2+incompatible // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0-rc3 // indirect
	github.com/ory/keto-client-go v0.11.0-alpha.0 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/princjef/gomarkdoc v0.4.1 // indirect
	github.com/princjef/mageutil v1.0.0 // indirect
	github.com/prometheus/common v0.49.0 // indirect
	github.com/prometheus/common/sigv4 v0.1.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/prometheus/statsd_exporter v0.25.0 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/assertions v1.0.0 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.8.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.8.1 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tomnomnom/linkheader v0.0.0-20180905144013-02ca5825eb80 // indirect
	github.com/valyala/fastjson v1.6.3 // indirect
	github.com/vbatts/tar-split v0.11.3 // indirect
	github.com/x-cray/logrus-prefixed-formatter v0.5.2 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.49.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/goleak v1.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.21.0 // indirect
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/net v0.22.0 // indirect
	golang.org/x/oauth2 v0.18.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/term v0.18.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	gonum.org/v1/gonum v0.11.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20240213162025-012b6fc9bca9 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240221002015-b0ce06bbee7c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240304161311-37d4d3c04a78 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	mvdan.cc/xurls/v2 v2.2.0 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
)

require (
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220124234850-424119656bbf // indirect
	k8s.io/utils v0.0.0-20221107191617-1a15be271d1d // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	github.com/caraml-dev/merlin-pyspark-app => ../python/batch-predictor
	github.com/go-gota/gota => github.com/gojekfarm/gota v0.12.1-0.20230221101638-6cd9260bd598
	github.com/googleapis/gnostic => github.com/google/gnostic v0.5.5
	github.com/prometheus/tsdb => github.com/prometheus/tsdb v0.3.1
)
