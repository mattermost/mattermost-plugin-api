package pluginapi_test

import (
	"database/sql"
	"testing"

	pluginapi "github.com/mattermost/mattermost-plugin-api"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"

	_ "github.com/proullon/ramsql/driver"
	"github.com/stretchr/testify/require"
)

func TestStoreSingleton(t *testing.T) {
	api1 := &plugintest.API{}
	defer api1.AssertExpectations(t)
	client1 := pluginapi.NewClient(api1)
	storePtr1 := client1.Store

	api2 := &plugintest.API{}
	defer api2.AssertExpectations(t)
	client2 := pluginapi.NewClient(api2)
	storePtr2 := client2.Store

	require.Same(t, storePtr1, storePtr2)
}
func TestStore(t *testing.T) {
	t.Run("no license", func(t *testing.T) {
		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		store := pluginapi.NewStore(api)

		api.On("GetLicense").Return(nil)
		db, err := store.GetMasterDB()
		require.Error(t, err)
		require.Nil(t, db)
	})

	t.Run("master db singleton", func(t *testing.T) {
		db, err := sql.Open("ramsql", "TestStore-master-db")
		require.NoError(t, err)
		defer db.Close()

		config := &model.Config{
			SqlSettings: model.SqlSettings{
				DriverName:                  model.NewString("ramsql"),
				DataSource:                  model.NewString("TestStore-master-db"),
				ConnMaxLifetimeMilliseconds: model.NewInt(2),
			},
		}

		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		api.On("GetLicense").Return(&model.License{})
		api.On("GetUnsanitizedConfig").Return(config)

		store := pluginapi.NewStore(api)

		db1, err := store.GetMasterDB()
		require.NoError(t, err)
		require.NotNil(t, db1)

		db2, err := store.GetMasterDB()
		require.NoError(t, err)
		require.NotNil(t, db2)

		require.Same(t, db1, db2)
	})

	t.Run("master db", func(t *testing.T) {
		db, err := sql.Open("ramsql", "TestStore-master-db")
		require.NoError(t, err)
		defer db.Close()

		_, err = db.Exec("CREATE TABLE test (id INT);")
		require.NoError(t, err)
		_, err = db.Exec("INSERT INTO test (id) VALUES (2);")
		require.NoError(t, err)

		config := &model.Config{
			SqlSettings: model.SqlSettings{
				DriverName:                  model.NewString("ramsql"),
				DataSource:                  model.NewString("TestStore-master-db"),
				ConnMaxLifetimeMilliseconds: model.NewInt(2),
			},
		}

		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		store := pluginapi.NewStore(api)
		api.On("GetLicense").Return(&model.License{})

		api.On("GetUnsanitizedConfig").Return(config)
		db, err = store.GetMasterDB()
		require.NoError(t, err)
		require.NotNil(t, db)

		var id int
		err = db.QueryRow("SELECT id FROM test").Scan(&id)
		require.NoError(t, err)
		require.Equal(t, 2, id)

		// No replica is set up
		db, err = store.GetReplicaDB()
		require.NoError(t, err)
		require.Nil(t, db)
	})

	t.Run("replica db singleton", func(t *testing.T) {
		db, err := sql.Open("ramsql", "TestStore-master-db")
		require.NoError(t, err)
		defer db.Close()

		config := &model.Config{
			SqlSettings: model.SqlSettings{
				DriverName:                  model.NewString("ramsql"),
				DataSource:                  model.NewString("TestStore-master-db"),
				DataSourceReplicas:          []string{"TestStore-master-db"},
				ConnMaxLifetimeMilliseconds: model.NewInt(2),
			},
		}

		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		api.On("GetLicense").Return(&model.License{})
		api.On("GetUnsanitizedConfig").Return(config)

		store := pluginapi.NewStore(api)

		db1, err := store.GetReplicaDB()
		require.NoError(t, err)
		require.NotNil(t, db1)

		db2, err := store.GetReplicaDB()
		require.NoError(t, err)
		require.NotNil(t, db2)

		require.Same(t, db1, db2)
	})

	t.Run("replica db", func(t *testing.T) {
		masterDB, err := sql.Open("ramsql", "TestStore-replica-db-1")
		require.NoError(t, err)
		defer masterDB.Close()

		_, err = masterDB.Exec("CREATE TABLE test (id INT);")
		require.NoError(t, err)

		replicaDB, err := sql.Open("ramsql", "TestStore-replica-db-2")
		require.NoError(t, err)
		defer masterDB.Close()

		_, err = replicaDB.Exec("CREATE TABLE test (id INT);")
		require.NoError(t, err)
		_, err = replicaDB.Exec("INSERT INTO test (id) VALUES (3);")
		require.NoError(t, err)

		config := &model.Config{
			SqlSettings: model.SqlSettings{
				DriverName:                  model.NewString("ramsql"),
				DataSource:                  model.NewString("TestStore-replica-db-1"),
				DataSourceReplicas:          []string{"TestStore-replica-db-2"},
				ConnMaxLifetimeMilliseconds: model.NewInt(2),
			},
		}

		api := &plugintest.API{}
		defer api.AssertExpectations(t)
		store := pluginapi.NewStore(api)
		api.On("GetLicense").Return(&model.License{})

		api.On("GetUnsanitizedConfig").Return(config)
		storeMasterDB, err := store.GetMasterDB()
		require.NoError(t, err)
		require.NotNil(t, storeMasterDB)

		var count int
		err = storeMasterDB.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
		require.NoError(t, err)
		require.Equal(t, 0, count)

		storeReplicaDB, err := store.GetReplicaDB()
		require.NoError(t, err)
		require.NotNil(t, storeReplicaDB)

		var id int
		err = storeReplicaDB.QueryRow("SELECT id FROM test").Scan(&id)
		require.NoError(t, err)
		require.Equal(t, 3, id)
	})
}
