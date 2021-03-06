package main

import (
	"github.com/flynn/flynn/pkg/cluster"
	"github.com/flynn/flynn/pkg/postgres"
	"github.com/flynn/flynn/pkg/random"

	. "github.com/flynn/flynn/Godeps/_workspace/src/github.com/flynn/go-check"
)

type MigrateSuite struct{}

var _ = Suite(&MigrateSuite{})

type testMigrator struct {
	c  *C
	db *postgres.DB
	id int
}

func (t *testMigrator) migrateTo(id int) {
	t.c.Assert((*migrations)[t.id:id].Migrate(t.db), IsNil)
	t.id = id
}

// TestMigrateJobStates checks that migrating to ID 9 does not break existing
// job records
func (MigrateSuite) TestMigrateJobStates(c *C) {
	db := setupTestDB(c, "controllertest_migrate_job_states")
	m := &testMigrator{c: c, db: db}

	// start from ID 7
	m.migrateTo(7)

	// insert a job
	hostID := "host1"
	uuid := random.UUID()
	jobID := cluster.GenerateJobID(hostID, uuid)
	appID := random.UUID()
	releaseID := random.UUID()
	c.Assert(db.Exec(`INSERT INTO apps (app_id, name) VALUES ($1, $2)`, appID, "migrate-app"), IsNil)
	c.Assert(db.Exec(`INSERT INTO releases (release_id) VALUES ($1)`, releaseID), IsNil)
	c.Assert(db.Exec(`INSERT INTO job_cache (job_id, app_id, release_id, state) VALUES ($1, $2, $3, $4)`, jobID, appID, releaseID, "up"), IsNil)

	// migrate to 8 and check job states are still constrained
	m.migrateTo(8)
	err := db.Exec(`UPDATE job_cache SET state = 'foo' WHERE job_id = $1`, jobID)
	c.Assert(err, NotNil)
	if !postgres.IsPostgresCode(err, postgres.ForeignKeyViolation) {
		c.Fatalf("expected postgres foreign key violation, got %s", err)
	}

	// migrate to 9 and check job IDs are correct, pending state is valid
	m.migrateTo(9)
	var clusterID, dbUUID, dbHostID string
	c.Assert(db.QueryRow("SELECT cluster_id, job_id, host_id FROM job_cache WHERE cluster_id = $1", jobID).Scan(&clusterID, &dbUUID, &dbHostID), IsNil)
	c.Assert(clusterID, Equals, jobID)
	c.Assert(dbUUID, Equals, uuid)
	c.Assert(dbHostID, Equals, hostID)
	c.Assert(db.Exec(`UPDATE job_cache SET state = 'pending' WHERE job_id = $1`, uuid), IsNil)
}

func (MigrateSuite) TestMigrateCriticalApps(c *C) {
	db := setupTestDB(c, "controllertest_migrate_critical_apps")
	m := &testMigrator{c: c, db: db}

	// start from ID 12
	m.migrateTo(12)

	// create the critical apps with system app meta
	criticalApps := []string{"discoverd", "flannel", "postgres", "controller"}
	meta := map[string]string{"flynn-system-app": "true"}
	for _, name := range criticalApps {
		c.Assert(db.Exec(`INSERT INTO apps (app_id, name, meta) VALUES ($1, $2, $3)`, random.UUID(), name, meta), IsNil)
	}

	// migrate to 13 and check critical app meta was updated
	m.migrateTo(13)
	for _, name := range criticalApps {
		var meta map[string]string
		c.Assert(db.QueryRow("SELECT meta FROM apps WHERE name = $1", name).Scan(&meta), IsNil)
		c.Assert(meta["flynn-system-app"], Equals, "true")
		c.Assert(meta["flynn-system-critical"], Equals, "true")
	}
}
