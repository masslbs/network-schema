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
    "storage_pb2",
]

import cbor2
from eth_account import messages

def hash_patchset(evt):
    encoded = cbor2.dumps(evt)
    return messages.encode_defunct(encoded)
