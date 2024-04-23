# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import binascii
import random
random.seed("massmarket-testing")

from massmarket_hash_event import schema_pb2

def hex(b: bytes) -> str:
  return binascii.hexlify(b).decode('utf-8')

def test_encode_event():
  manifest = schema_pb2.StoreManifest()
  manifest.event_id = random.randbytes(32)
  assert hex(manifest.event_id) == "bdb2914879b87165be2f3f51555499d06df7c08c77b7511b4efcaeadbf1f566a"
  manifest.store_token_id = random.randbytes(32)
  manifest.published_tag_id = random.randbytes(32)
  manifest.domain = "shop.mass.market"
  data = manifest.SerializeToString()
  assert hex(data) == "0a20bdb2914879b87165be2f3f51555499d06df7c08c77b7511b4efcaeadbf1f566a122087e7f0ff4b672cbbf634364fde123b58402da96fc46df0414f4952d85ffc9f6f1a1073686f702e6d6173732e6d61726b65742220b026c43b2358eed1f57233c2e8300c3dba995cd8b0377216fca829e211e6805c"
  event = schema_pb2.Event(store_manifest=manifest)
  data = event.SerializeToString()
  assert hex(data) == "12780a20bdb2914879b87165be2f3f51555499d06df7c08c77b7511b4efcaeadbf1f566a122087e7f0ff4b672cbbf634364fde123b58402da96fc46df0414f4952d85ffc9f6f1a1073686f702e6d6173732e6d61726b65742220b026c43b2358eed1f57233c2e8300c3dba995cd8b0377216fca829e211e6805c"
