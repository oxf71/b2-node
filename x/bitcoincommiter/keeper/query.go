package keeper

import (
	"github.com/evmos/ethermint/x/bitcoincommiter/types"
)

var _ types.QueryServer = Keeper{}
