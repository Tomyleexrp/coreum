package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// GetQueryCmd returns the parent command for all x/feemodel CLI query commands. The
// provided clientCtx should have, at a minimum, a verifier, Tendermint RPC client,
// and marshaler set.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the feemodel module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetMinGasPriceCmd(),
	)

	return cmd
}

// GetMinGasPriceCmd returns command for getting minimum gas price required by the network.
func GetMinGasPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "min-gas-price",
		Short: "Query for minimum gas price for current block required by the network",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			ctx := cmd.Context()
			res, err := queryClient.MinGasPrice(ctx, &types.QueryMinGasPriceRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.MinGasPrice)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
