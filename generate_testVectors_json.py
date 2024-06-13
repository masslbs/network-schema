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

from massmarket_hash_event import Hasher, shop_events_pb2
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

def unhex(a):
    return binascii.a2b_hex(a[2:])

kc1 = Account.from_key(random.randbytes(32))
debug(f"kc1: {kc1.address}")
kc2 = Account.from_key(random.randbytes(32))
debug(f"kc2: {kc2.address}")

events = []

manifest = shop_events_pb2.ShopManifest()
manifest.event_id = random.randbytes(32)
manifest.shop_token_id = random.randbytes(32)
manifest.domain = "https://masslbs.xyz"
manifest.published_tag_id = random.randbytes(32)
events.append(manifest)

newKc1 = shop_events_pb2.NewKeyCard()
newKc1.event_id = random.randbytes(32)
newKc1.card_public_key = public_key_from_account(kc1).to_bytes()
newKc1.user_wallet_addr = random.randbytes(20)
events.append(newKc1)

newKc2 = shop_events_pb2.NewKeyCard()
newKc2.event_id = random.randbytes(32)
newKc2.card_public_key = public_key_from_account(kc2).to_bytes()
newKc2.user_wallet_addr = random.randbytes(20)
events.append(newKc2)

update2 = shop_events_pb2.UpdateShopManifest()
update2.event_id = random.randbytes(32)
update2.domain = "https://shop.mass.market"
events.append(update2)

newTag1 = shop_events_pb2.CreateTag()
newTag1.event_id = random.randbytes(32)
newTag1.name = "Tag1"
events.append(newTag1)

update3 = shop_events_pb2.UpdateShopManifest()
update3.event_id = random.randbytes(32)
update3.published_tag_id = newTag1.event_id
events.append(update3)

updateMeta1 = shop_events_pb2.UpdateShopManifest()
updateMeta1.event_id = random.randbytes(32)
updateMeta1.name = "little shop of horrors"
updateMeta1.description = "Seymour and Audrey are going digital!"
events.append(updateMeta1)

updateMeta2 = shop_events_pb2.UpdateShopManifest()
updateMeta2.event_id = random.randbytes(32)
updateMeta2.profile_picture_url = "https://s3.amazonaws.com/nightjarprod/content/uploads/sites/34/2019/04/28111813/bxRm8F7rnCsVDyrNh3K9yIFFBKL.jpg"
events.append(updateMeta2)

# Add Currencies
zero_addr = bytes(20)
vanilla_eth = shop_events_pb2.UpdateShopManifest.ShopCurrency(chain=1, addr=zero_addr)
addEth = shop_events_pb2.UpdateShopManifest(add_accepted_currency=vanilla_eth)
addEth.event_id = random.randbytes(32)
events.append(addEth)

erc20_one = random.randbytes(20)
erc20_two = random.randbytes(20)

c_one = shop_events_pb2.UpdateShopManifest.ShopCurrency(chain=1, addr=erc20_one)
addErc20One = shop_events_pb2.UpdateShopManifest(add_accepted_currency=c_one)
addErc20One.event_id = random.randbytes(32)
events.append(addErc20One)

c_two = shop_events_pb2.UpdateShopManifest.ShopCurrency(chain=23, addr=erc20_two)
addErc20Two = shop_events_pb2.UpdateShopManifest(add_accepted_currency=c_two)
addErc20Two.event_id = random.randbytes(32)
events.append(addErc20Two)

rmTokenOne = shop_events_pb2.UpdateShopManifest(remove_accepted_currency=c_one)
rmTokenOne.event_id = random.randbytes(32)
events.append(rmTokenOne)

# listing managment
newItem = shop_events_pb2.CreateItem()
newItem.event_id = random.randbytes(32)
newItem.price = "42.00"
newItem.metadata = b'{"title":"best schoes", "image":"https://masslbs.xyz/schoes.jpg"}'
events.append(newItem)

publishItem = shop_events_pb2.UpdateTag()
publishItem.event_id = random.randbytes(32)
publishItem.tag_id = newTag1.event_id
publishItem.add_item_id = newItem.event_id
events.append(publishItem)

changePrice = shop_events_pb2.UpdateItem()
changePrice.event_id = random.randbytes(32)
changePrice.item_id = newItem.event_id
changePrice.price = "23.00"
events.append(changePrice)

notPublishedItem = shop_events_pb2.CreateItem()
notPublishedItem.event_id = random.randbytes(32)
notPublishedItem.price = "13.12"
notPublishedItem.metadata = b'{"title":"not yet published", "image":"https://masslbs.xyz/schoes.jpg"}'
events.append(notPublishedItem)

publishItem2 = shop_events_pb2.UpdateTag()
publishItem2.event_id = random.randbytes(32)
publishItem2.tag_id = newTag1.event_id
publishItem2.add_item_id = notPublishedItem.event_id
events.append(publishItem2)

unpublishItem = shop_events_pb2.UpdateTag()
unpublishItem.event_id = random.randbytes(32)
unpublishItem.tag_id = newTag1.event_id
unpublishItem.remove_item_id = notPublishedItem.event_id
events.append(unpublishItem)

changeStock = shop_events_pb2.ChangeStock(item_ids = [newItem.event_id, notPublishedItem.event_id], diffs = [100, 123] )
changeStock.event_id = random.randbytes(32)
events.append(changeStock)

# open order
order1 = shop_events_pb2.CreateOrder()
order1.event_id = random.randbytes(32)
events.append(order1)

change = shop_events_pb2.UpdateOrder.ChangeItems()
change.item_id = newItem.event_id
change.quantity = 23
atc1 = shop_events_pb2.UpdateOrder(change_items=change)
atc1.event_id = random.randbytes(32)
atc1.order_id = order1.event_id
events.append(atc1)

# order to be payed
order2 = shop_events_pb2.CreateOrder()
order2.event_id = random.randbytes(32)
events.append(order2)

change2 = shop_events_pb2.UpdateOrder.ChangeItems()
change2.item_id = newItem.event_id
change2.quantity = 42
atc2 = shop_events_pb2.UpdateOrder(change_items=change2)
atc2.event_id = random.randbytes(32)
atc2.order_id = order2.event_id
events.append(atc2)

# created by the relay.
# client would send a CommitOrderRequest to finalize the order.
commit_order = shop_events_pb2.UpdateOrder.ItemsFinalized()
commit_order.ttl = "1"
commit_order.is_payment_endpoint = True
commit_order.payee_addr = random.randbytes(20)
commit_order.payment_id = random.randbytes(32)
commit_order.currency_addr = random.randbytes(20)
commit_order.order_hash = random.randbytes(32)
commit_order.shop_signature = random.randbytes(64)
commit_order.sub_total = "1764.00" # 42*42
commit_order.sales_tax = "88.20"   # 5%
commit_order.total = "1852.20"
commit_order.total_in_crypto = "1852.20"
update_order = shop_events_pb2.UpdateOrder(items_finalized=commit_order)
update_order.event_id = random.randbytes(32)
update_order.order_id = order2.event_id
events.append(update_order)

# would be created by the relay after payment is completed
payedOrder = shop_events_pb2.ChangeStock(item_ids=[newItem.event_id], diffs=[-42])
payedOrder.event_id = random.randbytes(32)
payedOrder.order_id = order2.event_id
payedOrder.tx_hash = random.randbytes(32)
events.append(payedOrder)

# this order will be abandoned
order3 = shop_events_pb2.CreateOrder()
order3.event_id = random.randbytes(32)
events.append(order3)

change3 = shop_events_pb2.UpdateOrder.ChangeItems()
change3.item_id = newItem.event_id
change3.quantity = 1
atc3 = shop_events_pb2.UpdateOrder(change_items=change3)
atc3.event_id = random.randbytes(32)
atc3.order_id = order3.event_id
events.append(atc3)

commit_order3 = shop_events_pb2.UpdateOrder.ItemsFinalized()
commit_order3.ttl = "2"
commit_order3.payee_addr = unhex(kc1.address)
commit_order3.payment_id = random.randbytes(32)
commit_order3.order_hash = random.randbytes(32)
commit_order3.is_payment_endpoint = False
commit_order3.shop_signature = random.randbytes(64)
commit_order3.sub_total = "42.00"
commit_order3.sales_tax = "2.10"
commit_order3.total = "44.10"
commit_order3.total_in_crypto = "xxxx"
update_order2 = shop_events_pb2.UpdateOrder(items_finalized=commit_order3)
update_order2.event_id = random.randbytes(32)
update_order2.order_id = order3.event_id
events.append(update_order2)

# 24hrs pass and the sale times out
cancel = shop_events_pb2.UpdateOrder.OrderCanceled(timestamp=23)
update_order3 = shop_events_pb2.UpdateOrder(order_canceled=cancel)
update_order3.event_id = random.randbytes(32)
update_order3.order_id = order3.event_id
events.append(update_order3)

# order 4 is in limbo. Finalized but not yet payed
order4 = shop_events_pb2.CreateOrder()
order4.event_id = random.randbytes(32)
events.append(order4)

change4 = shop_events_pb2.UpdateOrder.ChangeItems()
change4.item_id = newItem.event_id
change4.quantity = 4
atc4 = shop_events_pb2.UpdateOrder(change_items=change4)
atc4.event_id = random.randbytes(32)
atc4.order_id = order4.event_id
events.append(atc4)

commit_order4 = shop_events_pb2.UpdateOrder.ItemsFinalized()
commit_order4.ttl = "3"
commit_order4.payee_addr = unhex(kc1.address)
commit_order4.is_payment_endpoint = False
commit_order4.payment_id = random.randbytes(32)
commit_order4.order_hash = random.randbytes(32)
commit_order4.shop_signature = random.randbytes(64)
commit_order4.sub_total = "168.00"
commit_order4.sales_tax = "8.40"
commit_order4.total = "176.40"
commit_order4.total_in_crypto = "xxxx"
update_order3 = shop_events_pb2.UpdateOrder(items_finalized=commit_order4)
update_order3.event_id = random.randbytes(32)
update_order3.order_id = order4.event_id
events.append(update_order3)

wrapped_events = []
for idx, evt in enumerate(events):
  type_name = evt.__class__.__name__

  debug(f"\nEvent idx={idx} type={type_name}\n{evt}")

  wrapped = None
  if type_name == "ShopManifest":
    wrapped = shop_events_pb2.ShopEvent(shop_manifest=evt)
  elif type_name == "UpdateShopManifest":
    wrapped = shop_events_pb2.ShopEvent(update_shop_manifest=evt)
  elif type_name == "CreateItem":
    wrapped = shop_events_pb2.ShopEvent(create_item=evt)
  elif type_name == "UpdateItem":
    wrapped = shop_events_pb2.ShopEvent(update_item=evt)
  elif type_name == "CreateTag":
    wrapped = shop_events_pb2.ShopEvent(create_tag=evt)
  elif type_name == "UpdateTag":
    wrapped = shop_events_pb2.ShopEvent(update_tag=evt)
  elif type_name == "NewKeyCard":
    wrapped = shop_events_pb2.ShopEvent(new_key_card=evt)
  elif type_name == "ChangeStock":
    wrapped = shop_events_pb2.ShopEvent(change_stock=evt)
  elif type_name == "CreateOrder":
    wrapped = shop_events_pb2.ShopEvent(create_order=evt)
  elif type_name == "UpdateOrder":
    wrapped = shop_events_pb2.ShopEvent(update_order=evt)
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
    "contract_address": hasher.shopRegAddress,
    "signer_address": kc1.address,
  },
  "events": wrapped_events,
  "reduced": {
    "manifest": {
      "shop_token_id": hex(manifest.shop_token_id),
      "name": updateMeta1.name,
      "description": updateMeta1.description,
      "profile_picture_url": updateMeta2.profile_picture_url,
      "domain": update2.domain,
      "published_tag": {
        hex(newTag1.event_id): {
          "text": newTag1.name
        }
      },
      "accepted_currencies": [
          {"chain": 1, "addr": hex(zero_addr)},
          {"chain": 23, "addr": hex(erc20_two)},
      ],
      "base_currency": {"chain": 1, "addr": hex(zero_addr)},
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
    "published_items": [hex(publishItem.add_item_id)],

    "open_orders": {
      hex(order1.event_id): {
          hex(newItem.event_id): 23
      }
    },

    "commited_orders": {
      hex(order4.event_id): {
          "payment_id": hex(commit_order4.payment_id),
          "items": {
              hex(newItem.event_id): 4
          },
          "total": "176.40"
      },
    },

    "payed_orders": [{
        "order_id": hex(order2.event_id),
        "tx_hash": hex(payedOrder.tx_hash),
    }],

    "abandoned_orders": [
        hex(order3.event_id),
    ],

    "inventory": {
      hex(newItem.event_id): 58,
      hex(notPublishedItem.event_id): 123
    }
  }
}

with open("testVectors.json", "w") as file:
    json.dump(output, file, indent=2)
