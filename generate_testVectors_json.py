# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

import os
import sys
import binascii
import json
import pprint
import random
random.seed("mass-market-test-vectors")

from protobuf_to_dict import protobuf_to_dict

from massmarket_hash_event import Hasher
hasher = Hasher(31337, "0x0000000000000000000000001234567890abcdef")

from web3 import Account

from eth_keys import keys
def public_key_from_account(account):
    k = keys.PrivateKey(account.key)
    return k.public_key

def debug(message):
  if os.getenv('DEBUG') != None:
    sys.stderr.write(message + '\n')

def hex(b):
  return binascii.hexlify(b).decode('utf-8')

import schema_pb2

kc1 = Account.from_key(random.randbytes(32))
debug(f"kc1: {kc1.address}")
kc2 = Account.from_key(random.randbytes(32))
debug(f"kc2: {kc2.address}")

events = []

manifest = schema_pb2.StoreManifest()
manifest.event_id = random.randbytes(32)
manifest.store_token_id = random.randbytes(32)
manifest.domain = "https://masslbs.xyz"
manifest.published_tag_id = random.randbytes(32)
events.append(manifest)

newKc1 = schema_pb2.NewKeyCard()
newKc1.event_id = random.randbytes(32)
newKc1.card_public_key = public_key_from_account(kc1).to_bytes()
newKc1.user_wallet_addr = random.randbytes(20)
events.append(newKc1)

newKc2 = schema_pb2.NewKeyCard()
newKc2.event_id = random.randbytes(32)
newKc2.card_public_key = public_key_from_account(kc2).to_bytes()
newKc2.user_wallet_addr = random.randbytes(20)
events.append(newKc2)

update2 = schema_pb2.UpdateManifest()
update2.event_id = random.randbytes(32)
update2.field = schema_pb2.UpdateManifest.ManifestField.MANIFEST_FIELD_DOMAIN
update2.string = "https://store.mass.market"
events.append(update2)

newTag1 = schema_pb2.CreateTag()
newTag1.event_id = random.randbytes(32)
newTag1.name = "Tag1"
events.append(newTag1)

update3 = schema_pb2.UpdateManifest()
update3.event_id = random.randbytes(32)
update3.field = schema_pb2.UpdateManifest.ManifestField.MANIFEST_FIELD_PUBLISHED_TAG
update3.tag_id = newTag1.event_id
events.append(update3)

# Add ERC20s
erc20_one = random.randbytes(20)
erc20_two = random.randbytes(20)

addErc20One = schema_pb2.UpdateManifest()
addErc20One.event_id = random.randbytes(32)
addErc20One.field = schema_pb2.UpdateManifest.ManifestField.MANIFEST_FIELD_ADD_ERC20
addErc20One.erc20_addr = erc20_one
events.append(addErc20One)

addErc20Two = schema_pb2.UpdateManifest()
addErc20Two.event_id = random.randbytes(32)
addErc20Two.field = schema_pb2.UpdateManifest.ManifestField.MANIFEST_FIELD_ADD_ERC20
addErc20Two.erc20_addr = erc20_two
events.append(addErc20Two)

rmTokenOne = schema_pb2.UpdateManifest()
rmTokenOne.event_id = random.randbytes(32)
rmTokenOne.field = schema_pb2.UpdateManifest.ManifestField.MANIFEST_FIELD_REMOVE_ERC20
rmTokenOne.erc20_addr = erc20_one
events.append(rmTokenOne)

# listing managment
newItem = schema_pb2.CreateItem()
newItem.event_id = random.randbytes(32)
newItem.price = "42.00"
newItem.metadata = b'{"title":"best schoes", "image":"https://masslbs.xyz/schoes.jpg"}'
events.append(newItem)

publishItem = schema_pb2.AddToTag()
publishItem.event_id = random.randbytes(32)
publishItem.tag_id = newTag1.event_id
publishItem.item_id = newItem.event_id
events.append(publishItem)

changePrice = schema_pb2.UpdateItem()
changePrice.event_id = random.randbytes(32)
changePrice.item_id = newItem.event_id
changePrice.field = schema_pb2.UpdateItem.ItemField.ITEM_FIELD_PRICE
changePrice.price = "23.00"
events.append(changePrice)

notPublishedItem = schema_pb2.CreateItem()
notPublishedItem.event_id = random.randbytes(32)
notPublishedItem.price = "13.12"
notPublishedItem.metadata = b'{"title":"not yet published", "image":"https://masslbs.xyz/schoes.jpg"}'
events.append(notPublishedItem)

publishItem2 = schema_pb2.AddToTag()
publishItem2.event_id = random.randbytes(32)
publishItem2.tag_id = newTag1.event_id
publishItem2.item_id = notPublishedItem.event_id
events.append(publishItem2)

unpublishItem = schema_pb2.RemoveFromTag()
unpublishItem.event_id = random.randbytes(32)
unpublishItem.tag_id = newTag1.event_id
unpublishItem.item_id = notPublishedItem.event_id
events.append(unpublishItem)

changeStock = schema_pb2.ChangeStock(item_ids = [newItem.event_id, notPublishedItem.event_id], diffs = [100, 123] )
changeStock.event_id = random.randbytes(32)
events.append(changeStock)

# open cart
cart1 = schema_pb2.CreateCart()
cart1.event_id = random.randbytes(32)
events.append(cart1)

atc1 = schema_pb2.ChangeCart()
atc1.event_id = random.randbytes(32)
atc1.cart_id = cart1.event_id
atc1.item_id = newItem.event_id
atc1.quantity = 23
events.append(atc1)

# cart to be payed
cart2 = schema_pb2.CreateCart()
cart2.event_id = random.randbytes(32)
events.append(cart2)

atc2 = schema_pb2.ChangeCart()
atc2.event_id = random.randbytes(32)
atc2.cart_id = cart2.event_id
atc2.item_id = newItem.event_id
atc2.quantity = 42
events.append(atc2)

# created by the relay.
# client would send a CommitCartRequest to finalize the cart.
commit_cart = schema_pb2.CartFinalized()
commit_cart.event_id = random.randbytes(32)
commit_cart.cart_id = cart2.event_id
commit_cart.purchase_addr = random.randbytes(20)
commit_cart.erc20_addr = random.randbytes(20)
commit_cart.sub_total = "1764.00" # 42*42
commit_cart.sales_tax = "88.20"   # 5%
commit_cart.total = "1852.20"
commit_cart.total_in_crypto = "1852.20"
events.append(commit_cart)

# would be created by the relay after payment is completed
payedCart = schema_pb2.ChangeStock(item_ids=[newItem.event_id], diffs=[-42])
payedCart.event_id = random.randbytes(32)
payedCart.cart_id = cart2.event_id
payedCart.tx_hash = random.randbytes(32)
events.append(payedCart)

# this cart will be abandoned
cart3 = schema_pb2.CreateCart()
cart3.event_id = random.randbytes(32)
events.append(cart3)

atc3 = schema_pb2.ChangeCart()
atc3.event_id = random.randbytes(32)
atc3.cart_id = cart3.event_id
atc3.item_id = newItem.event_id
atc3.quantity = 1
events.append(atc3)

commit_cart3 = schema_pb2.CartFinalized()
commit_cart3.event_id = random.randbytes(32)
commit_cart3.cart_id = cart3.event_id
commit_cart3.purchase_addr = random.randbytes(20)
commit_cart3.sub_total = "42.00"
commit_cart3.sales_tax = "2.10"
commit_cart3.total = "44.10"
commit_cart3.total_in_crypto = "xxxx"
events.append(commit_cart3)

# 24hrs pass and the sale times out
cart_abandoned = schema_pb2.CartAbandoned()
cart_abandoned.event_id = random.randbytes(32)
cart_abandoned.cart_id = cart3.event_id
events.append(cart_abandoned)

# cart 4 is in limbo. Finalized but not yet payed
cart4 = schema_pb2.CreateCart()
cart4.event_id = random.randbytes(32)
events.append(cart4)

atc3 = schema_pb2.ChangeCart()
atc3.event_id = random.randbytes(32)
atc3.cart_id = cart4.event_id
atc3.item_id = newItem.event_id
atc3.quantity = 4
events.append(atc3)

commit_cart4 = schema_pb2.CartFinalized()
commit_cart4.event_id = random.randbytes(32)
commit_cart4.cart_id = cart4.event_id
commit_cart4.purchase_addr = random.randbytes(20)
commit_cart4.sub_total = "168.00"
commit_cart4.sales_tax = "8.40"
commit_cart4.total = "176.40"
commit_cart4.total_in_crypto = "xxxx"
events.append(commit_cart4)

wrapped_events = []
for idx, evt in enumerate(events):
  debug("\nEvent: {}".format(idx))

  wrapped = None
  type_name = evt.__class__.__name__
  if type_name == "StoreManifest":
    wrapped = schema_pb2.Event(store_manifest=evt)
  elif type_name == "UpdateManifest":
    wrapped = schema_pb2.Event(update_manifest=evt)
  elif type_name == "CreateItem":
    wrapped = schema_pb2.Event(create_item=evt)
  elif type_name == "CreateTag":
    wrapped = schema_pb2.Event(create_tag=evt)
  elif type_name == "UpdateItem":
    wrapped = schema_pb2.Event(update_item=evt)
  elif type_name == "AddToTag":
    wrapped = schema_pb2.Event(add_to_tag=evt)
  elif type_name == "RemoveFromTag":
    wrapped = schema_pb2.Event(remove_from_tag=evt)
  elif type_name == "NewKeyCard":
    wrapped = schema_pb2.Event(new_key_card=evt)
  elif type_name == "ChangeStock":
    wrapped = schema_pb2.Event(change_stock=evt)
  elif type_name == "CreateCart":
    wrapped = schema_pb2.Event(create_cart=evt)
  elif type_name == "ChangeCart":
    wrapped = schema_pb2.Event(change_cart=evt)
  elif type_name == "CartFinalized":
    wrapped = schema_pb2.Event(cart_finalized=evt)
  elif type_name == "CartAbandoned":
    wrapped = schema_pb2.Event(cart_abandoned=evt)
  else:
    raise Exception(f"Unknown event type: {type_name}")

  h = hasher.hash_event(wrapped)
  msg = kc1.sign_message(h)
  wrapped.signature = msg.signature

  debug(pprint.pformat(wrapped))
  bin = wrapped.SerializeToString()
  debug(f"binary: {bin}")

  wrapped_events.append({
    "type": type_name,
    "object": protobuf_to_dict(evt),
    "signature": hex(msg.signature),
    "hash": hex(msg.messageHash),
    "encoded": binascii.hexlify(bin).decode('utf-8')
  })

output = {
  "signatures": {
    "chain_id": hasher.chain_id,
    "contract_address": hasher.storeRegAddress,
    "signer_address": kc1.address,
  },
  "events": wrapped_events,
  "reduced": {
    "manifest": {
      "store_token_id": hex(manifest.store_token_id),
      "domain": update2.string,
      "published_tag": {
        hex(newTag1.event_id): {
          "text": newTag1.name
        }
      },
      "enabled_erc20s": {
          hex(erc20_one): False,
          hex(erc20_two): True,
      },
    },
    "keycards": {
      hex(newKc1.card_public_key): hex(newKc1.user_wallet_addr),
      hex(newKc2.card_public_key): hex(newKc2.user_wallet_addr),
    },
    "items": {
      hex(newItem.event_id): {
        "price": changePrice.price,
        "metadata": newItem.metadata.decode('utf-8'),
        "tag_id" : [hex(newTag1.event_id)],
        "stock_qty": 58

      },
      hex(notPublishedItem.event_id): {
        "price": notPublishedItem.price,
        "metadata": notPublishedItem.metadata.decode('utf-8'),
        "tag_id":[],
        "stock_qty": 123
      }
    },
    # an array of items published
    "published_items": [hex(publishItem.item_id)],

    "open_carts": {
      hex(cart1.event_id): {
          hex(newItem.event_id): 23
      }
    },

    "commited_carts": {
      hex(cart4.event_id): {
          "purchase_addr": hex(commit_cart4.purchase_addr),
          "items": {
              hex(newItem.event_id): 4
          },
          "total": "176.40"
      },
    },

    "payed_carts": [{
        "cart_id": hex(cart2.event_id),
        "tx_hash": hex(payedCart.tx_hash),
    }],

    "abandoned_carts": [
        hex(cart3.event_id),
    ],

    "inventory": {
      hex(newItem.event_id): 58,
      hex(notPublishedItem.event_id): 123
    }
  }
}

with open("testVectors.json", "w") as file:
    json.dump(output, file, indent=2)
