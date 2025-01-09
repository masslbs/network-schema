# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

__all__ = [
    "hash_event",
    "transport_pb2",
    "authentication_pb2",
    "shop_pb2",
    "shop_requests_pb2",
    "error_pb2",
    "shop_events_cbor",
    "storage_pb2",
]

import cbor2
from eth_account import messages

from massmarket_hash_event import shop_events_cbor


def hash_patchset(evt: shop_events_cbor.PatchSet):
    encoded = cbor2.dumps(evt)
    return messages.encode_defunct(encoded)
