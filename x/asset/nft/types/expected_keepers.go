package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/nft"
)

// NFTKeeper defines the expected NFT interface.
type NFTKeeper interface {
	SaveClass(ctx sdk.Context, class nft.Class) error
	HasClass(ctx sdk.Context, classID string) bool
	GetClasses(ctx sdk.Context) (classes []*nft.Class)
	HasNFT(ctx sdk.Context, classID, id string) bool
	Mint(ctx sdk.Context, token nft.NFT, receiver sdk.AccAddress) error
}
