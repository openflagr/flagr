module github.com/openflagr/flagr

go 1.25.2

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/DataDog/datadog-go v4.8.3+incompatible
	github.com/Shopify/sarama v1.38.0
	github.com/a8m/kinesis-producer v0.2.1-0.20240916120605-7e5fa98d32a5
	github.com/auth0/go-jwt-middleware v1.0.2-0.20210804140707-b4090e955b98
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/brandur/simplebox v0.1.0
	github.com/bsm/ratelimit v2.0.0+incompatible
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/certifi/gocertifi v0.0.0-20210507211836-431795d63e8d // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dchest/uniuri v1.2.0
	github.com/evalphobia/logrus_sentry v0.8.2
	github.com/form3tech-oss/jwt-go v3.2.5+incompatible
	github.com/getsentry/raven-go v0.2.0
	github.com/go-openapi/analysis v0.24.0 // indirect
	github.com/go-openapi/errors v0.22.4
	github.com/go-openapi/jsonreference v0.21.3 // indirect
	github.com/go-openapi/loads v0.23.2
	github.com/go-openapi/runtime v0.29.0
	github.com/go-openapi/spec v0.22.1
	github.com/go-openapi/strfmt v0.25.0
	github.com/go-openapi/swag v0.25.1
	github.com/go-openapi/validate v0.25.0
	github.com/gohttp/pprof v0.0.0-20141119085724-c9d246cbb3ba
	github.com/jessevdk/go-flags v1.5.0
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/meatballhat/negroni-logrus v1.1.1
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/phyber/negroni-gzip v1.0.0
	github.com/prashantv/gostub v1.1.0
	github.com/prometheus/client_golang v1.20.4
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rs/cors v1.11.1
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cast v1.10.0
	github.com/stretchr/testify v1.11.1
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/urfave/negroni v1.0.0
	github.com/yadvendar/negroni-newrelic-go-agent v0.0.0-20160803090806-3dc58758cb67
	github.com/zhouzhuojie/conditions v0.2.3
	github.com/zhouzhuojie/withtimeout v0.0.0-20190405051827-12b39eb2edd5
	golang.org/x/net v0.47.0
	google.golang.org/api v0.247.0
	google.golang.org/grpc v1.74.2
	gopkg.in/DataDog/dd-trace-go.v1 v1.46.0
)

require (
	cloud.google.com/go/pubsub v1.49.0
	github.com/glebarez/sqlite v1.6.0
	github.com/newrelic/go-agent v2.1.0+incompatible
	gorm.io/driver/mysql v1.4.5
	gorm.io/driver/postgres v1.4.6
	gorm.io/gorm v1.24.3 // we will need to fix unscoped preload before upgrading gorm
)

require (
	github.com/aws/aws-sdk-go-v2/config v1.31.20
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.42.3
)

require (
	cloud.google.com/go/auth v0.16.4 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.8.0 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.41.1 // indirect
	github.com/DataDog/datadog-agent/pkg/remoteconfig/state v0.42.0-rc.5 // indirect
	github.com/DataDog/datadog-go/v5 v5.2.0 // indirect
	github.com/DataDog/go-tuf v0.3.0--fix-localmeta-fork // indirect
	github.com/DataDog/sketches-go v1.4.1 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.39.6 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.24 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.40.2 // indirect
	github.com/aws/smithy-go v1.23.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/eapache/go-resiliency v1.3.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230111030713-bf00bc1b83b6 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/glebarez/go-sqlite v1.21.2 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.22.1 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.1 // indirect
	github.com/go-openapi/swag/conv v0.25.1 // indirect
	github.com/go-openapi/swag/fileutils v0.25.1 // indirect
	github.com/go-openapi/swag/jsonname v0.25.1 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.1 // indirect
	github.com/go-openapi/swag/loading v0.25.1 // indirect
	github.com/go-openapi/swag/mangling v0.25.1 // indirect
	github.com/go-openapi/swag/netutils v0.25.1 // indirect
	github.com/go-openapi/swag/stringutils v0.25.1 // indirect
	github.com/go-openapi/swag/typeutils v0.25.1 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.1 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.4 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.3 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.4.0 // indirect
	go.einride.tech/aip v0.68.1 // indirect
	go.mongodb.org/mongo-driver v1.17.6 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go4.org/intern v0.0.0-20220617035311-6925f38cc365 // indirect
	go4.org/unsafe/assume-no-moving-gc v0.0.0-20220617031537-928513b29760 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/genproto v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250721164621-a45f3dfb1074 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	inet.af/netaddr v0.0.0-20220811202034-502d2d690317 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.23.1 // indirect
)
