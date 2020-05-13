package pluginapi

import (
	"database/sql"
	"math/rand"
	"sync"
	"time"

	// import sql drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

// StoreService exposes the underlying database.
type StoreService struct {
	initialized bool
	api         plugin.API
	mutex       sync.Mutex

	masterDB  *sql.DB
	replicaDB *sql.DB
}

func NewStore(api plugin.API) *StoreService {
	return &StoreService{api: api}
}

// Gets the master database handle.
//
// Minimum server version: 5.16
func (s *StoreService) GetMasterDB() (*sql.DB, error) {
	if err := s.initialize(); err != nil {
		return nil, err
	}

	return s.masterDB, nil
}

// Gets the replica database handle.
// Returns masterDB if a replica is not configured.
//
// Minimum server version: 5.16
func (s *StoreService) GetReplicaDB() (*sql.DB, error) {
	if err := s.initialize(); err != nil {
		return nil, err
	}

	if s.replicaDB != nil {
		return s.replicaDB, nil
	}

	return s.masterDB, nil
}

// Close closes any open resources. This method is idempotent.
func (s *StoreService) Close() error {
	if !s.initialized {
		return nil
	}

	if err := s.masterDB.Close(); err != nil {
		return err
	}

	if s.replicaDB != nil {
		if err := s.replicaDB.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (s *StoreService) DriverName() string {
	return *s.api.GetConfig().SqlSettings.DriverName
}

func (s *StoreService) initialize() error {
	// This check is not atomic; not worth the extra complexity here
	// considering this method is not called very often.
	if s.initialized {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.initialized {
		if s.api.GetLicense() == nil {
			return errors.New("this feature requires a valid enterprise license")
		}

		config := s.api.GetUnsanitizedConfig()

		// Set up master db
		db, err := setupConnection(*config.SqlSettings.DataSource, config.SqlSettings)
		if err != nil {
			return errors.Wrap(err, "failed to connect to master db")
		}
		s.masterDB = db

		// Set up replica db
		if len(config.SqlSettings.DataSourceReplicas) > 0 {
			replicaSource := config.SqlSettings.DataSourceReplicas[rand.Intn(len(config.SqlSettings.DataSourceReplicas))]

			db, err := setupConnection(replicaSource, config.SqlSettings)
			if err != nil {
				return errors.Wrap(err, "failed to connect to replica db")
			}
			s.replicaDB = db
		}

		s.initialized = true
	}

	return nil
}

func setupConnection(dataSourceName string, settings model.SqlSettings) (*sql.DB, error) {
	driverName := *settings.DriverName
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open SQL connection")
	}

	// Set at most 2 connections for plugins
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Duration(*settings.ConnMaxLifetimeMilliseconds) * time.Millisecond)

	return db, nil
}
