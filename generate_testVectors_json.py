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

from google.protobuf import timestamp_pb2
from web3 import Account
from eth_keys import keys
from protobuf_to_dict import protobuf_to_dict

from massmarket_hash_event import (
    hash_event,
    base_types_pb2 as mtypes,
    shop_events_pb2 as mevents,
    storage_pb2 as mstorage,
)

from datetime import datetime, timezone
from json import JSONEncoder


def rand_obj_id():
    return mtypes.ObjectId(raw=random.randbytes(8))


def random_ethereum_address():
    return mtypes.EthereumAddress(raw=random.randbytes(20))


def random_hash():
    return mtypes.Hash(raw=random.randbytes(32))


def new_uint256(value: int) -> mtypes.Uint256:
    return mtypes.Uint256(raw=value.to_bytes(32, "big"))


def public_key_from_account(account):
    k = keys.PrivateKey(account.key)
    return k.public_key


def debug(message):
    if os.getenv("DEBUG") != None:
        sys.stderr.write(message + "\n")


def hex(b):
    return "0x" + binascii.hexlify(b).decode("utf-8")


# TODO: might want to patch pb_to_dict to use this, too
def id_to_hex(i: mtypes.ObjectId):
    return hex(i.raw)


def unhex(a):
    if a.startswith("0x"):
        a = a[2:]
    return binascii.a2b_hex(a)


class HexEncoder(JSONEncoder):
    def default(self, obj):
        # fix rendering timestamps in json output
        if isinstance(obj, datetime):
            utc = obj.astimezone(timezone.utc)
            return utc.isoformat()
        if isinstance(obj, mtypes.ObjectId):
            return id_to_hex(obj)
        return super().default(obj)


shop_id = mtypes.Uint256(raw=random.randbytes(32))

user1Addr = random_ethereum_address()
kc1 = Account.from_key(random.randbytes(32))
debug(f"kc1: {kc1.address}")
user2Addr = random_ethereum_address()
kc2 = Account.from_key(random.randbytes(32))
debug(f"kc2: {kc2.address}")
guestKeyPair = Account.from_key(random.randbytes(32))

zero_addr = mtypes.EthereumAddress(raw=bytes(20))
erc20_one = random_ethereum_address()
erc20_two = random_ethereum_address()

events = []


def append_event(event):
    events.append(event)
    return event


##############
## Manifest ##
##############

mod_eu_vat = mtypes.OrderPriceModifier(
    title="EU VAT",
    percentage=new_uint256(19),
)
mod_dhl_local = mtypes.OrderPriceModifier(
    title="DHL Local",
    absolute=mtypes.PlusMinus(
        plus_sign=True,
        # TODO: assuming 2 decimals for now
        diff=new_uint256(500),
    ),
)

mod_dhl_international = mtypes.OrderPriceModifier(
    title="DHL International",
    absolute=mtypes.PlusMinus(
        plus_sign=True,
        # TODO: assuming 2 decimals for now
        diff=new_uint256(4200),
    ),
)

region_local = mtypes.ShippingRegion(
    name="domestic",
    country="germany",
    order_price_modifiers=[mod_eu_vat, mod_dhl_local],
)

region_other = mtypes.ShippingRegion(
    name="other",
    order_price_modifiers=[mod_dhl_international],
)

payee23 = mtypes.Payee(
    name="L23",
    address=user1Addr,
    chain_id=23,
)

manifest = mevents.Manifest(
    token_id=shop_id,
    payees=[payee23],
    shipping_regions=[region_local, region_other],
)
append_event(manifest)

remove_payee = mevents.UpdateManifest(
    remove_payee=payee23,
)
append_event(remove_payee)


region_test = mtypes.ShippingRegion(
    name="test",
)
add_region_test = mevents.UpdateManifest(
    add_shipping_regions=[region_test],
)
append_event(add_region_test)

remove_shipping_region = mevents.UpdateManifest(
    remove_shipping_regions=[region_test.name],
)
append_event(remove_shipping_region)


##############
## Accounts ##
##############

newKc1 = mevents.Account(
    enroll_keycard=mevents.Account.KeyCardEnroll(
        user_wallet=user1Addr,
        keycard_pubkey=mtypes.PublicKey(raw=public_key_from_account(kc1).to_bytes()),
    )
)
append_event(newKc1)

newKc2 = mevents.Account(
    enroll_keycard=mevents.Account.KeyCardEnroll(
        user_wallet=user2Addr,
        keycard_pubkey=mtypes.PublicKey(raw=public_key_from_account(kc2).to_bytes()),
    )
)
append_event(newKc2)

guestKc = mevents.Account(
    enroll_keycard=mevents.Account.KeyCardEnroll(
        user_wallet=zero_addr,
        keycard_pubkey=mtypes.PublicKey(
            raw=public_key_from_account(guestKeyPair).to_bytes()
        ),
    )
)
append_event(guestKc)

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
    set_pricing_currency=vanilla_eth,
)
append_event(addEth)

c_one = mtypes.ShopCurrency(
    chain_id=1,
    address=erc20_one,
)
addErc20One = mevents.UpdateManifest(
    add_accepted_currencies=[c_one],
)
append_event(addErc20One)

c_two = mtypes.ShopCurrency(
    chain_id=23,
    address=erc20_two,
)
addErc20Two = mevents.UpdateManifest(
    add_accepted_currencies=[c_two],
    remove_accepted_currencies=[c_one],
)
append_event(addErc20Two)


##########
## Tags ##
##########


tag_stuff = mevents.Tag(
    id=rand_obj_id(),
    name="Stuff",
)
append_event(tag_stuff)

update_tag = mevents.UpdateTag(
    id=tag_stuff.id,
    rename="Better Stuff",
)
append_event(update_tag)

tag_clothes = mevents.Tag(
    id=rand_obj_id(),
    name="Clothes",
)
append_event(tag_clothes)

#######################
## listing managment ##
#######################

# no options
listing_simple = mevents.Listing(
    id=rand_obj_id(),
    price=new_uint256(100),
    metadata=mtypes.ListingMetadata(
        title="the pen",
        description="great pen",
        images=["https://masslbs.xyz/pen.jpg"],
    ),
)
append_event(listing_simple)

sort_listing_simple = mevents.UpdateTag(
    id=tag_stuff.id,
    add_listing_ids=[listing_simple.id],
)

# and and remove from tag
err_add_tag = mevents.UpdateTag(
    id=tag_clothes.id,
    add_listing_ids=[listing_simple.id],
)
append_event(err_add_tag)

err_remove_tag = mevents.UpdateTag(
    id=tag_clothes.id,
    remove_listing_ids=[listing_simple.id],
)
append_event(err_remove_tag)

change_price = mevents.UpdateListing(
    id=listing_simple.id,
    price=new_uint256(123400),
)
append_event(change_price)
listing_simple.price.CopyFrom(change_price.price)

change_inventory = mevents.ChangeInventory(
    id=listing_simple.id,
    diff=100,
)
append_event(change_inventory)

stock_status = mtypes.ListingStockStatus(in_stock=True)
publish_simple = mevents.UpdateListing(
    id=listing_simple.id,
    view_state=mtypes.ListingViewState.LISTING_VIEW_STATE_PUBLISHED,
    stock_updates=[stock_status],
)
append_event(publish_simple)

update_listing_metadata = mevents.UpdateListing(
    id=listing_simple.id,
    metadata=mtypes.ListingMetadata(
        title="Updated Pen",
        description="Even better pen now!",
        images=["https://masslbs.xyz/updated_pen.jpg"],
    ),
)
append_event(update_listing_metadata)
listing_simple.metadata.CopyFrom(update_listing_metadata.metadata)

# one option
# ==========
l1_small = rand_obj_id()
l1_medium = rand_obj_id()
l1_large = rand_obj_id()

listing_w_sizes = mevents.Listing(
    id=rand_obj_id(),
    price=new_uint256(500),
    metadata=mtypes.ListingMetadata(
        title="The Painting (print)",
        description="Beautiful, in all sizes",
        images=["https://masslbs.xyz/painting.jpg"],
    ),
    view_state=mtypes.ListingViewState.LISTING_VIEW_STATE_PUBLISHED,
    options=[
        mtypes.ListingOption(
            id=rand_obj_id(),
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
                    diff=mtypes.PlusMinus(
                        plus_sign=True,
                        diff=new_uint256(200),
                    ),
                ),
                mtypes.ListingVariation(
                    id=l1_large,
                    variation_info=mtypes.ListingMetadata(
                        title="Large",
                        description="800x600",
                    ),
                    diff=mtypes.PlusMinus(
                        plus_sign=True,
                        diff=new_uint256(400),
                    ),
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
append_event(listing_w_sizes)

# update stock/inventory of individual variations
change_inventory2_small = mevents.ChangeInventory(
    id=listing_w_sizes.id,
    variation_ids=[l1_small],
    diff=10,
)
append_event(change_inventory2_small)
change_inventory2_medium = mevents.ChangeInventory(
    id=listing_w_sizes.id,
    variation_ids=[l1_medium],
    diff=20,
)
append_event(change_inventory2_medium)

# two options
# ===========
l2_opt_size = rand_obj_id()

l2_size10 = rand_obj_id()
l2_size11 = rand_obj_id()
l2_size12 = rand_obj_id()

l2_color_red = rand_obj_id()
l2_color_green = rand_obj_id()
l2_color_blue = rand_obj_id()

listing_color_and_size = mevents.Listing(
    id=rand_obj_id(),
    price=new_uint256(10000),
    metadata=mtypes.ListingMetadata(
        title="The Shoes",
        description="Beautiful, in all sizes",
        images=["https://masslbs.xyz/shoes.jpg"],
    ),
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
            id=rand_obj_id(),
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
append_event(listing_color_and_size)

# add a variation to an option
l2_size13 = rand_obj_id()
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
append_event(add_size_evt)

update_stock_evt = mevents.UpdateListing(
    id=listing_color_and_size.id,
    stock_updates=[
        mtypes.ListingStockStatus(
            variation_ids=[l2_size13, l2_color_blue],
            in_stock=True,
        ),
    ],
)
append_event(update_stock_evt)

# remove a variation
rm_combo_evt = mevents.UpdateListing(
    id=listing_color_and_size.id,
    remove_variation_ids=[
        l2_color_red,
    ],
)
append_event(add_size_evt)

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
append_event(change_inventory3_10green)
change_inventory2_13blue = mevents.ChangeInventory(
    id=listing_color_and_size.id,
    variation_ids=[l2_size13, l2_color_blue],
    diff=171,
)
append_event(change_inventory2_13blue)


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
order_open = mevents.CreateOrder(id=rand_obj_id())
append_event(order_open)

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
append_event(add_to_order_open)

# paid
# ====
order_paid = mevents.CreateOrder(id=rand_obj_id())
append_event(order_paid)

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
append_event(add_to_order_paid)

order_paid_add_addr = mevents.UpdateOrder(
    id=order_paid.id,
    set_invoice_address=addr,
)
append_event(order_paid_add_addr)


commit_order_paid = mevents.UpdateOrder(
    id=order_paid.id,
    commit_items=mevents.UpdateOrder.CommitItems(),
)
append_event(commit_order_paid)

choose_payment = mevents.UpdateOrder(
    id=order_paid.id,
    choose_payment=mevents.UpdateOrder.ChoosePaymentMethod(
        currency=c_two,
        payee=payee,
    ),
)
append_event(choose_payment)

# created by the relay on receiving the commit
listing_simple_hash = mtypes.IPFSAddress(cid="/ipfs/foobar")
total_price = int(
    100 * order_paid_item.quantity
    # region shipping
    + 4200
)
order_open_payment_details = mtypes.PaymentDetails(
    # TODO: hash the actual value
    payment_id=random_hash(),
    shipping_region=region_other,
    ttl="1",
    total=mtypes.Uint256(
        raw=total_price.to_bytes(32, "big"),
    ),
    listing_hashes=[listing_simple_hash],
    shop_signature=mtypes.Signature(raw=random.randbytes(64)),
)
finalized_order = mevents.UpdateOrder(
    id=order_paid.id,
    set_payment_details=order_open_payment_details,
)
append_event(finalized_order)

# would be created by the relay after payment is completed
paid_stock_change = mevents.ChangeInventory(
    id=listing_simple.id,
    diff=-42,
)
append_event(paid_stock_change)

order_is_paid = mevents.UpdateOrder(
    id=order_paid.id,
    add_payment_tx=mtypes.OrderTransaction(
        tx_hash=random_hash(),
        block_hash=random_hash(),
    ),
)
append_event(order_is_paid)

# this will be canceled
# =====================
order_canceled = mevents.CreateOrder(
    id=rand_obj_id(),
)
append_event(order_canceled)

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
append_event(add_to_order_canceled)

order_canceled_add_addr = mevents.UpdateOrder(
    id=order_canceled.id,
    set_invoice_address=addr,
)
append_event(order_canceled_add_addr)

commit_order_canceled = mevents.UpdateOrder(
    id=order_canceled.id,
    commit_items=mevents.UpdateOrder.CommitItems(),
)
append_event(commit_order_canceled)

chose_payment_order_canceled = mevents.UpdateOrder(
    id=order_canceled.id,
    choose_payment=mevents.UpdateOrder.ChoosePaymentMethod(
        currency=c_two,
        payee=payee,
    ),
)
append_event(chose_payment_order_canceled)

payment_details3 = mtypes.PaymentDetails(
    payment_id=random_hash(),
    shipping_region=region_local,
    ttl="2",
    shop_signature=mtypes.Signature(raw=random.randbytes(65)),
    total=new_uint256(int(1400 * 1.19 + 500)),
    listing_hashes=[listing_simple_hash],
)
update_order_paid = mevents.UpdateOrder(
    id=order_canceled.id,
    set_payment_details=payment_details3,
)
append_event(update_order_paid)

# 24hrs pass and the sale times out
canceld_at = timestamp_pb2.Timestamp(seconds=23)
update_order_canceled = mevents.UpdateOrder(
    id=order_canceled.id, cancel=mevents.UpdateOrder.Cancel()
)
append_event(update_order_canceled)

# order 4 is in limbo.
# Finalized but not yet payed
# ===========================
order4 = mevents.CreateOrder(
    id=rand_obj_id(),
)
append_event(order4)

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
append_event(add_to_order4)

commit_order4 = mevents.UpdateOrder(
    id=order4.id,
    commit_items=mevents.UpdateOrder.CommitItems(),
)
append_event(commit_order4)

payment_order4 = mevents.UpdateOrder(
    id=order4.id,
    choose_payment=mevents.UpdateOrder.ChoosePaymentMethod(
        currency=c_two,
        payee=payee,
    ),
)
append_event(payment_order4)


order4_add_addr = mevents.UpdateOrder(
    id=order4.id,
    set_invoice_address=addr,
)
append_event(order4_add_addr)

commit_order4 = mtypes.PaymentDetails(
    payment_id=random_hash(),
    ttl="3",
    shop_signature=mtypes.Signature(raw=random.randbytes(65)),
    total=new_uint256(16800),
    listing_hashes=[listing_simple_hash],
)
update_order4 = mevents.UpdateOrder(
    id=order4.id,
    set_payment_details=commit_order4,
)
append_event(update_order4)


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
    ],
    accepted_currencies=[
        vanilla_eth,
        c_two,
    ],
    pricing_currency=vanilla_eth,
)

current_order_open = mstorage.Order(
    id=order_open.id,
    items=[order_open_item],
    invoice_address=addr,
)

current_order_paid = mstorage.Order(
    id=order_paid.id,
    items=[order_paid_item],
    invoice_address=addr,
    chosen_payee=payee,
    chosen_currency=c_two,
    payment_details=order_open_payment_details,
)

current_order_canceled = mstorage.Order(
    id=order_canceled.id,
    canceled_at=canceld_at,
    items=[order_canceled_item],
    invoice_address=addr,
    chosen_payee=payee,
    chosen_currency=c_two,
    payment_details=payment_details3,
)

current_order_unpaid = mstorage.Order(
    id=order4.id,
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
        "signer": {
            "address": kc1.address,
            "key": hex(kc1.key),
        },
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
            id_to_hex(tag_stuff.id): {
                "name": update_tag.rename,
                "item_ids": [listing_simple.id],
            },
            id_to_hex(tag_clothes.id): {
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

# Update listing_simple metadata in the reduced state
for listing in output["reduced"]["listings"]:
    if listing["id"] == id_to_hex(listing_simple.id):
        listing["metadata"] = protobuf_to_dict(update_listing_metadata.metadata)
        break

with open("testVectors.json", "w") as file:
    json.dump(output, file, indent=2, cls=HexEncoder)
