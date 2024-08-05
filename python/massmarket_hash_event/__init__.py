# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

__all__ = [
    "hash_event",
    "transport_pb2",
    "authentication_pb2",
    "shop_pb2",
    "shop_requests_pb2",
    "error_pb2",
    "shop_events_pb2",
]

import json
import binascii
from pprint import pprint
from importlib.resources import files
from eth_account import messages

from massmarket_hash_event import shop_events_pb2


def hash_event(evt: shop_events_pb2.ShopEvent):
    encoded = evt.SerializeToString()
    return messages.encode_defunct(encoded)
