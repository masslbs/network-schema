# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import pytest
import binascii
import json
import random
random.seed("massmarket-testing")

from web3 import Account, Web3

from massmarket_hash_event import Hasher, schema_pb2

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
    (schema_pb2.Event(store_manifest=schema_pb2.StoreManifest()),
     "0x3d4abe03c92aa50e2b2d59420ea24842e2455690ef1f715e245afe5da9618040"),

    (schema_pb2.Event(update_manifest=schema_pb2.UpdateManifest(field=schema_pb2.UpdateManifest.MANIFEST_FIELD_DOMAIN)),
      "0x29bbab0c3b2a448fba1cdd96cdf0ff95ca0000b529bc6fec098667bee841f106"),

    (schema_pb2.Event(create_item=schema_pb2.CreateItem()),
      "0x8225d75532d725939e5025e027de840bb5c9c3d2a13a94d17ce495bdd0cdc594"),

    (schema_pb2.Event(update_item=schema_pb2.UpdateItem(field=schema_pb2.UpdateItem.ITEM_FIELD_PRICE)),
      "0x92a19570a7e9bc04c58c984ad7203c5bfd94149cf44eda537fe6a0aefbc9da92"),

    (schema_pb2.Event(create_tag=schema_pb2.CreateTag()),
      "0xa517d7d534cba5e6ebaf002fe765653cbd7beb9cf035641f9c9d1f0b5d3bed74"),

    (schema_pb2.Event(add_to_tag=schema_pb2.AddToTag()),
      "0xbc870529d724c8184c5c350aa28513b0d142b09cd54f6308dc6538c7bd25a2f0"),

    (schema_pb2.Event(remove_from_tag=schema_pb2.RemoveFromTag()),
      "0x83e0a83733d803fd6964123fa3b5fe96efb1390c80069ad10e065d5bfd761ade"),

    (schema_pb2.Event(rename_tag=schema_pb2.RenameTag()),
      "0x2fb4f99089834b30f27ecf356cb1bab0cdd3a0299491fc7f6f154799de626518"),

    (schema_pb2.Event(delete_tag=schema_pb2.DeleteTag()),
      "0x4e180b73a08f78f2f170a2d014832035a095debf51ea423f8c2f77ea2b068580"),

    (schema_pb2.Event(create_cart=schema_pb2.CreateCart()),
      "0x86d113747e19f65535e20f14ca9526e683bd0255e4b8fdaaffa87e0e1a840854"),

    (schema_pb2.Event(change_cart=schema_pb2.ChangeCart()),
      "0xe033b1f86a7ebe1eacdbfb332c6fe7ea0294d1eebff55b5910a248d3d2baff2b"),

    (schema_pb2.Event(cart_finalized=schema_pb2.CartFinalized(purchase_addr=random.randbytes(20))),
      "0xb890ad659de9ae16576c7a6b0267ccdca64e7f771cf8853f6f33b411118ed559"),

    (schema_pb2.Event(cart_abandoned=schema_pb2.CartAbandoned()),
      "0xaeb33fbed4304027350b36c723fba123e9c954422617595ec29e7c86730aa986"),

    (schema_pb2.Event(change_stock=schema_pb2.ChangeStock()),
      "0x06dda3da556003d920246f05238d29717768408673ebc209de3fd128391cf4d8"),

    (schema_pb2.Event(new_key_card=schema_pb2.NewKeyCard(user_wallet_addr=random.randbytes(20))),
      "0xbbdc5122897f0772c86eb5b79bbc6d8c5cf917886c8138678f4c9b2ff4495a21")
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
    parsed = schema_pb2.Event()
    parsed.ParseFromString(bytes.fromhex(evt['encoded']))
    assert len(parsed.signature) == 65, f"invalid signature on event {idx}"
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
    (schema_pb2.Event(update_manifest=schema_pb2.UpdateManifest(field=schema_pb2.UpdateManifest.MANIFEST_FIELD_DOMAIN, string="cryptix.pizza")),
     "0x21cf3bfc85a9e4662a7772aad753dc80c3766da6b92048a5aa2a40821ce64076"),

    (schema_pb2.Event(update_manifest=schema_pb2.UpdateManifest(field=schema_pb2.UpdateManifest.MANIFEST_FIELD_PUBLISHED_TAG, tag_id=test_event_id)),
     "0x2423c7f055c8d62ec772bbca830086dcf5774b99e16b651247e1178bc5a94a93"),

    (schema_pb2.Event(update_manifest=schema_pb2.UpdateManifest(field=schema_pb2.UpdateManifest.MANIFEST_FIELD_ADD_ERC20, erc20_addr=test_addr)),
     "0xc714bd0124f3bf8b15a2ea6911738c18e1e1beb33b82ca316cb3fe9437b19c62"),

    (schema_pb2.Event(update_item=schema_pb2.UpdateItem(field=schema_pb2.UpdateItem.ITEM_FIELD_PRICE, item_id=test_event_id, price="123.00")),
      "0xcdf10775185454310cf3965afab693e3de0ecb11af57f4cf3013f0341d9b47b4"),

    (schema_pb2.Event(update_item=schema_pb2.UpdateItem(field=schema_pb2.UpdateItem.ITEM_FIELD_METADATA, item_id=test_event_id, metadata=b'{ "name": "test" }')),
      "0x88b1733cb885c6bf06bc6d3c12b5a7c08dc66b952759d84a10a1f4c1d7ae3130"),

    (schema_pb2.Event(change_stock=schema_pb2.ChangeStock(item_ids=[test_event_id], diffs=[1])),
      "0xe6c9896786eaa08950cb49cca8ed1ed5fb3c128979ea7bf0d0187e478fcd88d1"),

    (schema_pb2.Event(change_stock=schema_pb2.ChangeStock(item_ids=[test_event_id], diffs=[1], cart_id=test_event_id, tx_hash=bytes(bytearray(32)))),
      "0x5fb473c812fc700fb609043a5760565fbc14551645b3efd79bb6766b13bd9c22"),
  ]
  for idx, (evt, expected) in enumerate(events):
    data = h.hash_event(evt)
    signed_message = pk.sign_message(data)
    msg_hash = signed_message.messageHash.hex()
    assert msg_hash == expected, f"Failed on event {idx}"
