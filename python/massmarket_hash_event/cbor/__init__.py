# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from io import BytesIO

import cbor2
import hashlib
from pprint import pprint
from massmarket_hash_event.hamt import Trie
from massmarket_hash_event.cbor.manifest import Manifest
from massmarket_hash_event.cbor.listing import Listing
from massmarket_hash_event.cbor.base_types import Tag, Account


# construct default encoder
def cbor_encode(obj):
    with BytesIO() as fp:
        cbor2.CBOREncoder(
            fp,
            canonical=True,
            date_as_datetime=True,
        ).encode(obj)
        return fp.getvalue()


@dataclass
class Shop:
    schema_version: int
    manifest: Manifest
    accounts: Trie[Account]
    listings: Trie[Listing]
    inventory: Trie
    tags: Trie[Tag]
    orders: Trie

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
        # pprint(cbor_encode(hashed).hex())
        return hash_object(hashed)


def hash_object(obj: dict) -> bytes:
    return hashlib.sha256(cbor_encode(obj)).digest()


