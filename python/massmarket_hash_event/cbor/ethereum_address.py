# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import re
from typing import Union

class EthereumAddress:
    """Represents an Ethereum address"""
    
    SIZE = 20  # common.AddressLength in Go

    def __init__(self, value: Union[str, bytes, "EthereumAddress"]):
        if isinstance(value, EthereumAddress):
            self._bytes = value._bytes
        elif isinstance(value, str):
            # Strip 0x prefix if present
            hex_str = value.lower().replace('0x', '')
            if not re.match(r'^[0-9a-f]{40}$', hex_str):
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
        if not isinstance(other, EthereumAddress):
            raise ValueError("Cannot compare EthereumAddress with non-EthereumAddress")
        return self._bytes == other._bytes

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