package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/evmos/ethermint/x/bitcoincommiter/types"
	"github.com/spf13/cobra"
)

var (
// DefaultRelativePacketTimeoutTimestamp = (time.Duration(10) * time.Minute).Nanoseconds()
)

const (
// flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
// listSeparator              = ","
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// this line is used by starport scaffolding # 1

	return cmd
}
