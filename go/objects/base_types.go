// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"bytes"
	"encoding"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	masscbor "github.com/masslbs/network-schema/go/cbor"
)

// ObjectID is the unique identifier for an object
type ObjectID = uint64

// SignatureSize is the size of a signature
const SignatureSize = crypto.SignatureLength

// Signature represents a cryptographic signature
type Signature [SignatureSize]byte

// UnmarshalBinary unmarshals a signature from a byte slice
func (val *Signature) UnmarshalBinary(data []byte) error {
	if n := uint(len(data)); n != SignatureSize {
		return ErrBytesTooShort{Want: SignatureSize, Got: n}
	}
	copy(val[:], data)
	return nil
}

// PublicKeySize is the size of a public key
const PublicKeySize = 33

// PublicKey represents a (compressed) ecdsa public key
type PublicKey [PublicKeySize]byte

// UnmarshalBinary unmarshals a public key from a byte slice
func (val *PublicKey) UnmarshalBinary(data []byte) error {
	if n := uint(len(data)); n != PublicKeySize {
		return ErrBytesTooShort{Want: PublicKeySize, Got: n}
	}
	copy(val[:], data)
	return nil
}

// HashSize is the size of a hash
const HashSize = 32

// Hash represents a cryptographic hash
type Hash [HashSize]byte

// UnmarshalBinary unmarshals a hash from a byte slice
func (val *Hash) UnmarshalBinary(data []byte) error {
	if n := uint(len(data)); n != HashSize {
		return ErrBytesTooShort{Want: HashSize, Got: n}
	}
	copy(val[:], data)
	return nil
}

// EthereumAddressSize is the size of an Ethereum address
const EthereumAddressSize = common.AddressLength

// EthereumAddress represents an Ethereum address
type EthereumAddress struct {
	common.Address
}

// UnmarshalBinary unmarshals an Ethereum address from a byte slice
func (val *EthereumAddress) UnmarshalBinary(data []byte) error {
	if n := uint(len(data)); n != EthereumAddressSize {
		return ErrBytesTooShort{Want: EthereumAddressSize, Got: n}
	}
	copy(val.Address[:], data)
	return nil
}

// make sure these types implement encoding.BinaryUnmarshaler
var (
	_ encoding.BinaryUnmarshaler = (*Signature)(nil)
	_ encoding.BinaryUnmarshaler = (*PublicKey)(nil)
	_ encoding.BinaryUnmarshaler = (*Hash)(nil)
	_ encoding.BinaryUnmarshaler = (*EthereumAddress)(nil)
)

// Uint256 represents a 256-bit unsigned integer
type Uint256 = big.Int

// ChainAddress is an Ethereum address with a chain ID attached
type ChainAddress struct {
	ChainID uint64 `validate:"required,gt=0"`
	// when representing an ERC20 the zero address is used for the native currency
	EthereumAddress
}

// AddrFromHex creates a ChainAddress from a hex string
func AddrFromHex(chain uint64, hexAddr string) (ChainAddress, error) {
	addr := ChainAddress{ChainID: chain}
	hexAddr = strings.TrimPrefix(hexAddr, "0x")
	decoded, err := hex.DecodeString(hexAddr)
	if err != nil {
		return ChainAddress{}, err
	}
	n := copy(addr.Address[:], decoded)
	if n != EthereumAddressSize {
		return ChainAddress{}, fmt.Errorf("copy failed: %d != %d", n, EthereumAddressSize)
	}
	return addr, nil
}

// MustAddrFromHex creates a ChainAddress from a hex string, panics on error
func MustAddrFromHex(chain uint64, hexAddr string) ChainAddress {
	addr, err := AddrFromHex(chain, hexAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// Equal checks if two ChainAddresses are equal
func (ca ChainAddress) Equal(other ChainAddress) bool {
	return ca.ChainID == other.ChainID && ca.Address == other.Address
}

// String returns a string representation of a ChainAddress
func (ca *ChainAddress) String() string {
	addr := ca.Address
	return fmt.Sprintf("%s (%d)", addr.String(), ca.ChainID)
}

// Payee represents a payment recipient
type Payee struct {
	Address ChainAddress

	// controls how the payment is reaches the payee.
	// true:  forwarded via pay() method
	// false: normal transfer
	// See also:
	// https://github.com/masslbs/contracts/
	// commit: 377aba24796e029945696350db581ec1f65da657
	// file: src/IPayments.sol#L90-L95.
	CallAsContract bool
}

func (p *Payee) String() string {
	return fmt.Sprintf("%s (Contract=%v)", p.Address.String(), p.CallAsContract)
}

// ShippingRegion represents a shipping region
type ShippingRegion struct {
	Country        string
	PostalCode     string
	City           string
	PriceModifiers map[string]PriceModifier `cbor:",omitempty"`
}

// PriceModifier represents a price modification
type PriceModifier priceModifierHack

// using pointers here to express optionality clearer
// TODO: add validate:"either_or=ModificationPrecents,ModificationAbsolute"`
type priceModifierHack struct {
	// one of the following should be set
	// this is multiplied with the sub-total before being divided by 100.
	ModificationPrecents *Uint256              `cbor:",omitempty"`
	ModificationAbsolute *ModificationAbsolute `cbor:",omitempty"`
}

// UnmarshalCBOR unmarshals a PriceModifier from a byte slice
func (pm *PriceModifier) UnmarshalCBOR(data []byte) error {
	var pm2 priceModifierHack
	dec := masscbor.DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&pm2)
	if err != nil {
		return err
	}
	if pm2.ModificationPrecents == nil && pm2.ModificationAbsolute == nil {
		return fmt.Errorf("one of ModificationPrecents or ModificationAbsolute must be set")
	}
	*pm = PriceModifier(pm2)
	return nil
}

// ModificationAbsolute represents an absolute price modification
type ModificationAbsolute struct {
	Amount Uint256 `validate:"required"`
	Plus   bool    // false means subtract
}
