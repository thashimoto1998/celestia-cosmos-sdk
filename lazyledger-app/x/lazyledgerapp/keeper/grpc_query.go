package keeper

import (
	"github.com/cosmos/cosmos-sdk/lazyledger-app/x/lazyledgerapp/types"
)

var _ types.QueryServer = Keeper{}
