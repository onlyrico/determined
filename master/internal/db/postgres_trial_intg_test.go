//go:build integration
// +build integration

package db

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/determined-ai/determined/master/pkg/etc"
	"github.com/determined-ai/determined/master/pkg/ptrs"
	"github.com/determined-ai/determined/proto/pkg/trialv1"

	"github.com/stretchr/testify/require"

	"github.com/determined-ai/determined/master/pkg/model"
)

func TestProtoGetTrial(t *testing.T) {
	etc.SetRootPath(rootFromDB)
	db := MustResolveTestPostgres(t)
	MustMigrateTestPostgres(t, db, migrationsFromDB)

	exp := model.ExperimentModel()
	err := db.AddExperiment(exp)
	require.NoError(t, err, "failed to add experiment")

	tr := model.TrialModel(exp.ID, exp.JobID)
	err = db.AddTrial(tr)
	require.NoError(t, err, "failed to add trial")

	startTime := time.Now().UTC()
	for i := 0; i < 3; i++ {
		a := &model.Allocation{
			AllocationID: model.NewAllocationID(fmt.Sprintf("%s-%d", tr.TaskID, i)),
			TaskID:       tr.TaskID,
			StartTime:    ptrs.Ptr(startTime.Add(time.Duration(i) * time.Second)),
			EndTime:      ptrs.Ptr(startTime.Add(time.Duration(i+1) * time.Second)),
		}
		err = db.AddAllocation(a)
		require.NoError(t, err, "failed to add allocation")
		err = db.CompleteAllocation(a)
		require.NoError(t, err, "failed to complete allocation")
	}

	var trResp trialv1.Trial
	err = db.QueryProto(
		"proto_get_trials_plus",
		&trResp,
		"{"+strconv.Itoa(int(tr.ID))+"}",
	)
	require.NoError(t, err, "failed to query trial")
	require.Equal(t, trResp.WallClockTime, float64(3), "wall clock time is wrong")
}
