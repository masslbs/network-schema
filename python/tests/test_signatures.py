# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import pytest
import binascii
import json
import random
random.seed("massmarket-testing")

from web3 import Account, Web3

from massmarket_hash_event import hash_event, shop_events_pb2

def test_hash_empty_event():
  pk = Account.from_key("0x1234567890123456789012345678901234567890123456789012345678901234")
  events = [
    (shop_events_pb2.ShopEvent(shop_manifest=shop_events_pb2.ShopManifest()),
      "0x5b2347782d89e88fbb2f0ecb9edbcd02ae0767450e9460762a31b17056a62d95"),

    (shop_events_pb2.ShopEvent(update_shop_manifest=shop_events_pb2.UpdateShopManifest()),
      "0xb5ea4e7477b0ce09e3cfd7578b832ab3622d38eccb66274585d8850471a44ff7"),

    (shop_events_pb2.ShopEvent(create_item=shop_events_pb2.CreateItem()),
      "0x81441024380a9ea9aef75d56c2084d5066c9e41675485baf9eeb4498ee78c2b6"),

    (shop_events_pb2.ShopEvent(update_item=shop_events_pb2.UpdateItem()),
      "0x7361a3cd19635c22194e5ac719577c0058a2d7e78333a759364d066d186c4aa5"),

    (shop_events_pb2.ShopEvent(create_tag=shop_events_pb2.CreateTag()),
      "0x50933230f8c622f4b480f875a8ef319550c163ad724a06e3a85132b71a8aea24"),

    (shop_events_pb2.ShopEvent(update_tag=shop_events_pb2.UpdateTag()),
      "0xa42cebb4dae1107494c6a8675c95cdd9231d799284cbd33b8377f4af96063ecb"),

    (shop_events_pb2.ShopEvent(create_order=shop_events_pb2.CreateOrder()),
      "0xe923fcad78e694a6b3b36f9d23db5c0785d0fae691e81eb42d9f6f00b9bd1d4c"),

    (shop_events_pb2.ShopEvent(update_order=shop_events_pb2.UpdateOrder(
      order_canceled=shop_events_pb2.UpdateOrder.OrderCanceled(timestamp=1))),
      "0x0f1127079e743a5acb7048b205f898ce4e867f02bc7a2bbcecc9798b552f254d"),

    (shop_events_pb2.ShopEvent(change_stock=shop_events_pb2.ChangeStock()),
      "0x418c48952bc438a3e32c2afb241d836c0d0fa303273211229c3e6ca44809e763"),

    (shop_events_pb2.ShopEvent(new_key_card=shop_events_pb2.NewKeyCard(user_wallet_addr=random.randbytes(20))),
      "0x42b06e7f9e2c3a4162679c2f938bc0b3d777de6acbc11e7a44df944337632759")
  ]
  for idx, (evt, expected) in enumerate(events):
    data = hash_event(evt)
    signed_message = pk.sign_message(data)
    msg_hash = signed_message.messageHash.hex()
    evt_name = evt.WhichOneof("union")
    assert msg_hash == expected, f"Failed on event {idx} ({evt_name})"

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
    parsed.ParseFromString(bytes.fromhex(evt['encoded']))
    assert len(parsed.signature) == 65, f"invalid signature on event {idx}"
    print(f"hashing {idx}")
    encoded_data = hash_event(parsed)
    pub_key = Account.recover_message(encoded_data, signature=parsed.signature)
    their_addr = Web3.to_checksum_address(pub_key)
    assert their_addr == signer, f"invalid signer on event {idx}"

def test_optional_fields():
  pk = Account.from_key("0x1234567890123456789012345678901234567890123456789012345678901234")
  test_event_id = binascii.unhexlify("beef" * 16)
  assert len(test_event_id) == 32
  test_addr = bytes(20)
  test_currency = shop_events_pb2.UpdateShopManifest.ShopCurrency(chain=42, addr=test_addr)
  events = [
    (shop_events_pb2.ShopEvent(update_shop_manifest=shop_events_pb2.UpdateShopManifest(domain="cryptix.pizza")),
     "0x96819fb3b634e1c6d690f75ba43279f1eb2a23016ae4db0a00b1e7badb1b7fa2"),

    (shop_events_pb2.ShopEvent(update_shop_manifest=shop_events_pb2.UpdateShopManifest(published_tag_id=test_event_id)),
     "0x8bee526dd02c85ceeaac7217ee74eca062709654d5cd2b976e37b4feeb6fecd1"),

    (shop_events_pb2.ShopEvent(update_shop_manifest=shop_events_pb2.UpdateShopManifest(set_base_currency=test_currency)),
     "0x2a07e820193194f25a07940f9f41c0196c3411f51abd065c8ac2f2190b98cfdb"),

    (shop_events_pb2.ShopEvent(update_item=shop_events_pb2.UpdateItem(item_id=test_event_id, price="123.00")),
      "0xd15c3125c09e8ce65d4bce08434d065a339ea3d4b2f9e9a410926c16f5d3fa84"),

    (shop_events_pb2.ShopEvent(update_item=shop_events_pb2.UpdateItem(item_id=test_event_id, metadata=b'{ "name": "test" }')),
      "0x16fe5abbbe35d9e269970db6a844f9df697b809a36fc9c36d245685a887ad592"),

    (shop_events_pb2.ShopEvent(change_stock=shop_events_pb2.ChangeStock(item_ids=[test_event_id], diffs=[1])),
      "0xe6c9896786eaa08950cb49cca8ed1ed5fb3c128979ea7bf0d0187e478fcd88d1"),

    (shop_events_pb2.ShopEvent(change_stock=shop_events_pb2.ChangeStock(item_ids=[test_event_id], diffs=[1], order_id=test_event_id, tx_hash=bytes(bytearray(32)))),
      "0x20d541a8d06cbbde3810533a4314b939eebe43ab19d0a9b134c7d5c2a2f379f8"),
  ]
  for idx, (evt, expected) in enumerate(events):
    data = hash_event(evt)
    signed_message = pk.sign_message(data)
    msg_hash = signed_message.messageHash.hex()
    assert msg_hash == expected, f"Failed on event {idx}"
