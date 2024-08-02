# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import json
import random
random.seed("massmarket-testing")

from web3 import Account, Web3

from massmarket_hash_event import hash_event, base_types_pb2 as mtypes, shop_events_pb2 as mevents

def unhex(b):
  if b.startswith("0x"):
    b = b[2:]
  return bytes.fromhex(b)

def test_hash_empty_event():
  pk = Account.from_key("0x1234567890123456789012345678901234567890123456789012345678901234")
  events = [
    (mevents.ShopEvent(manifest=mevents.Manifest()),
      "81441024380a9ea9aef75d56c2084d5066c9e41675485baf9eeb4498ee78c2b6"),

    (mevents.ShopEvent(update_manifest=mevents.UpdateManifest()),
      "7361a3cd19635c22194e5ac719577c0058a2d7e78333a759364d066d186c4aa5"),

    (mevents.ShopEvent(listing=mevents.Listing()),
      "a42cebb4dae1107494c6a8675c95cdd9231d799284cbd33b8377f4af96063ecb"),

    (mevents.ShopEvent(update_listing=mevents.UpdateListing()),
      "e923fcad78e694a6b3b36f9d23db5c0785d0fae691e81eb42d9f6f00b9bd1d4c"),

    (mevents.ShopEvent(tag=mevents.Tag()),
      "fff15ebd53ce762b85eaed62ad69744ce2bc423f6c895658ec344b195ceaa524"),

    (mevents.ShopEvent(update_tag=mevents.UpdateTag()),
      "5f9b91d87e007365d26d5dfe7b3f6abbb40f01bfd49d536d7a54cb1ca13a739c"),

    (mevents.ShopEvent(create_order=mevents.CreateOrder()),
      "418c48952bc438a3e32c2afb241d836c0d0fa303273211229c3e6ca44809e763"),

    (mevents.ShopEvent(update_order=mevents.UpdateOrder()),
      "23b205f0bca3a484edb8c86aefa48469286dbe5195fd160142bc0041b7b1b6a8"),

    (mevents.ShopEvent(change_inventory=mevents.ChangeInventory()),
      "e746bbdbaa60a9e78bb8dc9eb2bf500c071f6883be67829850bf51573e7cc3e9"),

    (mevents.ShopEvent(account=mevents.Account()),
      "50933230f8c622f4b480f875a8ef319550c163ad724a06e3a85132b71a8aea24")
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
  assert signer == "0xB8b8985e55aBEa8E36C777c28C08ECBe0104a37d"
  for idx, evt in enumerate(vector['events']):
    parsed = mevents.ShopEvent()
    parsed.ParseFromString(unhex(evt['encoded']))
    print(f"hashing {idx}")
    encoded_data = hash_event(parsed)
    pub_key = Account.recover_message(encoded_data, signature=unhex(evt["signature"]))
    their_addr = Web3.to_checksum_address(pub_key)
    assert their_addr == signer, f"invalid signer on event {idx}"

def test_optional_fields():
  pk = Account.from_key("0x1234567890123456789012345678901234567890123456789012345678901234")
  test_id = 2342
  test_addr = bytes(20)
  test_currency = mtypes.ShopCurrency(chain_id=42, address=mtypes.EthereumAddress(raw=test_addr))
  test_price = mtypes.Uint256(raw=int(0).to_bytes(32, 'big'))
  events = [
    (mevents.ShopEvent(update_manifest=mevents.UpdateManifest(set_base_currency=test_currency)),
     "d1d3cf16b228fa92e53516dcf593b87144aab079933b494f7de98219d9c0e783"),

    (mevents.ShopEvent(update_listing=mevents.UpdateListing(listing_id=test_id, base_price=test_price)),
      "9f96e73818875daff93b09172ee453cb7d7b6426769b28f0be35fe10ab8b094d"),

    (mevents.ShopEvent(update_order=mevents.UpdateOrder(order_id=test_id, canceled=mevents.UpdateOrder.Canceled())),
     "670045c5508875928a0f24f9a4a002ed0939e744530f99934c8403b1a892033d"),

    (mevents.ShopEvent(change_inventory=mevents.ChangeInventory(listing_id=test_id, diff=1)),
      "92cb832482f77592e8c1da65d2667fb270356d3491a9b00ebfe49e9265d7e784"),

    (mevents.ShopEvent(change_inventory=mevents.ChangeInventory(listing_id=test_id, diff=1, variation_ids=[1, 2, 3])),
      "28373b514955d2ee579fd567cbeed4adecf55c8e171fc38181ea5a2276ef5a04"),
  ]
  for idx, (evt, expected) in enumerate(events):
    data = hash_event(evt)
    signed_message = pk.sign_message(data)
    msg_hash = signed_message.messageHash.hex()
    assert msg_hash == expected, f"Failed on event {idx}"
