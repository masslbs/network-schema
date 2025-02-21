# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from io import BytesIO

import cbor2

from massmarket_hash_event.hamt import Trie
from massmarket_hash_event.cbor.manifest import Manifest
from massmarket_hash_event.cbor.listing import Listing


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
    listings: Trie[Listing]

    def to_cbor_dict(self) -> dict:
        return {
            "SchemaVersion": self.schema_version,
            "Manifest": self.manifest.to_cbor_dict(),
            "Listings": self.listings.to_cbor_array(),
        }

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Shop":
        return cls(
            schema_version=d["SchemaVersion"],
            manifest=Manifest.from_cbor_dict(d["Manifest"]),
            listings=Trie.from_cbor_array(d["Listings"]),
        )

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "Shop":
        return cls.from_cbor_dict(cbor2.loads(cbor_data))
