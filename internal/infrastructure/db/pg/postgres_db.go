package dbpg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/tdex-network/tdex-analytics/internal/core/domain"
	"github.com/tdex-network/tdex-analytics/internal/infrastructure/db/pg/sqlc/queries"
	"sync"

	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
)

const (
	postgresDriver             = "postgres"
	insecureDataSourceTemplate = "postgresql://%s:%s@%s:%d/%s?sslmode=disable"
	dataSourceTemplate         = "host=%s port=%d user=%s password=%s dbname=%s"
	postgresDialect            = "postgres"
)

type Service interface {
	domain.MarketRepository
	Close() error
	CreateLoader(fixturesPath string) error
	LoadFixtures() error
}

type Config struct {
	Org             string
	AuthToken       string
	DbUrl           string
	AnalyticsBucket string
}

type postgresDbService struct {
	db             *sql.DB
	querier        *queries.Queries
	fixturesLoader *testfixtures.Loader
	mutex          *sync.RWMutex
}

func New(dbConfig DbConfig) (Service, error) {
	db, err := connect(dbConfig)
	if err != nil {
		return nil, err
	}

	if err = migrateDb(db, dbConfig.MigrationSourceURL); err != nil {
		return nil, err
	}

	return &postgresDbService{
		db:      db,
		querier: queries.New(db),
	}, nil
}

type DbConfig struct {
	DbUser             string
	DbPassword         string
	DbHost             string
	DbPort             int
	DbName             string
	MigrationSourceURL string
	DbInsecure         bool
	AwsRegion          string
}

func (p *postgresDbService) Close() error {
	return p.db.Close()
}

func connect(dbConfig DbConfig) (*sql.DB, error) {
	dataSource := insecureDataSourceStr(dbConfig)

	if !dbConfig.DbInsecure {
		d, err := dataSourceStr(dbConfig)
		if err != nil {
			return nil, err
		}

		dataSource = d
	}

	db, err := sql.Open(
		postgresDialect,
		dataSource,
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDb(db *sql.DB, migrationSourceUrl string) error {
	dbInstance, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrationSourceUrl,
		postgresDriver,
		dbInstance,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func insecureDataSourceStr(dbConfig DbConfig) string {
	return fmt.Sprintf(
		insecureDataSourceTemplate,
		dbConfig.DbUser,
		dbConfig.DbPassword,
		dbConfig.DbHost,
		dbConfig.DbPort,
		dbConfig.DbName,
	)
}

func dataSourceStr(dbConfig DbConfig) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", err
	}

	authenticationToken, err := auth.BuildAuthToken(
		context.TODO(),
		fmt.Sprintf("%s:%d", dbConfig.DbHost, dbConfig.DbPort),
		dbConfig.AwsRegion,
		dbConfig.DbUser,
		cfg.Credentials,
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		dataSourceTemplate,
		dbConfig.DbHost,
		dbConfig.DbPort,
		dbConfig.DbUser,
		authenticationToken,
		dbConfig.DbName,
	), nil
}

func (p *postgresDbService) CreateLoader(fixturesPath string) error {
	f, err := testfixtures.New(
		testfixtures.Database(p.db),
		testfixtures.Dialect(postgresDialect),
		testfixtures.Directory(fixturesPath),
	)
	if err != nil {
		return err
	}

	p.fixturesLoader = f
	p.mutex = new(sync.RWMutex)

	return nil
}

func (p *postgresDbService) LoadFixtures() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	err := p.fixturesLoader.Load()
	if err != nil {
		return err
	}

	return nil
}
