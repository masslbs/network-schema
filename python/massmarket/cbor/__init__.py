# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from io import BytesIO

import cbor2
import hashlib
from pprint import pprint

from massmarket.hamt import Trie
from massmarket.cbor.base_types import Tag, Account
from massmarket.cbor.manifest import Manifest
from massmarket.cbor.listing import Listing
from massmarket.cbor.order import Order
from massmarket.cbor_encoder import cbor_encode


@dataclass
class Shop:
    schema_version: int
    manifest: Manifest
    accounts: Trie[Account] = None
    listings: Trie[Listing] = None
    inventory: Trie[int] = None
    tags: Trie[Tag] = None
    orders: Trie[Order] = None

    def __post_init__(self):
        if self.accounts is None:
            self.accounts = Trie.new()
        if self.listings is None:
            self.listings = Trie.new()
        if self.inventory is None:
            self.inventory = Trie.new()
        if self.tags is None:
            self.tags = Trie.new()
        if self.orders is None:
            self.orders = Trie.new()

    def serialize(self) -> dict:
        return {
            "SchemaVersion": self.schema_version,
            "Manifest": self.manifest.to_cbor_dict(),
            "Accounts": self.accounts.to_cbor_array(),
            "Listings": self.listings.to_cbor_array(),
            "Inventory": self.inventory.to_cbor_array(),
            "Tags": self.tags.to_cbor_array(),
            "Orders": self.orders.to_cbor_array(),
        }

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Shop":
        return cls(
            schema_version=d["SchemaVersion"],
            manifest=Manifest.from_cbor_dict(d["Manifest"]),
            accounts=Trie.from_cbor_array(d["Accounts"]),
            listings=Trie.from_cbor_array(d["Listings"]),
            inventory=Trie.from_cbor_array(d["Inventory"]),
            tags=Trie.from_cbor_array(d["Tags"]),
            orders=Trie.from_cbor_array(d["Orders"]),
        )

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "Shop":
        return cls.from_cbor_dict(cbor2.loads(cbor_data))

    def hash(self) -> bytes:
        hashed = {
            "SchemaVersion": self.schema_version,
            "Manifest": self.manifest.to_cbor_dict(),
            "Accounts": self.accounts.hash(),
            "Listings": self.listings.hash(),
            "Inventory": self.inventory.hash(),
            "Tags": self.tags.hash(),
            "Orders": self.orders.hash(),
        }
        return hash_object(hashed)


def hash_object(obj: dict) -> bytes:
    # print(obj)
    encoded = cbor_encode(obj)
    # print(encoded.hex())
    return hashlib.sha256(encoded).digest()
