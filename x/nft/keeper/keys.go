package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/CoreumFoundation/coreum/pkg/store"
	"github.com/CoreumFoundation/coreum/x/nft"
)

var (
	// ClassKey is store prefix of the Class
	ClassKey = []byte{0x01}
	// NFTKey is store prefix of the NFT
	NFTKey = []byte{0x02}
	// NFTOfClassByOwnerKey is store prefix of the NFTOfClassByOwner
	NFTOfClassByOwnerKey = []byte{0x03}
	// OwnerKey is store prefix of the Owner
	OwnerKey = []byte{0x04}
	// ClassTotalSupply is store prefix of the ClassTotalSupply
	ClassTotalSupply = []byte{0x05}

	// Delimiter is store key Delimiter
	Delimiter = []byte{0x00}
	// Placeholder is store key Placeholder
	Placeholder = []byte{0x01}
)

// StoreKey is the store key string for nft
const StoreKey = nft.ModuleName

// classStoreKey returns the byte representation of the nft class key
func classStoreKey(classID string) []byte {
	key := make([]byte, len(ClassKey)+len(classID))
	copy(key, ClassKey)
	copy(key[len(ClassKey):], classID)
	return key
}

// nftStoreKey returns the byte representation of the nft
func nftStoreKey(classID string) []byte {
	key := make([]byte, len(NFTKey)+len(classID)+len(Delimiter))
	copy(key, NFTKey)
	copy(key[len(NFTKey):], classID)
	copy(key[len(NFTKey)+len(classID):], Delimiter)
	return key
}

// classTotalSupply returns the byte representation of the ClassTotalSupply
func classTotalSupply(classID string) []byte {
	key := make([]byte, len(ClassTotalSupply)+len(classID))
	copy(key, ClassTotalSupply)
	copy(key[len(ClassTotalSupply):], classID)
	return key
}

// nftOfClassByOwnerStoreKey returns the byte representation of the nft owner
// Items are stored with the following key: values
// 0x03<owner><Delimiter(1 Byte)><classID><Delimiter(1 Byte)>
func nftOfClassByOwnerStoreKey(owner sdk.AccAddress, classID string) []byte {
	owner = address.MustLengthPrefix(owner)
	classIDBz := store.UnsafeStrToBytes(classID)

	key := make([]byte, len(NFTOfClassByOwnerKey)+len(owner)+len(Delimiter)+len(classIDBz)+len(Delimiter))
	copy(key, NFTOfClassByOwnerKey)
	copy(key[len(NFTOfClassByOwnerKey):], owner)
	copy(key[len(NFTOfClassByOwnerKey)+len(owner):], Delimiter)
	copy(key[len(NFTOfClassByOwnerKey)+len(owner)+len(Delimiter):], classIDBz)
	copy(key[len(NFTOfClassByOwnerKey)+len(owner)+len(Delimiter)+len(classIDBz):], Delimiter)
	return key
}

// prefixNftOfClassByOwnerStoreKey returns the prefix of the result of the method nftOfClassByOwnerStoreKey
// Items are stored with the following key: values
// 0x03<owner><Delimiter>
func prefixNftOfClassByOwnerStoreKey(owner sdk.AccAddress) []byte {
	owner = address.MustLengthPrefix(owner)

	key := make([]byte, len(NFTOfClassByOwnerKey)+len(owner)+len(Delimiter))
	copy(key, NFTOfClassByOwnerKey)
	copy(key[len(NFTOfClassByOwnerKey):], owner)
	copy(key[len(NFTOfClassByOwnerKey)+len(owner):], Delimiter)
	return key
}

// Note: the full path of the nftOfClassByOwnerStoreKey stored in the store: 0x03<owner><Delimiter><classID><Delimiter><nftID>,
// the key of the prefix store query result constructed using the prefixNftOfClassByOwnerStoreKey function needs to remove the 0x03<owner><Delimiter> prefix
func parseNftOfClassByOwnerStoreKey(key []byte) (classID, nftID string) {
	ret := bytes.Split(key, Delimiter)
	if len(ret) != 2 {
		panic("invalid nftOfClassByOwnerStoreKey")
	}
	classID = store.UnsafeBytesToStr(ret[0])
	nftID = string(ret[1])
	return classID, nftID
}

// ownerStoreKey returns the byte representation of the nft owner
// Items are stored with the following key: values
// 0x04<classID><Delimiter(1 Byte)><nftID>
func ownerStoreKey(classID, nftID string) []byte {
	// key is of format:
	classIDBz := store.UnsafeStrToBytes(classID)
	nftIDBz := store.UnsafeStrToBytes(nftID)

	key := make([]byte, len(OwnerKey)+len(classIDBz)+len(Delimiter)+len(nftIDBz))
	copy(key, OwnerKey)
	copy(key[len(OwnerKey):], classIDBz)
	copy(key[len(OwnerKey)+len(classIDBz):], Delimiter)
	copy(key[len(OwnerKey)+len(classIDBz)+len(Delimiter):], nftIDBz)
	return key
}
