package baseapp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/db/memdb"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	"github.com/cosmos/cosmos-sdk/snapshots"
	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGetBlockRentionHeight(t *testing.T) {
	logger := defaultLogger()
	name := t.Name()
	snapshotStore, err := snapshots.NewStore(memdb.NewDB(), testutil.GetTempDir(t))
	require.NoError(t, err)
	testCases := map[string]struct {
		bapp         *BaseApp
		maxAgeBlocks int64
		commitHeight int64
		expected     int64
	}{
		"defaults": {
			bapp:         NewBaseApp(name, logger, memdb.NewDB(), nil),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     0,
		},
		"pruning unbonding time only": {
			bapp:         NewBaseApp(name, logger, memdb.NewDB(), nil, SetMinRetainBlocks(1)),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     136120,
		},
		"pruning iavl snapshot only": {
			bapp: NewBaseApp(
				name, logger, memdb.NewDB(), nil,
				SetPruning(pruningtypes.PruneNothing),
				SetMinRetainBlocks(1),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     490000,
		},
		"pruning state sync snapshot only": {
			bapp: NewBaseApp(
				name, logger, memdb.NewDB(), nil,
				SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
				SetMinRetainBlocks(1),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     349000,
		},
		"pruning min retention only": {
			bapp: NewBaseApp(
				name, logger, memdb.NewDB(), nil,
				SetMinRetainBlocks(400000),
			),
			maxAgeBlocks: 0,
			commitHeight: 499000,
			expected:     99000,
		},
		"pruning all conditions": {
			bapp: NewBaseApp(
				name, logger, memdb.NewDB(), nil,
				SetPruning(pruningtypes.NewPruningOptions(0, 0, 0)),
				SetMinRetainBlocks(400000),
				SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     99000,
		},
		"no pruning due to no persisted state": {
			bapp: NewBaseApp(
				name, logger, memdb.NewDB(), nil,
				SetPruning(pruningtypes.NewPruningOptions(0, 0, 0)),
				SetMinRetainBlocks(400000),
			),
			maxAgeBlocks: 362880,
			commitHeight: 10000,
			expected:     0,
		},
		"disable pruning": {
			bapp: NewBaseApp(
				name, logger, memdb.NewDB(), nil,
				SetPruning(pruningtypes.NewPruningOptions(0, 0, 0)),
				SetMinRetainBlocks(0),
				SetSnapshot(snapshotStore, snapshottypes.NewSnapshotOptions(50000, 3)),
			),
			maxAgeBlocks: 362880,
			commitHeight: 499000,
			expected:     0,
		},
	}

	for name, tc := range testCases {
		tc := tc
		db := memdb.NewDB()
		tc.bapp.SetParamStore(&paramStore{db: db})
		tc.bapp.InitChain(abci.RequestInitChain{
			ConsensusParams: &abci.ConsensusParams{
				Evidence: &tmprototypes.EvidenceParams{
					MaxAgeNumBlocks: tc.maxAgeBlocks,
				},
			},
		})

		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.bapp.GetBlockRetentionHeight(tc.commitHeight))
		})
	}
}

// Test and ensure that negative heights always cause errors.
// See issue https://github.com/cosmos/cosmos-sdk/issues/7662.
func TestBaseAppCreateQueryContextRejectsNegativeHeights(t *testing.T) {
	t.Parallel()

	logger := defaultLogger()
	db := memdb.NewDB()
	name := t.Name()
	app := NewBaseApp(name, logger, db, nil)

	proves := []bool{
		false, true,
	}
	for _, prove := range proves {
		t.Run(fmt.Sprintf("prove=%t", prove), func(t *testing.T) {
			sctx, err := app.createQueryContext(-10, true)
			require.Error(t, err)
			require.Equal(t, sctx, sdk.Context{})
		})
	}
}
