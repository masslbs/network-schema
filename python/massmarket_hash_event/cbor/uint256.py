# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from typing import Union
import os
from decimal import Decimal


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
