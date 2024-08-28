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
import math

max_uint64 = int(math.pow(2, 64) - 1)


def rand_uint64():
    return random.randint(0, max_uint64)


from google.protobuf import timestamp_pb2
from web3 import Account
from eth_keys import keys
from protobuf_to_dict import protobuf_to_dict

from massmarket_hash_event import (
    hash_event,
    base_types_pb2 as mtypes,
    shop_events_pb2 as mevents,
)


def public_key_from_account(account):
    k = keys.PrivateKey(account.key)
    return k.public_key


def debug(message):
    if os.getenv("DEBUG") != None:
        sys.stderr.write(message + "\n")


def hex(b):
    return "0x" + binascii.hexlify(b).decode("utf-8")


def unhex(a):
    if a.startswith("0x"):
        a = a[2:]
    return binascii.a2b_hex(a)


shop_id = mtypes.Uint256(raw=random.randbytes(32))

user1Addr = mtypes.EthereumAddress(raw=random.randbytes(20))
kc1 = Account.from_key(random.randbytes(32))
debug(f"kc1: {kc1.address}")
user2Addr = mtypes.EthereumAddress(raw=random.randbytes(20))
kc2 = Account.from_key(random.randbytes(32))
debug(f"kc2: {kc2.address}")
guestKeyPair = Account.from_key(random.randbytes(32))

events = []

##############
## Manifest ##
##############

payee23 = mtypes.Payee(
    name="L23",
    address=user1Addr,
    chain_id=23,
)
manifest = mevents.Manifest(token_id=shop_id, payees=[payee23])
events.append(manifest)
##############
## Accounts ##
##############

newKc1 = mevents.Account(
    enroll_keycard=mevents.Account.KeyCardEnroll(
        user_wallet=user1Addr,
        keycard_pubkey=mtypes.PublicKey(raw=public_key_from_account(kc1).to_bytes()),
    )
)
events.append(newKc1)

newKc2 = mevents.Account(
    enroll_keycard=mevents.Account.KeyCardEnroll(
        user_wallet=user2Addr,
        keycard_pubkey=mtypes.PublicKey(raw=public_key_from_account(kc2).to_bytes()),
    )
)
events.append(newKc2)

zero_addr = mtypes.EthereumAddress(raw=bytes(20))
guestKc = mevents.Account(
    enroll_keycard=mevents.Account.KeyCardEnroll(
        user_wallet=zero_addr,
        keycard_pubkey=mtypes.PublicKey(
            raw=public_key_from_account(guestKeyPair).to_bytes()
        ),
    )
)
events.append(guestKc)

################
## Currencies ##
################

payee = mtypes.Payee(
    name="default",
    address=user1Addr,
    chain_id=1,
)
vanilla_eth = mtypes.ShopCurrency(
    chain_id=1,
    address=zero_addr,
)
addEth = mevents.UpdateManifest(
    add_payee=payee,
    add_accepted_currencies=[vanilla_eth],
    set_base_currency=vanilla_eth,
)
events.append(addEth)

erc20_one = mtypes.EthereumAddress(raw=random.randbytes(20))
erc20_two = mtypes.EthereumAddress(raw=random.randbytes(20))

c_one = mtypes.ShopCurrency(
    chain_id=1,
    address=erc20_one,
)
addErc20One = mevents.UpdateManifest(
    add_accepted_currencies=[c_one],
)
events.append(addErc20One)

c_two = mtypes.ShopCurrency(
    chain_id=23,
    address=erc20_two,
)
addErc20Two = mevents.UpdateManifest(
    add_accepted_currencies=[c_two],
    remove_accepted_currencies=[c_one],
)
events.append(addErc20Two)


##########
## Tags ##
##########


tag_stuff = mevents.Tag(
    id=rand_uint64(),
    name="Stuff",
)
events.append(tag_stuff)

tag_clothes = mevents.Tag(
    id=rand_uint64(),
    name="Clothes",
)
events.append(tag_clothes)

#######################
## listing managment ##
#######################

# no options
listing_simple = mevents.Listing(
    id=rand_uint64(),
    base_price=mtypes.Uint256(raw=random.randbytes(32)),
    base_info=mtypes.ListingMetadata(
        title="the pen",
        description="great pen",
        images=["https://masslbs.xyz/pen.jpg"],
    ),
)
events.append(listing_simple)

sort_listing_simple = mevents.UpdateTag(
    id=tag_stuff.id,
    add_listing_ids=[listing_simple.id],
)
events.append(sort_listing_simple)

change_price = mevents.UpdateListing(
    id=listing_simple.id,
    base_price=mtypes.Uint256(raw=int(123400).to_bytes(32, "big")),
)
events.append(change_price)
listing_simple.base_price.CopyFrom(change_price.base_price)

change_inventory = mevents.ChangeInventory(
    id=listing_simple.id,
    diff=100,
)
events.append(change_inventory)

stock_status = mtypes.ListingStockStatus(in_stock=True)
publish_simple = mevents.UpdateListing(
    id=listing_simple.id,
    view_state=mtypes.ListingViewState.LISTING_VIEW_STATE_PUBLISHED,
    stock_updates=[stock_status],
)
events.append(publish_simple)

# one option
# ==========
l1_small = rand_uint64()
l1_medium = rand_uint64()
l1_large = rand_uint64()

listing_w_sizes = mevents.Listing(
    id=rand_uint64(),
    base_price=mtypes.Uint256(raw=int(500).to_bytes(32, "big")),
    base_info=mtypes.ListingMetadata(
        title="The Painting (print)",
        description="Beautiful, in all sizes",
        images=["https://masslbs.xyz/painting.jpg"],
    ),
    view_state=mtypes.ListingViewState.LISTING_VIEW_STATE_PUBLISHED,
    options=[
        mtypes.ListingOption(
            id=rand_uint64(),
            title="Size",
            variations=[
                mtypes.ListingVariation(
                    id=l1_small,
                    variation_info=mtypes.ListingMetadata(
                        title="Small",
                        description="400x300",
                    ),
                ),
                mtypes.ListingVariation(
                    id=l1_medium,
                    variation_info=mtypes.ListingMetadata(
                        title="Medium",
                        description="600x450",
                    ),
                    price_diff_sign=True,
                    price_diff=mtypes.Uint256(raw=int(200).to_bytes(32, "big")),
                ),
                mtypes.ListingVariation(
                    id=l1_large,
                    variation_info=mtypes.ListingMetadata(
                        title="Large",
                        description="800x600",
                    ),
                    price_diff_sign=True,
                    price_diff=mtypes.Uint256(raw=int(400).to_bytes(32, "big")),
                ),
            ],
        )
    ],
    stock_statuses=[
        mtypes.ListingStockStatus(
            variation_ids=[l1_small],
            in_stock=True,
        ),
        mtypes.ListingStockStatus(
            variation_ids=[l1_medium],
            in_stock=True,
        ),
        mtypes.ListingStockStatus(
            variation_ids=[l1_large],
            expected_in_stock_by=timestamp_pb2.Timestamp(seconds=9999999),
        ),
    ],
)
events.append(listing_w_sizes)

# update stock/inventory of individual variations
change_inventory2_small = mevents.ChangeInventory(
    id=listing_w_sizes.id,
    variation_ids=[l1_small],
    diff=10,
)
events.append(change_inventory2_small)
change_inventory2_medium = mevents.ChangeInventory(
    id=listing_w_sizes.id,
    variation_ids=[l1_medium],
    diff=20,
)
events.append(change_inventory2_medium)

# two options
# ===========
l2_opt_size = rand_uint64()

l2_size10 = rand_uint64()
l2_size11 = rand_uint64()
l2_size12 = rand_uint64()

l2_color_red = rand_uint64()
l2_color_green = rand_uint64()
l2_color_blue = rand_uint64()

listing_color_and_size = mevents.Listing(
    id=rand_uint64(),
    base_price=mtypes.Uint256(raw=int(10000).to_bytes(32, "big")),
    options=[
        mtypes.ListingOption(
            id=l2_opt_size,
            title="Size",
            variations=[
                mtypes.ListingVariation(
                    id=l2_size10,
                    variation_info=mtypes.ListingMetadata(
                        title="10 (US)",
                        description="US 10 equals EU xyz",
                    ),
                ),
                mtypes.ListingVariation(
                    id=l2_size11,
                    variation_info=mtypes.ListingMetadata(
                        title="11",
                    ),
                ),
                mtypes.ListingVariation(
                    id=l2_size12,
                    variation_info=mtypes.ListingMetadata(
                        title="12",
                    ),
                ),
            ],
        ),
        mtypes.ListingOption(
            id=rand_uint64(),
            title="Color",
            variations=[
                mtypes.ListingVariation(
                    id=l2_color_red,
                    variation_info=mtypes.ListingMetadata(
                        title="Red",
                    ),
                ),
                mtypes.ListingVariation(
                    id=l2_color_green,
                    variation_info=mtypes.ListingMetadata(
                        title="Green",
                    ),
                ),
                mtypes.ListingVariation(
                    id=l2_color_blue,
                    variation_info=mtypes.ListingMetadata(
                        title="Blue",
                    ),
                ),
            ],
        ),
    ],
    stock_statuses=[
        mtypes.ListingStockStatus(
            variation_ids=[l2_size10, l2_color_red],
            in_stock=True,
        ),
        mtypes.ListingStockStatus(
            variation_ids=[l2_size11, l2_color_red],
            in_stock=True,
        ),
        mtypes.ListingStockStatus(
            variation_ids=[l2_size12, l2_color_red],
            in_stock=False,
        ),
        mtypes.ListingStockStatus(
            variation_ids=[l2_size10, l2_color_green],
            in_stock=False,
        ),
        mtypes.ListingStockStatus(
            variation_ids=[l2_size10, l2_color_blue],
            in_stock=True,
        ),
        mtypes.ListingStockStatus(
            variation_ids=[l2_size11, l2_color_blue],
            in_stock=False,
        ),
        mtypes.ListingStockStatus(
            variation_ids=[l2_size12, l2_color_blue],
            in_stock=True,
        ),
    ],
)
events.append(listing_color_and_size)

# add a variation to an option
l2_size13 = rand_uint64()
add_size_var = mevents.UpdateListing.AddVariation(
    option_id=l2_opt_size,
    variation=mtypes.ListingVariation(
        id=l2_size13, variation_info=mtypes.ListingMetadata(title="13")
    ),
)
add_size_evt = mevents.UpdateListing(
    id=listing_color_and_size.id,
    add_variations=[add_size_var],
)
events.append(add_size_evt)

update_stock_evt = mevents.UpdateListing(
    id=listing_color_and_size.id,
    stock_updates=[
        mtypes.ListingStockStatus(
            variation_ids=[l2_size13, l2_color_blue],
            in_stock=True,
        ),
    ],
)
events.append(update_stock_evt)

# remove a variation
rm_combo_evt = mevents.UpdateListing(
    id=listing_color_and_size.id,
    remove_variations=[
        l2_color_red,
    ],
)
events.append(add_size_evt)

# manually apply update to listing
current = listing_color_and_size.options[1]
listing_color_and_size.options[1].CopyFrom(
    mtypes.ListingOption(
        id=current.id,
        variations=[v for v in current.variations if v.id != l2_color_red],
    )
)
current = listing_color_and_size.stock_statuses
for found in [ss for ss in current if l2_color_red in ss.variation_ids]:
    listing_color_and_size.stock_statuses.remove(found)

# update stock/inventory of individual variations
change_inventory3_10green = mevents.ChangeInventory(
    id=listing_color_and_size.id,
    variation_ids=[l2_size10, l2_color_green],
    diff=127,
)
events.append(change_inventory3_10green)
change_inventory2_13blue = mevents.ChangeInventory(
    id=listing_color_and_size.id,
    variation_ids=[l2_size13, l2_color_blue],
    diff=171,
)
events.append(change_inventory2_13blue)


############
## Orders ##
############

addr = mtypes.AddressDetails()
addr.name = "Max Mustermann"
addr.address1 = "Somestreet 1"
addr.city = "City"
addr.postal_code = "12345"
addr.country = "Isla de Muerta"
addr.phone_number = "+0155512345"
addr.email_address = "some1@no.where"

# open
# ====
order_open = mevents.CreateOrder(id=rand_uint64())
events.append(order_open)

order_open_item = mtypes.OrderedItem(
    listing_id=listing_simple.id,
    quantity=23,
)
add_to_order_open = mevents.UpdateOrder(
    id=order_open.id,
    change_items=mevents.UpdateOrder.ChangeItems(
        adds=[order_open_item],
    ),
)
events.append(add_to_order_open)

# paid
# ====
order_paid = mevents.CreateOrder(id=rand_uint64())
events.append(order_paid)

order_paid_item = mtypes.OrderedItem(
    listing_id=listing_simple.id,
    quantity=42,
)
add_to_order_paid = mevents.UpdateOrder(
    id=order_paid.id,
    change_items=mevents.UpdateOrder.ChangeItems(
        adds=[order_paid_item],
    ),
)
events.append(add_to_order_paid)

order_paid_add_addr = mevents.UpdateOrder(
    id=order_paid.id,
    invoice_address=addr,
)
events.append(order_paid_add_addr)


commit_order_paid = mevents.UpdateOrder(
    id=order_paid.id,
    commit_items=mevents.UpdateOrder.CommitItems(),
)
events.append(commit_order_paid)

choose_payment = mevents.UpdateOrder(
    id=order_paid.id,
    choose_payment=mevents.UpdateOrder.ChoosePaymentMethod(
        currency=c_two,
        payee=payee,
    ),
)
events.append(choose_payment)

# created by the relay on receiving the commit
listing_simple_hash = mtypes.IPFSAddress(cid="/ipfs/foobar")
order_open_payment_details = mtypes.PaymentDetails(
    ttl="1",
    total=mtypes.Uint256(
        raw=int(123400 * order_paid_item.quantity).to_bytes(32, "big"),
    ),
    # TODO: hash the actual value
    payment_id=mtypes.Hash(raw=random.randbytes(32)),
    listing_hashes=[listing_simple_hash],
    shop_signature=mtypes.Signature(raw=random.randbytes(64)),
)
finalized_order = mevents.UpdateOrder(
    id=order_paid.id,
    payment_details=order_open_payment_details,
)
events.append(finalized_order)

# would be created by the relay after payment is completed
paid_stock_change = mevents.ChangeInventory(
    id=listing_simple.id,
    diff=-42,
)
events.append(paid_stock_change)

order_is_paid = mevents.UpdateOrder(
    id=order_paid.id,
    paid=mtypes.OrderPaid(
        tx_hash=mtypes.Hash(raw=random.randbytes(32)),
    ),
)
events.append(order_is_paid)

# this will be canceled
# =====================
order_canceled = mevents.CreateOrder(
    id=rand_uint64(),
)
events.append(order_canceled)

order_canceled_item = mtypes.OrderedItem(
    listing_id=listing_w_sizes.id,
    variation_ids=[l1_medium],
    quantity=2,
)

add_to_order_canceled = mevents.UpdateOrder(
    id=order_canceled.id,
    change_items=mevents.UpdateOrder.ChangeItems(
        adds=[order_canceled_item],
    ),
)
events.append(add_to_order_canceled)

order_canceled_add_addr = mevents.UpdateOrder(
    id=order_canceled.id,
    invoice_address=addr,
)
events.append(order_canceled_add_addr)

commit_order_canceled = mevents.UpdateOrder(
    id=order_canceled.id,
    commit_items=mevents.UpdateOrder.CommitItems(),
)
events.append(commit_order_canceled)

chose_payment_order_canceled = mevents.UpdateOrder(
    id=order_canceled.id,
    choose_payment=mevents.UpdateOrder.ChoosePaymentMethod(
        currency=c_two,
        payee=payee,
    ),
)
events.append(chose_payment_order_canceled)


payment_details3 = mtypes.PaymentDetails(
    payment_id=mtypes.Hash(raw=random.randbytes(32)),
    ttl="2",
    shop_signature=mtypes.Signature(raw=random.randbytes(65)),
    total=mtypes.Uint256(raw=int(1400).to_bytes(32, "big")),
    listing_hashes=[listing_simple_hash],
)
update_order_paid = mevents.UpdateOrder(
    id=order_canceled.id,
    payment_details=payment_details3,
)
events.append(update_order_paid)

# 24hrs pass and the sale times out
cancel = mevents.UpdateOrder.Canceled(
    canceld_at=timestamp_pb2.Timestamp(seconds=23),
)
update_order_canceled = mevents.UpdateOrder(
    id=order_canceled.id,
    canceled=cancel,
)
events.append(update_order_canceled)

# order 4 is in limbo.
# Finalized but not yet payed
# ===========================
order4 = mevents.CreateOrder(
    id=rand_uint64(),
)
events.append(order4)

change4_1 = mtypes.OrderedItem(
    listing_id=listing_color_and_size.id,
    variation_ids=[l2_size10, l2_color_green],
    quantity=4,
)
change4_2 = mtypes.OrderedItem(
    listing_id=listing_color_and_size.id,
    variation_ids=[l2_size13, l2_color_blue],
    quantity=10,
)
add_to_order4 = mevents.UpdateOrder(
    id=order4.id,
    change_items=mevents.UpdateOrder.ChangeItems(
        adds=[change4_1, change4_2],
    ),
)
events.append(add_to_order4)

commit_order4 = mevents.UpdateOrder(
    id=order4.id,
    commit_items=mevents.UpdateOrder.CommitItems(),
)
events.append(commit_order4)

payment_order4 = mevents.UpdateOrder(
    id=order4.id,
    choose_payment=mevents.UpdateOrder.ChoosePaymentMethod(
        currency=c_two,
        payee=payee,
    ),
)
events.append(payment_order4)


order4_add_addr = mevents.UpdateOrder(
    id=order4.id,
    invoice_address=addr,
)
events.append(order4_add_addr)

commit_order4 = mtypes.PaymentDetails(
    payment_id=mtypes.Hash(raw=random.randbytes(32)),
    ttl="3",
    shop_signature=mtypes.Signature(raw=random.randbytes(65)),
    total=mtypes.Uint256(raw=int(16800).to_bytes(32, "big")),
    listing_hashes=[listing_simple_hash],
)
update_order4 = mevents.UpdateOrder(
    id=order4.id,
    payment_details=commit_order4,
)
events.append(update_order4)


wrapped_events = []
for idx, evt in enumerate(events):
    type_name = evt.__class__.__name__

    debug(f"\nEvent idx={idx} type={type_name}\n{evt}")

    wrapped = None

    if type_name == "Manifest":
        wrapped = mevents.ShopEvent(manifest=evt)
    elif type_name == "UpdateManifest":
        wrapped = mevents.ShopEvent(update_manifest=evt)
    elif type_name == "Account":
        wrapped = mevents.ShopEvent(account=evt)

    elif type_name == "Listing":
        wrapped = mevents.ShopEvent(listing=evt)
    elif type_name == "UpdateListing":
        wrapped = mevents.ShopEvent(update_listing=evt)

    elif type_name == "ChangeInventory":
        wrapped = mevents.ShopEvent(change_inventory=evt)

    elif type_name == "Tag":
        wrapped = mevents.ShopEvent(tag=evt)
    elif type_name == "UpdateTag":
        wrapped = mevents.ShopEvent(update_tag=evt)

    elif type_name == "CreateOrder":
        wrapped = mevents.ShopEvent(create_order=evt)
    elif type_name == "UpdateOrder":
        wrapped = mevents.ShopEvent(update_order=evt)

    else:
        raise Exception(f"Unknown event[{idx}] type: {type_name}")

    assert wrapped is not None
    wrapped.shop_id.CopyFrom(shop_id)
    n = len(wrapped_events)
    wrapped.nonce = n
    wrapped.timestamp.CopyFrom(timestamp_pb2.Timestamp(seconds=3600 * n))

    h = hash_event(wrapped)
    sig = kc1.sign_message(h)
    # pprint.pprint(msg)

    debug(pprint.pformat(wrapped))
    bin = wrapped.SerializeToString()
    debug(f"binary: {bin}")

    obj_dict = protobuf_to_dict(evt)
    # pprint.pp(obj_dict)
    wrapped_events.append(
        {
            "type": type_name,
            "object": obj_dict,
            "signature": hex(sig.signature),
            "hash": hex(sig.messageHash),
            "encoded": hex(bin),
        }
    )


# stitch together current state
# TODO: implement actual reducer logic

current_manifest = mevents.Manifest(
    token_id=shop_id,
    payees=[
        payee,
        payee23,
    ],
    accepted_currencies=[
        vanilla_eth,
        c_two,
    ],
    base_currency=vanilla_eth,
)

current_order_open = mevents.Order(
    id=order_open.id,
    state=mevents.Order.State.STATE_OPEN,
    items=[order_open_item],
    invoice_address=addr,
)

current_order_paid = mevents.Order(
    id=order_paid.id,
    state=mevents.Order.State.STATE_PAID,
    items=[order_paid_item],
    invoice_address=addr,
    chosen_payee=payee,
    chosen_currency=c_two,
    payment_details=order_open_payment_details,
)

current_order_canceled = mevents.Order(
    id=order_canceled.id,
    state=mevents.Order.State.STATE_CANCELED,
    canceled_at=cancel.canceld_at,
    items=[order_canceled_item],
    invoice_address=addr,
    chosen_payee=payee,
    chosen_currency=c_two,
    payment_details=payment_details3,
)

current_order_unpaid = mevents.Order(
    id=order4.id,
    state=mevents.Order.State.STATE_UNPAID,
    items=[change4_1, change4_2],
    invoice_address=addr,
    chosen_payee=payee,
    chosen_currency=c_two,
    payment_details=commit_order4,
)

# construct json output object

output = {
    "signatures": {
        "shop_id": hex(shop_id.raw),
        "signer_address": kc1.address,
    },
    "events": wrapped_events,
    "reduced": {
        "manifest": protobuf_to_dict(current_manifest),
        "keycards": [
            kc1.address,
            kc2.address,
            guestKeyPair.address,
        ],
        "listings": [
            protobuf_to_dict(listing_simple),
            protobuf_to_dict(listing_w_sizes),
            protobuf_to_dict(listing_color_and_size),
        ],
        "tags": {
            tag_stuff.id: {
                "name": tag_stuff.name,
                "item_ids": [listing_simple.id],
            },
            tag_clothes.id: {
                "name": tag_clothes.name,
                "item_ids": [],
            },
        },
        "inventory": [
            {
                "listing_id": listing_simple.id,
                "variations": [],
                "quantity": 58,
            },
            {
                "listing_id": listing_w_sizes.id,
                "variations": [l1_small],
                "quantity": 10,
            },
            {
                "listing_id": listing_w_sizes.id,
                "variations": [l1_medium],
                "quantity": 20,
            },
            {
                "listing_id": listing_color_and_size.id,
                "variations": [l2_size10, l2_color_green],
                "quantity": 123,
            },
            {
                "listing_id": listing_color_and_size.id,
                "variations": [l2_size13, l2_color_blue],
                "quantity": 161,
            },
        ],
        "orders": [
            protobuf_to_dict(current_order_open),
            protobuf_to_dict(current_order_paid),
            protobuf_to_dict(current_order_canceled),
            protobuf_to_dict(current_order_unpaid),
        ],
    },
}

# fix rendering timestamps
from datetime import datetime, timezone
from json import JSONEncoder


class DateTimeEncoder(JSONEncoder):
    def default(self, obj):
        if isinstance(obj, datetime):
            utc = obj.astimezone(timezone.utc)
            return utc.isoformat()
        return super().default(obj)


with open("testVectors.json", "w") as file:
    json.dump(output, file, indent=2, cls=DateTimeEncoder)
