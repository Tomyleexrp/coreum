package nft

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers the interfaces in the registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSend{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc) //nolint:nosnakecase // generated code
}
