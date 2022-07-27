package multi

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/types"
)

const (
	latestVersionKey = "s/latest"
	pruneHeightsKey  = "s/pruneheights"
	commitInfoKeyFmt = "s/%d" // s/<version>
)

// LoadLatestVersion implements CommitMultiStore.
func LoadLatestVersion(db dbm.DB) error {
	ver := getLatestVersion(db)
	return loadVersion(db, ver)
}

func getLatestVersion(db dbm.DB) uint64 {
	bz, err := db.Get([]byte(latestVersionKey))
	if err != nil {
		panic(err)
	} else if bz == nil {
		return 0
	}

	var latestVersion uint64

	if err := gogotypes.StdUInt64Unmarshal(&latestVersion, bz); err != nil {
		panic(err)
	}

	return latestVersion
}

func loadVersion(db dbm.DB, ver uint64) error {
	infos := make(map[string]types.StoreInfo)

	cInfo := &types.CommitInfo{}

	// load old data if we are not version 0
	if ver != 0 {
		var err error
		cInfo, err = getCommitInfo(db, ver)
		if err != nil {
			return err
		}

		// convert StoreInfos slice to map
		for _, storeInfo := range cInfo.StoreInfos {
			infos[storeInfo.Name] = storeInfo
		}
	}

	// load each Store (note this doesn't panic on unmounted keys now)
	var newStores = make(map[types.StoreKey]types.CommitKVStore)

	storesKeys := make([]types.StoreKey, 0, len(rs.storesParams))

	for key := range rs.storesParams {
		storesKeys = append(storesKeys, key)
	}

	for _, key := range storesKeys {
		storeParams := rs.storesParams[key]
		commitID := rs.getCommitID(infos, key.Name())

		// If it has been added, set the initial version
		if upgrades.IsAdded(key.Name()) {
			storeParams.initialVersion = uint64(ver) + 1
		}

		store, err := rs.loadCommitStoreFromParams(key, commitID, storeParams)
		if err != nil {
			return errors.Wrap(err, "failed to load store")
		}

		newStores[key] = store

		// If it was deleted, remove all data
		if upgrades.IsDeleted(key.Name()) {
			if err := deleteKVStore(store.(types.KVStore)); err != nil {
				return errors.Wrapf(err, "failed to delete store %s", key.Name())
			}
		} else if oldName := upgrades.RenamedFrom(key.Name()); oldName != "" {
			// handle renames specially
			// make an unregistered key to satify loadCommitStore params
			oldKey := types.NewKVStoreKey(oldName)
			oldParams := storeParams
			oldParams.key = oldKey

			// load from the old name
			oldStore, err := rs.loadCommitStoreFromParams(oldKey, rs.getCommitID(infos, oldName), oldParams)
			if err != nil {
				return errors.Wrapf(err, "failed to load old store %s", oldName)
			}

			// move all data
			if err := moveKVStoreData(oldStore.(types.KVStore), store.(types.KVStore)); err != nil {
				return errors.Wrapf(err, "failed to move store %s -> %s", oldName, key.Name())
			}
		}
	}

	rs.lastCommitInfo = cInfo
	rs.stores = newStores

	// load any pruned heights we missed from disk to be pruned on the next run
	ph, err := getPruningHeights(rs.db)
	if err == nil && len(ph) > 0 {
		rs.pruneHeights = ph
	}

	return nil
}

// Gets commitInfo from disk.
func getCommitInfo(db dbm.DB, ver int64) (*types.CommitInfo, error) {
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)

	bz, err := db.Get([]byte(cInfoKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get commit info")
	} else if bz == nil {
		return nil, errors.New("no commit info found")
	}

	cInfo := &types.CommitInfo{}
	if err = cInfo.Unmarshal(bz); err != nil {
		return nil, errors.Wrap(err, "failed unmarshal commit info")
	}

	return cInfo, nil
}
