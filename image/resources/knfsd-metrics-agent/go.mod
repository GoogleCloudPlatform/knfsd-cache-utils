module github.com/GoogleCloudPlatform/knfsd-cache-utils/image/resources/knfsd-metrics-agent

go 1.20

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/cmd/mdatagen v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticsearchexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/googlecloudpubsubexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/influxdbexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/opencensusexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusremotewriteexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/stackdriverexporter v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.44.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.44.0
	github.com/prometheus/procfs v0.10.1
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/collector v0.44.0
	go.opentelemetry.io/collector/model v0.44.0
	go.uber.org/multierr v1.8.0
	go.uber.org/zap v1.20.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	cloud.google.com/go/compute v1.20.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/monitoring v1.13.0 // indirect
	cloud.google.com/go/trace v1.9.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.4.0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.10 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.0.0 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/antonmedv/expr v1.9.0 // indirect
	github.com/armon/go-metrics v0.3.10 // indirect
	github.com/aws/aws-sdk-go v1.42.44 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/census-instrumentation/opencensus-proto v0.4.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/elastic/go-elasticsearch/v7 v7.16.0 // indirect
	github.com/elastic/go-structform v0.0.9 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.1 // indirect
	github.com/go-logr/stdr v1.2.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.10.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/s2a-go v0.1.4 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.11.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.3 // indirect
	github.com/hashicorp/consul/api v1.12.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.0.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.9.6 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/influxdata/influxdb-observability/common v0.2.10 // indirect
	github.com/influxdata/influxdb-observability/otel2influx v0.2.10 // indirect
	github.com/influxdata/line-protocol/v2 v2.2.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/knadh/koanf v1.4.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.1.16 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/opencensus v0.44.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheusremotewrite v0.44.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_golang v1.12.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/prometheus v1.8.2-0.20220117154355-4855a0c067e2 // indirect
	github.com/prometheus/statsd_exporter v0.21.0 // indirect
	github.com/rs/cors v1.8.2 // indirect
	github.com/santhosh-tekuri/jsonschema v1.2.4 // indirect
	github.com/shirou/gopsutil/v3 v3.21.12 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/cobra v1.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.1 // indirect
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/tinylru v1.1.0 // indirect
	github.com/tidwall/wal v1.1.7 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.elastic.co/apm v1.15.0 // indirect
	go.elastic.co/fastjson v1.1.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.28.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.28.0 // indirect
	go.opentelemetry.io/contrib/zpages v0.28.0 // indirect
	go.opentelemetry.io/otel v1.3.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.26.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.26.0 // indirect
	go.opentelemetry.io/otel/metric v0.26.0 // indirect
	go.opentelemetry.io/otel/sdk v1.3.0 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.26.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.3.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/net v0.15.0 // indirect
	golang.org/x/oauth2 v0.12.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/term v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect; indirect	google.golang.org/api v0.66.0 // indirect
	google.golang.org/api v0.126.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230530153820-e85fd2cbaebc // indirect
	google.golang.org/grpc v1.55.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.23.3 // indirect
	k8s.io/apimachinery v0.23.3 // indirect
	k8s.io/client-go v0.23.3 // indirect
	k8s.io/klog/v2 v2.40.1 // indirect
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65 // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
