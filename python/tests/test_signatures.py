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

    (store_events_pb2.StoreEvent(update_store_manifest=store_events_pb2.UpdateStoreManifest(field=store_events_pb2.UpdateStoreManifest.MANIFEST_FIELD_DOMAIN)),
      "0x8747c6939554d2f6d4de0be055465d24eeea4be946e2a78dd226c3b9c52747ae"),

    (store_events_pb2.StoreEvent(create_item=store_events_pb2.CreateItem()),
      "0x8225d75532d725939e5025e027de840bb5c9c3d2a13a94d17ce495bdd0cdc594"),

    (store_events_pb2.StoreEvent(update_item=store_events_pb2.UpdateItem(field=store_events_pb2.UpdateItem.ITEM_FIELD_PRICE)),
      "0x92a19570a7e9bc04c58c984ad7203c5bfd94149cf44eda537fe6a0aefbc9da92"),

    (store_events_pb2.StoreEvent(create_tag=store_events_pb2.CreateTag()),
      "0xa517d7d534cba5e6ebaf002fe765653cbd7beb9cf035641f9c9d1f0b5d3bed74"),

    (store_events_pb2.StoreEvent(update_tag=store_events_pb2.UpdateTag(action=store_events_pb2.UpdateTag.TAG_ACTION_ADD_ITEM)),
      "0xa8b3551e44d5020bc5dd518f15a9876bc3aa3e6cffad959654664fd7d6b798d3"),

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
    (store_events_pb2.StoreEvent(update_store_manifest=store_events_pb2.UpdateStoreManifest(field=store_events_pb2.UpdateStoreManifest.MANIFEST_FIELD_DOMAIN, string="cryptix.pizza")),
     "0x52d85f147956d3ce5a2377aa5f5c0f85aeada9aa9123ae91d52f5f3a7e7f5435"),

    (store_events_pb2.StoreEvent(update_store_manifest=store_events_pb2.UpdateStoreManifest(field=store_events_pb2.UpdateStoreManifest.MANIFEST_FIELD_PUBLISHED_TAG, tag_id=test_event_id)),
     "0x18d7dc652fa959e9cc67356421f7af70b9565f2e76290a494427e5740f8b7a8b"),

    (store_events_pb2.StoreEvent(update_store_manifest=store_events_pb2.UpdateStoreManifest(field=store_events_pb2.UpdateStoreManifest.MANIFEST_FIELD_ADD_ERC20, erc20_addr=test_addr)),
     "0x51ea874edda006c21e7f4642269d63a85129758521e8a9b0a325b25a0a542d19"),

    (store_events_pb2.StoreEvent(update_item=store_events_pb2.UpdateItem(field=store_events_pb2.UpdateItem.ITEM_FIELD_PRICE, item_id=test_event_id, price="123.00")),
      "0xcdf10775185454310cf3965afab693e3de0ecb11af57f4cf3013f0341d9b47b4"),

    (store_events_pb2.StoreEvent(update_item=store_events_pb2.UpdateItem(field=store_events_pb2.UpdateItem.ITEM_FIELD_METADATA, item_id=test_event_id, metadata=b'{ "name": "test" }')),
      "0x88b1733cb885c6bf06bc6d3c12b5a7c08dc66b952759d84a10a1f4c1d7ae3130"),

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
