package pgtest

import (
	dbpg "github.com/tdex-network/tdex-analytics/internal/infrastructure/db/pg"

	"github.com/stretchr/testify/suite"
)

const (
	host     = "127.0.0.1"
	port     = 5432
	user     = "root"
	password = "secret"
	dbname   = "tdexa-test"
)

var (
	pgDbSvc dbpg.Service
)

type PgDbTestSuite struct {
	suite.Suite
}

func (s *PgDbTestSuite) SetupSuite() {
	svc, err := dbpg.New(dbpg.DbConfig{
		DbUser:     "root",
		DbPassword: "secret",
		DbHost:     "127.0.0.1",
		DbPort:     5432,
		DbName:     "tdexa-test",
		MigrationSourceURL: "file://../.." +
			"/internal/infrastructure/db/pg/migrations",
		DbInsecure: true,
	})
	if err != nil {
		s.FailNow(err.Error())
	}

	if svc != nil {
		err := svc.CreateLoader("../fixtures")
		if err != nil {
			s.FailNow(err.Error())
		}
	}

	pgDbSvc = svc
}

func (s *PgDbTestSuite) TearDownSuite() {
	if err := pgDbSvc.Close(); err != nil {
		s.FailNow(err.Error())
	}
}

func (s *PgDbTestSuite) BeforeTest(suiteName, testName string) {
	if err := pgDbSvc.LoadFixtures(); err != nil {
		s.FailNow(err.Error())
	}
}

func (s *PgDbTestSuite) AfterTest(suiteName, testName string) {

}
