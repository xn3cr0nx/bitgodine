package db

import (
	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btcd/wire"

	"github.com/stretchr/testify/suite"
)

type TestDBSuite struct {
	suite.Suite
	level *database.DB
}

func (suite *TestDBSuite) SetupSuite() {
	conf := &Config{
		Dir:  "/tmp",
		Name: "indexing",
		Net:  wire.MainNet,
	}
	suite.level, _ = LevelDB(conf)
}

// func (suite *TestDBSuite) SetupTest() {
// 	suite.clean.Acquire("compliances", "users", "files", "compliance_documents", "compliance_documents_files")
// }

// func (suite *TestDBSuite) TearDownTest() {
// 	suite.clean.Clean("compliances", "users", "files", "compliance_documents", "compliance_documents_files")
// }

// func TestDB(t *testing.T) {
// 	conf := &Config{
// 		"/tmp", "test", wire.MainNet,
// 	}
// 	db, err := LevelDB(conf)
// 	if err != nil {
// 		logger.Error("sync", err, logger.Params{})
// 		return
// 	}
// 	// defer os.RemoveAll(filepath.Join(conf.Dir, conf.Name))
// 	defer (*db).Close()

// }
