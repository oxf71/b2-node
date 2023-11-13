package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/evmos/ethermint/testutil/keeper"
	"github.com/evmos/ethermint/x/bitcoincommiter/keeper"
	"github.com/evmos/ethermint/x/bitcoincommiter/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.BitcoincommiterKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
