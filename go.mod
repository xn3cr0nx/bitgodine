module github.com/xn3cr0nx/bitgodine

require (
	github.com/blend/go-sdk v2.0.0+incompatible // indirect
	github.com/boltdb/bolt v1.3.1
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/deckarep/golang-set v1.7.1
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/dgraph-io/dgo/v2 v2.2.0
	github.com/dgraph-io/ristretto v0.0.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/go-farm v0.0.0-20200201041132-a6ae2369ad13 // indirect
	github.com/edsrzf/mmap-go v1.0.0
	github.com/fatih/color v1.9.0
	github.com/fatih/structs v1.1.0
	github.com/gocolly/colly/v2 v2.1.0
	github.com/jinzhu/gorm v1.9.14
	github.com/json-iterator/go v1.1.10
	github.com/labstack/echo/v4 v4.1.16
	github.com/labstack/gommon v0.3.0
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/olekukonko/tablewriter v0.0.4
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/pingcap/tidb v1.1.0-beta.0.20200701091151-ceec9d9c63c8
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	github.com/wcharczuk/go-chart v2.0.2-0.20190910040548-3a7bc5543113+incompatible
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/image v0.0.0-20200618115811-c13761719519 // indirect
	golang.org/x/net v0.0.0-20200602114024-627f9648deb9
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	golang.org/x/text v0.3.3 // indirect
	google.golang.org/grpc v1.27.0
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/go-playground/validator.v9 v9.31.0
)

replace (
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)


go 1.14
