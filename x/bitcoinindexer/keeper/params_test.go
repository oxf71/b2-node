package keeper_test

import (
	"testing"

	testkeeper "github.com/evmos/ethermint/testutil/keeper"
	"github.com/evmos/ethermint/x/bitcoinindexer/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.BitcoinindexerKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
