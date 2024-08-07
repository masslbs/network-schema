# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import pytest
import binascii
import json
import random
random.seed("massmarket-testing")

from web3 import Account, Web3

from massmarket_hash_event import hash_event, shop_pb2, shop_events_pb2

def unhex(b):
  if b.startswith("0x"):
    b = b[2:]
  return bytes.fromhex(b)

def test_hash_empty_event():
  pk = Account.from_key("0x1234567890123456789012345678901234567890123456789012345678901234")
  events = [
    (shop_events_pb2.ShopEvent(shop_manifest=shop_events_pb2.ShopManifest()),
      "5b2347782d89e88fbb2f0ecb9edbcd02ae0767450e9460762a31b17056a62d95"),

    (shop_events_pb2.ShopEvent(update_shop_manifest=shop_events_pb2.UpdateShopManifest()),
      "b5ea4e7477b0ce09e3cfd7578b832ab3622d38eccb66274585d8850471a44ff7"),

    (shop_events_pb2.ShopEvent(create_item=shop_events_pb2.CreateItem()),
      "81441024380a9ea9aef75d56c2084d5066c9e41675485baf9eeb4498ee78c2b6"),

    (shop_events_pb2.ShopEvent(update_item=shop_events_pb2.UpdateItem()),
      "7361a3cd19635c22194e5ac719577c0058a2d7e78333a759364d066d186c4aa5"),

    (shop_events_pb2.ShopEvent(create_tag=shop_events_pb2.CreateTag()),
      "50933230f8c622f4b480f875a8ef319550c163ad724a06e3a85132b71a8aea24"),

    (shop_events_pb2.ShopEvent(update_tag=shop_events_pb2.UpdateTag()),
      "a42cebb4dae1107494c6a8675c95cdd9231d799284cbd33b8377f4af96063ecb"),

    (shop_events_pb2.ShopEvent(create_order=shop_events_pb2.CreateOrder()),
      "e923fcad78e694a6b3b36f9d23db5c0785d0fae691e81eb42d9f6f00b9bd1d4c"),

    (shop_events_pb2.ShopEvent(update_order=shop_events_pb2.UpdateOrder(
      order_canceled=shop_events_pb2.UpdateOrder.OrderCanceled(timestamp=1))),
      "0f1127079e743a5acb7048b205f898ce4e867f02bc7a2bbcecc9798b552f254d"),

    (shop_events_pb2.ShopEvent(change_stock=shop_events_pb2.ChangeStock()),
      "418c48952bc438a3e32c2afb241d836c0d0fa303273211229c3e6ca44809e763"),

    (shop_events_pb2.ShopEvent(new_key_card=shop_events_pb2.NewKeyCard(user_wallet_addr=random.randbytes(20))),
      "42b06e7f9e2c3a4162679c2f938bc0b3d777de6acbc11e7a44df944337632759")
  ]
  for idx, (evt, expected) in enumerate(events):
    data = hash_event(evt)
    signed_message = pk.sign_message(data)
    msg_hash = signed_message.messageHash.hex()
    evt_name = evt.WhichOneof("union")
    assert msg_hash == expected, f"Failed on event {idx} ({evt_name})"


from pprint import pprint

# check that the test vectors we generated are valid
def test_verify_vector_file():
  with open("../testVectors.json") as f:
    vector = json.load(f)
  assert len(vector['events']) > 0
  assert "signatures" in vector
  vec_sigs = vector['signatures']
  signer = vec_sigs['signer_address']
  assert signer == "0x27B369BDD9b49C322D13e7E91d83cFD47d465713"
  for idx, evt in enumerate(vector['events']):
    parsed = shop_events_pb2.ShopEvent()
    parsed.ParseFromString(unhex(evt['encoded']))
    print(f"hashing {idx}")
    encoded_data = hash_event(parsed)
    pub_key = Account.recover_message(encoded_data, signature=unhex(evt["signature"]))
    their_addr = Web3.to_checksum_address(pub_key)
    assert their_addr == signer, f"invalid signer on event {idx}"

def test_optional_fields():
  pk = Account.from_key("0x1234567890123456789012345678901234567890123456789012345678901234")
  test_event_id = binascii.unhexlify("beef" * 16)
  assert len(test_event_id) == 32
  test_addr = bytes(20)
  test_currency = shop_pb2.ShopCurrency(chain_id=42, token_addr=test_addr)
  events = [
    (shop_events_pb2.ShopEvent(update_shop_manifest=shop_events_pb2.UpdateShopManifest(domain="cryptix.pizza")),
     "8a492edbcde76fabb5289ef6cbce5497d6fe7f47f40d2ecd08ef8464a3df6728"),

    (shop_events_pb2.ShopEvent(update_shop_manifest=shop_events_pb2.UpdateShopManifest(published_tag_id=test_event_id)),
     "4e38b37be03dfafdd2ce4801d06d6fbbe444cd01ff4b8dc6b1c68855df154432"),

    (shop_events_pb2.ShopEvent(update_shop_manifest=shop_events_pb2.UpdateShopManifest(set_base_currency=test_currency)),
     "3e7f32f280d8fcc767a9494e320a7b0bbb3c8165d2b90453070a2dc9af22e752"),

    (shop_events_pb2.ShopEvent(update_item=shop_events_pb2.UpdateItem(item_id=test_event_id, price="123.00")),
      "d3e8a09891568e3bfb5b3556a0de093c1906d9609b2992021fb20aa227a8ba85"),

    (shop_events_pb2.ShopEvent(update_item=shop_events_pb2.UpdateItem(item_id=test_event_id, metadata=b'{ "name": "test" }')),
      "0ee4003b9fc7c111e53c4a085b59898e9a0ace380044c7f7c945b5983a360dd3"),

    (shop_events_pb2.ShopEvent(change_stock=shop_events_pb2.ChangeStock(item_ids=[test_event_id], diffs=[1])),
      "b5d719f8e03391080d93585110d2b1fe6ba5f3dc7340dc92eef60cbee3bb3daa"),

    (shop_events_pb2.ShopEvent(change_stock=shop_events_pb2.ChangeStock(item_ids=[test_event_id], diffs=[1], order_id=test_event_id, tx_hash=bytes(bytearray(32)))),
      "a4053e374ff7e6c404ff9d577eda736ae36bedf26a4a8a877f2b0f7814a29f95"),
  ]
  for idx, (evt, expected) in enumerate(events):
    data = hash_event(evt)
    signed_message = pk.sign_message(data)
    msg_hash = signed_message.messageHash.hex()
    assert msg_hash == expected, f"Failed on event {idx}"
