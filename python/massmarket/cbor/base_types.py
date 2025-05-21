# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from typing import Dict, List, Optional, Any, Union
import os
import re
from decimal import Decimal

import cbor2


class Uint256:
    """Represents a 256-bit unsigned integer, similar to Go's big.Int"""

    @classmethod
    def random(cls) -> "Uint256":
        return cls(int.from_bytes(os.urandom(32), byteorder="big"))

    def __init__(self, value: Union[int, str, "Uint256", Decimal]):
        if isinstance(value, Uint256):
            self._value = value._value
        elif isinstance(value, str):
            # Handle hex strings
            if value.startswith("0x"):
                self._value = int(value, 16)
            else:
                self._value = int(value)
        elif isinstance(value, (int, Decimal)):
            if value < 0:
                raise ValueError("Uint256 cannot be negative")
            self._value = int(value)
        else:
            raise TypeError(f"Cannot create Uint256 from {type(value)}")

        # Ensure value fits in 256 bits
        if self._value >= (1 << 256):
            raise ValueError("Value exceeds 256 bits")

    def __str__(self) -> str:
        return str(self._value)

    def __repr__(self) -> str:
        return f"Uint256({self._value})"

    def __int__(self) -> int:
        return self._value

    def __eq__(self, other) -> bool:
        if isinstance(other, Uint256):
            return self._value == other._value
        elif isinstance(other, int):
            return self._value == other
        elif isinstance(other, str):
            return self._value == int(other)
        else:
            raise ValueError(f"Cannot compare Uint256 with {type(other)}")

    def to_hex(self) -> str:
        """Return hex representation with 0x prefix"""
        return hex(self._value)

    def to_bytes(self, length: int = 32) -> bytes:
        """Convert to big-endian bytes representation"""
        return self._value.to_bytes(length, byteorder="big")

    def to_cbor_dict(self) -> dict:
        return int(self._value)

    @classmethod
    def from_bytes(cls, data: bytes) -> "Uint256":
        """Create from big-endian bytes representation"""
        return cls(int.from_bytes(data, byteorder="big"))

    @classmethod
    def from_cbor(cls, data) -> "Uint256":
        """Create from CBOR-encoded data"""
        if isinstance(data, bytes):
            return cls.from_bytes(data)
        elif isinstance(data, int):
            return cls(data)
        else:
            raise TypeError(f"Cannot create Uint256 from {type(data)}")


class EthereumAddress:
    """Represents an Ethereum address"""

    SIZE = 20  # common.AddressLength in Go

    def __init__(self, value: Union[str, bytes, "EthereumAddress"]):
        if isinstance(value, EthereumAddress):
            self._bytes = value._bytes
        elif isinstance(value, str):
            # Strip 0x prefix if present
            hex_str = value.lower().replace("0x", "")
            if not re.match(r"^[0-9a-f]{40}$", hex_str):
                raise ValueError("Invalid Ethereum address format")
            self._bytes = bytes.fromhex(hex_str)
        elif isinstance(value, bytes):
            if len(value) != self.SIZE:
                raise ValueError(f"Ethereum address must be {self.SIZE} bytes")
            self._bytes = value
        else:
            raise TypeError(f"Cannot create EthereumAddress from {type(value)}")

    def __str__(self) -> str:
        return f"0x{self._bytes.hex()}"

    def __repr__(self) -> str:
        return f"EthereumAddress('{str(self)}')"

    def __eq__(self, other) -> bool:
        if isinstance(other, bytes):
            return self._bytes == other
        elif not isinstance(other, EthereumAddress):
            raise ValueError("Cannot compare EthereumAddress with non-EthereumAddress")
        return self._bytes == other._bytes

    def __hash__(self) -> int:
        return hash(self._bytes)

    def __bytes__(self) -> bytes:
        return self._bytes

    def to_bytes(self) -> bytes:
        return self._bytes

    @classmethod
    def from_bytes(cls, data: bytes) -> "EthereumAddress":
        return cls(data)

    def __cbor_encode__(self) -> bytes:
        """Custom CBOR encoding to match Go's implementation"""
        return self._bytes

    @classmethod
    def from_cbor(cls, data: bytes) -> "EthereumAddress":
        """Create from CBOR-encoded data"""
        return cls(data)


@dataclass
class ChainAddress:
    """Represents an Ethereum address with an associated chain ID"""

    chain_id: int
    address: EthereumAddress

    def __init__(self, chain_id: int, address: Union[str, bytes, EthereumAddress]):
        if chain_id <= 0:
            raise ValueError("ChainID must be greater than 0")
        self.chain_id = chain_id
        self.address = (
            address
            if isinstance(address, EthereumAddress)
            else EthereumAddress(address)
        )

    def __str__(self) -> str:
        return f"ChainAddress(chain_id={self.chain_id}, address={str(self.address)})"

    def __eq__(self, other) -> bool:
        if not isinstance(other, ChainAddress):
            return False
        return self.chain_id == other.chain_id and self.address == other.address

    @classmethod
    def from_cbor_dict(cls, d: Dict[str, Any]) -> "ChainAddress":
        """Create from a CBOR-decoded dictionary"""
        return cls(
            chain_id=d["ChainID"],
            address=(
                EthereumAddress.from_bytes(d["Address"])
                if isinstance(d["Address"], bytes)
                else EthereumAddress(d["Address"])
            ),
        )

    def to_cbor_dict(self) -> Dict[str, Any]:
        """Convert to a dictionary suitable for CBOR encoding"""
        return {"ChainID": self.chain_id, "Address": bytes(self.address)}


@dataclass
class PublicKey:
    key: bytes

    def to_cbor_dict(self) -> dict:
        return self.key

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "PublicKey":
        return cls(d)

    def __hash__(self):
        return hash(self.key)

    def __post_init__(self):
        if len(self.key) != 33:
            raise ValueError("PublicKey must be 33 bytes but got %d" % len(self.key))


@dataclass
class Account:
    keycards: List[PublicKey]
    guest: bool

    def to_cbor_dict(self) -> dict:
        return {
            "KeyCards": [k.to_cbor_dict() for k in self.keycards],
            "Guest": self.guest,
        }

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Account":
        assert isinstance(d["Guest"], bool)
        if d["KeyCards"] is None:
            d["KeyCards"] = []
        return cls(
            keycards=[PublicKey(k) for k in d["KeyCards"]],
            guest=d["Guest"],
        )

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "Account":
        d = cbor2.loads(cbor_data)
        return cls.from_cbor_dict(d)


# manifest and friends
@dataclass
class ModificationAbsolute:
    amount: Uint256
    plus: bool

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "ModificationAbsolute":
        return cls(
            amount=Uint256(d["Amount"]),
            plus=d["Plus"],
        )

    def to_cbor_dict(self) -> dict:
        return {
            "Amount": self.amount.to_cbor_dict(),
            "Plus": self.plus,
        }


@dataclass
class PriceModifier:
    modification_percent: Optional[Uint256] = None
    modification_absolute: Optional[ModificationAbsolute] = None

    def __post_init__(self):
        if self.modification_percent is None and self.modification_absolute is None:
            raise ValueError(
                "One of modification_percent or modification_absolute must be set"
            )
        if (
            self.modification_percent is not None
            and self.modification_absolute is not None
        ):
            raise ValueError(
                "Only one of modification_percent or modification_absolute can be set"
            )

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "PriceModifier":
        mp = d.get("ModificationPercent")
        ma = d.get("ModificationAbsolute")
        if ma is not None and isinstance(ma, dict):
            ma = ModificationAbsolute.from_cbor_dict(ma)
        return cls(
            modification_percent=mp,
            modification_absolute=ma,
        )

    def to_cbor_dict(self) -> dict:
        d = {}
        if self.modification_percent is not None:
            d["ModificationPercent"] = self.modification_percent
        if self.modification_absolute is not None:
            d["ModificationAbsolute"] = self.modification_absolute.to_cbor_dict()
        return d


@dataclass
class ShippingRegion:
    country: str
    postal_code: str
    city: str
    price_modifiers: Optional[Dict[str, PriceModifier]] = None

    def __post_init__(self):
        if self.price_modifiers is not None and not self.price_modifiers:
            raise ValueError("PriceModifiers map cannot be empty if present")

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "ShippingRegion":
        pm = d.get("PriceModifiers")
        if pm is not None:
            pm = {k: PriceModifier.from_cbor_dict(v) for k, v in pm.items()}
        return cls(
            country=d["Country"],
            postal_code=d["PostalCode"],
            city=d["City"],
            price_modifiers=pm,
        )

    def to_cbor_dict(self) -> dict:
        d = {
            "Country": self.country,
            "PostalCode": self.postal_code,
            "City": self.city,
        }
        if self.price_modifiers is not None:
            d["PriceModifiers"] = {
                k: v.to_cbor_dict() for k, v in self.price_modifiers.items()
            }
        return d


@dataclass
class Tag:
    name: str
    listings: List[int]

    def to_cbor_dict(self) -> dict:
        return {
            "Name": self.name,
            "Listings": self.listings,
        }

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Tag":
        return cls(
            name=d["Name"],
            listings=d["Listings"],
        )


@dataclass
class Payee:
    address: ChainAddress
    call_as_contract: bool

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Payee":
        return cls(
            address=ChainAddress.from_cbor_dict(d["Address"]),
            call_as_contract=d["CallAsContract"],
        )

    def to_cbor_dict(self) -> dict:
        return {
            "Address": self.address.to_cbor_dict(),
            "CallAsContract": self.call_as_contract,
        }


@dataclass
class PayeeMetadata:
    call_as_contract: bool

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "PayeeMetadata":
        return cls(
            call_as_contract=d["CallAsContract"],
        )

    def to_cbor_dict(self) -> dict:
        return {
            "CallAsContract": self.call_as_contract,
        }
