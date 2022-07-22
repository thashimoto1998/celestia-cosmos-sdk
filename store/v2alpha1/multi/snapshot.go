package multi

import (
	"bytes"
	"fmt"
	"io"
	"sort"

	snapshottypes "github.com/cosmos/cosmos-sdk/snapshots/types"

	"github.com/cosmos/cosmos-sdk/snapshots"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	types "github.com/cosmos/cosmos-sdk/store/v2alpha1"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Snapshot implements snapshottypes.Snapshotter.
func (rs *Store) Snapshot(height uint64, format uint32) (<-chan io.ReadCloser, error) {
	if format != snapshottypes.CurrentFormat {
		return nil, sdkerrors.Wrapf(snapshottypes.ErrUnknownFormat, "format %v", format)
	}

	if height == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrLogic, "cannot snapshot height 0")
	}
	if height > uint64(rs.LastCommitID().Version) {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot snapshot future height %v", height)
	}

	// get the saved snapshot at height
	vs, err := rs.getView(int64(height))
	if err != nil {
		return nil, sdkerrors.Wrap(err, fmt.Sprintf("error while get the version at height %d", height))
	}

	// sending the snapshot store schema
	var storeByteKeys [][]byte
	for sKey := range vs.schema {
		if vs.schema[sKey] == storetypes.StoreTypePersistent {
			storeByteKeys = append(storeByteKeys, []byte(sKey))
		}
	}

	sort.Slice(storeByteKeys, func(i, j int) bool {
		return bytes.Compare(storeByteKeys[i], storeByteKeys[j]) == -1
	})

	chunks := make(chan io.ReadCloser)
	protoWriter := snapshots.NewStreamWriter(chunks)

	err = protoWriter.WriteMsg(&storetypes.SnapshotItem{
		Item: &storetypes.SnapshotItem_Schema{
			Schema: &storetypes.SnapshotSchema{
				Keys: storeByteKeys,
			},
		},
	})
	if err != nil {
		return chunks, err
	}

	for _, sKey := range storeByteKeys {
		subStore, err := vs.getSubstore(string(sKey))
		if err != nil {
			return chunks, err
		}

		err = protoWriter.WriteMsg(&storetypes.SnapshotItem{
			Item: &storetypes.SnapshotItem_Store{
				Store: &storetypes.SnapshotStoreItem{
					Name: string(sKey),
				},
			},
		})
		if err != nil {
			return chunks, err
		}

		iter := subStore.Iterator(nil, nil)
		for ; iter.Valid(); iter.Next() {
			err = protoWriter.WriteMsg(&storetypes.SnapshotItem{
				Item: &storetypes.SnapshotItem_KV{
					KV: &storetypes.SnapshotKVItem{
						Key:   iter.Key(),
						Value: iter.Value(),
					},
				},
			})
			if err != nil {
				return chunks, err
			}
		}

		err = iter.Close()
		if err != nil {
			return chunks, err
		}
	}

	return chunks, nil
}

// Restore implements snapshottypes.Snapshotter.
func (rs *Store) Restore(
	height uint64, format uint32, chunks <-chan io.ReadCloser, ready chan<- struct{},
) error {

	if err := snapshots.ValidRestoreHeight(format, height); err != nil {
		return err
	}

	if rs.LastCommitID().Version != 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "cannot restore snapshot for non empty store at height %v", height)
	}

	var subStore *substore
	var storeSchemaReceived = false
	var receivedStoreSchema StoreSchema

	var snapshotItem storetypes.SnapshotItem

	protoReader, err := snapshots.NewStreamReader(chunks)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrLogic, "invalid chunks, could not initiliaze stream reader")
	}

loop:
	for {
		snapshotItem = storetypes.SnapshotItem{}
		err := protoReader.ReadMsg(&snapshotItem)
		if err == io.EOF {
			break
		} else if err != nil {
			return sdkerrors.Wrap(err, "invalid protobuf message")
		}

		switch item := snapshotItem.Item.(type) {
		case *storetypes.SnapshotItem_Schema:
			receivedStoreSchema = make(StoreSchema, len(item.Schema.GetKeys()))
			storeSchemaReceived = true
			for _, sKey := range item.Schema.GetKeys() {
				receivedStoreSchema[string(sKey)] = types.StoreTypePersistent
			}

			if !receivedStoreSchema.matches(rs.schema) {
				return sdkerrors.Wrap(sdkerrors.ErrLogic, "received schema does not match app schema")
			}

		case *storetypes.SnapshotItem_Store:
			storeName := item.Store.GetName()
			// checking the store schema is received or not
			if !storeSchemaReceived {
				return sdkerrors.Wrapf(sdkerrors.ErrLogic, "received store name before store schema %s", storeName)
			}
			// checking the store schema exists or not
			if _, has := receivedStoreSchema[storeName]; !has {
				return sdkerrors.Wrapf(sdkerrors.ErrLogic, "store is missing from schema %s", storeName)
			}

			// get the substore
			subStore, err = rs.getSubstore(storeName)
			if err != nil {
				return sdkerrors.Wrap(err, fmt.Sprintf("error while getting the substore for key %s", storeName))
			}

		case *storetypes.SnapshotItem_KV:
			if subStore == nil {
				return sdkerrors.Wrap(sdkerrors.ErrLogic, "received KV Item before store item")
			}
			// update the key/value SMT.Store
			subStore.Set(item.KV.Key, item.KV.Value)

		default:
			break loop
		}
	}

	// commit all key/values to store
	_, err = rs.commit(height)
	if err != nil {
		return sdkerrors.Wrap(err, fmt.Sprintf("error during commit the store at height %d", height))
	}

	return nil
}
