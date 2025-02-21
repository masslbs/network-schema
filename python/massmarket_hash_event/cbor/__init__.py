# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from typing import Dict, List, Optional
from io import BytesIO

import cbor2

# from massmarket_hash_event.hamt import Trie
from massmarket_hash_event.cbor.uint256 import Uint256
# from massmarket_hash_event.cbor.listing import Listing

# construct default encoder
def cbor_encode(obj):
    with BytesIO() as fp:
        cbor2.CBOREncoder(
            fp,
            canonical=True,
            date_as_datetime=True,
        ).encode(obj)
        return fp.getvalue()

# @dataclass
# class Shop:
#     version: int
#     manifest: Manifest
#     listings: Trie[Listing]

#     def to_cbor_dict(self) -> dict:
#         return {
#             "Version": self.version,
#             "Manifest": self.manifest.to_cbor_dict(),
#             "Listings": self.listings.to(),
#         }