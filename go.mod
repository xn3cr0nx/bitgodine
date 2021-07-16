module github.com/xn3cr0nx/bitgodine

require (
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/PuerkitoBio/goquery v1.6.0 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/antchfx/xmlquery v1.3.3 // indirect
	github.com/antchfx/xpath v1.1.11 // indirect
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/blend/go-sdk v2.0.0+incompatible // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/deckarep/golang-set v1.7.1
	github.com/dgraph-io/badger/v3 v3.2103.1
	github.com/dgraph-io/dgo/v2 v2.2.0
	github.com/dgraph-io/ristretto v0.1.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/edsrzf/mmap-go v1.0.0
	github.com/fatih/color v1.10.0
	github.com/fatih/structs v1.1.0
	github.com/frankban/quicktest v1.13.0 // indirect
	github.com/go-openapi/spec v0.20.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1
	github.com/go-redis/redis/v8 v8.4.4
	github.com/gocolly/colly/v2 v2.1.0
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/uuid v1.1.2
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/imdario/mergo v0.3.11
	github.com/jackc/pgmock v0.0.0-20201204152224-4fe30f7445fd // indirect
	github.com/jinzhu/gorm v1.9.16
	github.com/json-iterator/go v1.1.10
	github.com/labstack/echo/v4 v4.1.17
	github.com/labstack/gommon v0.3.0
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.9.0
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.0 // indirect
	github.com/nxadm/tail v1.4.6 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.4
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.9.0 // indirect
	github.com/robfig/cron/v3 v3.0.1
	github.com/segmentio/kafka-go v0.4.8
	github.com/sendgrid/rest v2.6.2+incompatible
	github.com/sendgrid/sendgrid-go v3.7.2+incompatible
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/echo-swagger v1.0.0
	github.com/swaggo/swag v1.6.7
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/wcharczuk/go-chart v2.0.2-0.20190910040548-3a7bc5543113+incompatible
	go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho v0.15.1
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.15.0
	go.opentelemetry.io/otel/exporters/stdout v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	golang.org/x/image v0.0.0-20201208152932-35266b937fa6 // indirect
	golang.org/x/mod v0.4.0 // indirect
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b // indirect
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	google.golang.org/api v0.36.0 // indirect
	google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d // indirect
	google.golang.org/grpc v1.34.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/ini.v1 v1.62.0 // indirect
	gorm.io/driver/postgres v1.0.6
	gorm.io/gorm v1.20.9
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.4
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	github.com/uber-go/atomic => go.uber.org/atomic v1.4.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

go 1.15
