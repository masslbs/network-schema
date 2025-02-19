# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from typing import Union, Dict, Any

from massmarket_hash_event.cbor.ethereum_address import EthereumAddress


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
