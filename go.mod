module github.com/xn3cr0nx/bitgodine

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/blend/go-sdk v2.0.0+incompatible // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/deckarep/golang-set v1.7.1
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/dgraph-io/dgo/v2 v2.2.0
	github.com/dgraph-io/ristretto v0.0.3
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/edsrzf/mmap-go v1.0.0
	github.com/fatih/color v1.9.0
	github.com/fatih/structs v1.1.0
	github.com/go-redis/redis/v8 v8.4.4
	github.com/gocolly/colly/v2 v2.1.0
	github.com/google/uuid v1.1.1
	github.com/imdario/mergo v0.3.11
	github.com/jinzhu/gorm v1.9.16
	github.com/json-iterator/go v1.1.10
	github.com/labstack/echo/v4 v4.1.17
	github.com/labstack/gommon v0.3.0
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pingcap/tidb v1.1.0-beta.0.20200701091151-ceec9d9c63c8
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/echo-swagger v1.1.0
	github.com/swaggo/swag v1.7.0
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/wcharczuk/go-chart v2.0.2-0.20190910040548-3a7bc5543113+incompatible
	go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho v0.15.1
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.15.0
	go.opentelemetry.io/otel/exporters/stdout v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9 // indirect
	golang.org/x/image v0.0.0-20200801110659-972c09e46d76 // indirect
	golang.org/x/net v0.0.0-20201209123823-ac852fbbde11
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/sys v0.0.0-20201221093633-bc327ba9c2f0 // indirect
	golang.org/x/tools v0.0.0-20201208233053-a543418bbed2 // indirect
	google.golang.org/grpc v1.31.1
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/go-playground/validator.v9 v9.31.0
	gorm.io/driver/postgres v1.0.0
	gorm.io/gorm v1.20.0
)

replace (
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

go 1.15
