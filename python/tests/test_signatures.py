# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import pytest
import binascii
import json
import random
random.seed("massmarket-testing")

from web3 import Account, Web3

from massmarket_hash_event import Hasher, store_events_pb2

def test_correct_contract_addr():
  with pytest.raises(Exception) as ex:
    Hasher(123, "aaa")
  assert str(ex.value) == "Invalid contract address: aaa"
  with pytest.raises(Exception) as ex:
    Hasher(123, "0xabc")
  assert str(ex.value) == 'Odd-length string'
  with pytest.raises(Exception) as ex:
    Hasher(123, "0xabcd")
  assert str(ex.value) == "Invalid contract address: 0xabcd"
  h = Hasher(123, "0x1234567890123456789012345678901234567890")
  assert h is not None

def test_hash_empty_event():
  h = Hasher(2342, "0x0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a0a")
  pk = Account.from_key("0x1234567890123456789012345678901234567890123456789012345678901234")
  events = [
    (store_events_pb2.StoreEvent(store_manifest=store_events_pb2.StoreManifest()),
     "0x3d4abe03c92aa50e2b2d59420ea24842e2455690ef1f715e245afe5da9618040"),

    (store_events_pb2.StoreEvent(update_store_manifest=store_events_pb2.UpdateStoreManifest()),
      "0xf14b1ced4f07519d89e82d4d56bddf367cbf071a47e507fde2061858bace9142"),

    (store_events_pb2.StoreEvent(create_item=store_events_pb2.CreateItem()),
      "0x8225d75532d725939e5025e027de840bb5c9c3d2a13a94d17ce495bdd0cdc594"),

    (store_events_pb2.StoreEvent(update_item=store_events_pb2.UpdateItem()),
      "0xdd672db57e0c5a8c4f3b8a0990cb6fdd1a7e9120b6ced01aa85646b78897389b"),

    (store_events_pb2.StoreEvent(create_tag=store_events_pb2.CreateTag()),
      "0xa517d7d534cba5e6ebaf002fe765653cbd7beb9cf035641f9c9d1f0b5d3bed74"),

    (store_events_pb2.StoreEvent(update_tag=store_events_pb2.UpdateTag()),
      "0x28a94a4382cecf49a5cfd602187567f2719e75a9120cea36c9b14f1d47f7e4cd"),

    (store_events_pb2.StoreEvent(create_order=store_events_pb2.CreateOrder()),
      "0x94b3409f62d4be96bbcc399b8eb153e331607e569f1993eadb24f715c86cd2e3"),

    (store_events_pb2.StoreEvent(update_order=store_events_pb2.UpdateOrder(
      order_canceled=store_events_pb2.UpdateOrder.OrderCanceled(timestamp=1))),
      "0x3ec3e627a9912a1f7bf17e33107870af946f39ecd5a8545c5cadfa588fe61982"),

    (store_events_pb2.StoreEvent(change_stock=store_events_pb2.ChangeStock()),
      "0x06dda3da556003d920246f05238d29717768408673ebc209de3fd128391cf4d8"),

    (store_events_pb2.StoreEvent(new_key_card=store_events_pb2.NewKeyCard(user_wallet_addr=random.randbytes(20))),
      "0xd4002eaa6e53a6cc6b79a056056eee0289d9dcddbdf48c15beaf65039372e78b")
  ]
  for idx, (evt, expected) in enumerate(events):
    data = h.hash_event(evt)
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
  h = Hasher(vec_sigs['chain_id'], vec_sigs['contract_address'])
  for idx, evt in enumerate(vector['events']):
    parsed = store_events_pb2.StoreEvent()
    parsed.ParseFromString(bytes.fromhex(evt['encoded']))
    assert len(parsed.signature) == 65, f"invalid signature on event {idx}"
    print(f"hashing {idx}")
    encoded_data = h.hash_event(parsed)
    pub_key = Account.recover_message(encoded_data, signature=parsed.signature)
    their_addr = Web3.to_checksum_address(pub_key)
    assert their_addr == signer, f"invalid signer on event {idx}"

def test_optional_fields():
  h = Hasher(2342, "0x0000000000000000000000000000000000000000")
  pk = Account.from_key("0x1234567890123456789012345678901234567890123456789012345678901234")
  test_event_id = binascii.unhexlify("beef" * 16)
  assert len(test_event_id) == 32
  test_addr = bytes(bytearray(20))
  events = [
    (store_events_pb2.StoreEvent(update_store_manifest=store_events_pb2.UpdateStoreManifest(domain="cryptix.pizza")),
     "0xe9ad42640da17d9a7e7a3ac81c861196cfbc80223537f4b099e983b85a6cf30f"),

    (store_events_pb2.StoreEvent(update_store_manifest=store_events_pb2.UpdateStoreManifest(published_tag_id=test_event_id)),
     "0x8e61569e9253d3c2361def961997e1f74f845c3d30f8b5cd27fd2f82902f8340"),

    (store_events_pb2.StoreEvent(update_store_manifest=store_events_pb2.UpdateStoreManifest(add_erc20_addr=test_addr)),
     "0x1743f110452b093c2e975fece1cc75378aaddf667c69343014f42a421f307396"),

    (store_events_pb2.StoreEvent(update_item=store_events_pb2.UpdateItem(item_id=test_event_id, price="123.00")),
      "0xd15c3125c09e8ce65d4bce08434d065a339ea3d4b2f9e9a410926c16f5d3fa84"),

    (store_events_pb2.StoreEvent(update_item=store_events_pb2.UpdateItem(item_id=test_event_id, metadata=b'{ "name": "test" }')),
      "0x16fe5abbbe35d9e269970db6a844f9df697b809a36fc9c36d245685a887ad592"),

    (store_events_pb2.StoreEvent(change_stock=store_events_pb2.ChangeStock(item_ids=[test_event_id], diffs=[1])),
      "0xe6c9896786eaa08950cb49cca8ed1ed5fb3c128979ea7bf0d0187e478fcd88d1"),

    (store_events_pb2.StoreEvent(change_stock=store_events_pb2.ChangeStock(item_ids=[test_event_id], diffs=[1], order_id=test_event_id, tx_hash=bytes(bytearray(32)))),
      "0x20d541a8d06cbbde3810533a4314b939eebe43ab19d0a9b134c7d5c2a2f379f8"),
  ]
  for idx, (evt, expected) in enumerate(events):
    data = h.hash_event(evt)
    signed_message = pk.sign_message(data)
    msg_hash = signed_message.messageHash.hex()
    assert msg_hash == expected, f"Failed on event {idx}"
