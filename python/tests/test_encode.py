# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import binascii
import random

random.seed("massmarket-testing")

from massmarket_hash_event import shop_events_pb2 as mevents, base_types_pb2 as mtypes


def hex(b: bytes) -> str:
    return binascii.hexlify(b).decode("utf-8")


def test_encode_event():
    manifest = mevents.Manifest(token_id=mtypes.Uint256(raw=random.randbytes(32)))
    assert (
        hex(manifest.token_id.raw)
        == "bdb2914879b87165be2f3f51555499d06df7c08c77b7511b4efcaeadbf1f566a"
    )
    data = manifest.SerializeToString()
    assert (
        hex(data)
        == "0a220a20bdb2914879b87165be2f3f51555499d06df7c08c77b7511b4efcaeadbf1f566a"
    )
    event = mevents.ShopEvent(manifest=manifest)
    data = event.SerializeToString()
    assert (
        hex(data)
        == "22240a220a20bdb2914879b87165be2f3f51555499d06df7c08c77b7511b4efcaeadbf1f566a"
    )
