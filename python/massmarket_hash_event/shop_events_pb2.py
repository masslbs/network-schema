# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: shop_events.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder

# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()
from massmarket_hash_event import base_types_pb2 as base__types__pb2
from google.protobuf import timestamp_pb2 as google_dot_protobuf_dot_timestamp__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(
    b'\n\x11shop_events.proto\x12\x0bmarket.mass\x1a\x10\x62\x61se_types.proto\x1a\x1fgoogle/protobuf/timestamp.proto"\xba\x02\n\x08Manifest\x12&\n\x08token_id\x18\x01 \x01(\x0b\x32\x14.market.mass.Uint256\x12"\n\x06payees\x18\x02 \x03(\x0b\x32\x12.market.mass.Payee\x12\x36\n\x13\x61\x63\x63\x65pted_currencies\x18\x03 \x03(\x0b\x32\x19.market.mass.ShopCurrency\x12\x33\n\x10pricing_currency\x18\x04 \x01(\x0b\x32\x19.market.mass.ShopCurrency\x12\x35\n\x10shipping_regions\x18\x05 \x03(\x0b\x32\x1b.market.mass.ShippingRegion\x12>\n\x15order_price_modifiers\x18\x06 \x03(\x0b\x32\x1f.market.mass.OrderPriceModifier"\xa2\x04\n\x0eUpdateManifest\x12*\n\tadd_payee\x18\x02 \x01(\x0b\x32\x12.market.mass.PayeeH\x00\x88\x01\x01\x12-\n\x0cremove_payee\x18\x03 \x01(\x0b\x32\x12.market.mass.PayeeH\x01\x88\x01\x01\x12:\n\x17\x61\x64\x64_accepted_currencies\x18\x04 \x03(\x0b\x32\x19.market.mass.ShopCurrency\x12=\n\x1aremove_accepted_currencies\x18\x05 \x03(\x0b\x32\x19.market.mass.ShopCurrency\x12<\n\x14set_pricing_currency\x18\x06 \x01(\x0b\x32\x19.market.mass.ShopCurrencyH\x02\x88\x01\x01\x12\x42\n\x19\x61\x64\x64_order_price_modifiers\x18\t \x03(\x0b\x32\x1f.market.mass.OrderPriceModifier\x12$\n\x1cremove_order_price_modifiers\x18\n \x03(\x04\x12\x39\n\x14\x61\x64\x64_shipping_regions\x18\x07 \x03(\x0b\x32\x1b.market.mass.ShippingRegion\x12\x1f\n\x17remove_shipping_regions\x18\x08 \x03(\tB\x0c\n\n_add_payeeB\x0f\n\r_remove_payeeB\x17\n\x15_set_pricing_currency"\xc7\x03\n\x07\x41\x63\x63ount\x12\x31\n\x03\x61\x64\x64\x18\x01 \x01(\x0b\x32".market.mass.Account.OnchainActionH\x00\x12\x34\n\x06remove\x18\x02 \x01(\x0b\x32".market.mass.Account.OnchainActionH\x00\x12<\n\x0e\x65nroll_keycard\x18\x03 \x01(\x0b\x32".market.mass.Account.KeyCardEnrollH\x00\x12\x30\n\x0erevoke_keycard\x18\x04 \x01(\x0b\x32\x16.market.mass.PublicKeyH\x00\x1a\x65\n\rOnchainAction\x12\x35\n\x0f\x61\x63\x63ount_address\x18\x01 \x01(\x0b\x32\x1c.market.mass.EthereumAddress\x12\x1d\n\x02tx\x18\x02 \x01(\x0b\x32\x11.market.mass.Hash\x1ar\n\rKeyCardEnroll\x12.\n\x0ekeycard_pubkey\x18\x01 \x01(\x0b\x32\x16.market.mass.PublicKey\x12\x31\n\x0buser_wallet\x18\x02 \x01(\x0b\x32\x1c.market.mass.EthereumAddressB\x08\n\x06\x61\x63tion"\x84\x02\n\x07Listing\x12\n\n\x02id\x18\x01 \x01(\x04\x12#\n\x05price\x18\x02 \x01(\x0b\x32\x14.market.mass.Uint256\x12/\n\tbase_info\x18\x03 \x01(\x0b\x32\x1c.market.mass.ListingMetadata\x12\x31\n\nview_state\x18\x04 \x01(\x0e\x32\x1d.market.mass.ListingViewState\x12+\n\x07options\x18\x05 \x03(\x0b\x32\x1a.market.mass.ListingOption\x12\x37\n\x0estock_statuses\x18\x06 \x03(\x0b\x32\x1f.market.mass.ListingStockStatus"\xda\x04\n\rUpdateListing\x12\n\n\x02id\x18\x01 \x01(\x04\x12-\n\nbase_price\x18\x02 \x01(\x0b\x32\x14.market.mass.Uint256H\x00\x88\x01\x01\x12\x34\n\tbase_info\x18\x03 \x01(\x0b\x32\x1c.market.mass.ListingMetadataH\x01\x88\x01\x01\x12\x36\n\nview_state\x18\x04 \x01(\x0e\x32\x1d.market.mass.ListingViewStateH\x02\x88\x01\x01\x12/\n\x0b\x61\x64\x64_options\x18\x06 \x03(\x0b\x32\x1a.market.mass.ListingOption\x12\x16\n\x0eremove_options\x18\x07 \x03(\x04\x12?\n\x0e\x61\x64\x64_variations\x18\t \x03(\x0b\x32\'.market.mass.UpdateListing.AddVariation\x12\x19\n\x11remove_variations\x18\x08 \x03(\x04\x12\x42\n\x11update_variations\x18\n \x03(\x0b\x32\'.market.mass.UpdateListing.AddVariation\x12\x36\n\rstock_updates\x18\x0b \x03(\x0b\x32\x1f.market.mass.ListingStockStatus\x1aS\n\x0c\x41\x64\x64Variation\x12\x11\n\toption_id\x18\x01 \x01(\x04\x12\x30\n\tvariation\x18\x02 \x01(\x0b\x32\x1d.market.mass.ListingVariationB\r\n\x0b_base_priceB\x0c\n\n_base_infoB\r\n\x0b_view_state"B\n\x0f\x43hangeInventory\x12\n\n\x02id\x18\x01 \x01(\x04\x12\x15\n\rvariation_ids\x18\x02 \x03(\x04\x12\x0c\n\x04\x64iff\x18\x03 \x01(\x11"E\n\x03Tag\x12\n\n\x02id\x18\x01 \x01(\x04\x12\x0c\n\x04name\x18\x02 \x01(\t\x12\x13\n\x0blisting_ids\x18\x03 \x03(\x04\x12\x0f\n\x07\x64\x65leted\x18\x04 \x01(\x08"\x8c\x01\n\tUpdateTag\x12\n\n\x02id\x18\x01 \x01(\x04\x12\x13\n\x06rename\x18\x02 \x01(\tH\x00\x88\x01\x01\x12\x17\n\x0f\x61\x64\x64_listing_ids\x18\x03 \x03(\x04\x12\x1a\n\x12remove_listing_ids\x18\x04 \x03(\x04\x12\x13\n\x06\x64\x65lete\x18\x05 \x01(\x08H\x01\x88\x01\x01\x42\t\n\x07_renameB\t\n\x07_delete"\xd5\x05\n\x05Order\x12\n\n\x02id\x18\x01 \x01(\x04\x12\'\n\x05items\x18\x02 \x03(\x0b\x32\x18.market.mass.OrderedItem\x12\'\n\x05state\x18\x03 \x01(\x0e\x32\x18.market.mass.Order.State\x12\x39\n\x0finvoice_address\x18\x04 \x01(\x0b\x32\x1b.market.mass.AddressDetailsH\x00\x88\x01\x01\x12:\n\x10shipping_address\x18\x05 \x01(\x0b\x32\x1b.market.mass.AddressDetailsH\x01\x88\x01\x01\x12\x34\n\x0b\x63\x61nceled_at\x18\x06 \x01(\x0b\x32\x1a.google.protobuf.TimestampH\x02\x88\x01\x01\x12-\n\x0c\x63hosen_payee\x18\x07 \x01(\x0b\x32\x12.market.mass.PayeeH\x03\x88\x01\x01\x12\x37\n\x0f\x63hosen_currency\x18\x08 \x01(\x0b\x32\x19.market.mass.ShopCurrencyH\x04\x88\x01\x01\x12\x39\n\x0fpayment_details\x18\t \x01(\x0b\x32\x1b.market.mass.PaymentDetailsH\x05\x88\x01\x01\x12)\n\x04paid\x18\n \x01(\x0b\x32\x16.market.mass.OrderPaidH\x06\x88\x01\x01"x\n\x05State\x12\x15\n\x11STATE_UNSPECIFIED\x10\x00\x12\x0e\n\nSTATE_OPEN\x10\x01\x12\x12\n\x0eSTATE_CANCELED\x10\x02\x12\x12\n\x0eSTATE_COMMITED\x10\x03\x12\x10\n\x0cSTATE_UNPAID\x10\x04\x12\x0e\n\nSTATE_PAID\x10\x05\x42\x12\n\x10_invoice_addressB\x13\n\x11_shipping_addressB\x0e\n\x0c_canceled_atB\x0f\n\r_chosen_payeeB\x12\n\x10_chosen_currencyB\x12\n\x10_payment_detailsB\x07\n\x05_paid"\x19\n\x0b\x43reateOrder\x12\n\n\x02id\x18\x01 \x01(\x04"\xb5\x06\n\x0bUpdateOrder\x12\n\n\x02id\x18\x01 \x01(\x04\x12\x35\n\x08\x63\x61nceled\x18\x02 \x01(\x0b\x32!.market.mass.UpdateOrder.CanceledH\x00\x12<\n\x0c\x63hange_items\x18\x03 \x01(\x0b\x32$.market.mass.UpdateOrder.ChangeItemsH\x00\x12<\n\x0c\x63ommit_items\x18\x04 \x01(\x0b\x32$.market.mass.UpdateOrder.CommitItemsH\x00\x12\x36\n\x0finvoice_address\x18\x05 \x01(\x0b\x32\x1b.market.mass.AddressDetailsH\x00\x12\x37\n\x10shipping_address\x18\x06 \x01(\x0b\x32\x1b.market.mass.AddressDetailsH\x00\x12\x46\n\x0e\x63hoose_payment\x18\x07 \x01(\x0b\x32,.market.mass.UpdateOrder.ChoosePaymentMethodH\x00\x12\x36\n\x0fpayment_details\x18\x08 \x01(\x0b\x32\x1b.market.mass.PaymentDetailsH\x00\x12&\n\x04paid\x18\t \x01(\x0b\x32\x16.market.mass.OrderPaidH\x00\x1a`\n\x0b\x43hangeItems\x12&\n\x04\x61\x64\x64s\x18\x01 \x03(\x0b\x32\x18.market.mass.OrderedItem\x12)\n\x07removes\x18\x02 \x03(\x0b\x32\x18.market.mass.OrderedItem\x1a\r\n\x0b\x43ommitItems\x1a\x96\x01\n\x13\x43hoosePaymentMethod\x12+\n\x08\x63urrency\x18\x01 \x01(\x0b\x32\x19.market.mass.ShopCurrency\x12!\n\x05payee\x18\x02 \x01(\x0b\x32\x12.market.mass.Payee\x12/\n\x0b\x63ommited_at\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x1a:\n\x08\x43\x61nceled\x12.\n\ncanceld_at\x18\x01 \x01(\x0b\x32\x1a.google.protobuf.TimestampB\x08\n\x06\x61\x63tion"\xd1\x04\n\tShopEvent\x12\r\n\x05nonce\x18\x01 \x01(\x04\x12%\n\x07shop_id\x18\x02 \x01(\x0b\x32\x14.market.mass.Uint256\x12-\n\ttimestamp\x18\x03 \x01(\x0b\x32\x1a.google.protobuf.Timestamp\x12)\n\x08manifest\x18\x04 \x01(\x0b\x32\x15.market.mass.ManifestH\x00\x12\x36\n\x0fupdate_manifest\x18\x05 \x01(\x0b\x32\x1b.market.mass.UpdateManifestH\x00\x12\'\n\x07\x61\x63\x63ount\x18\x06 \x01(\x0b\x32\x14.market.mass.AccountH\x00\x12\'\n\x07listing\x18\x07 \x01(\x0b\x32\x14.market.mass.ListingH\x00\x12\x34\n\x0eupdate_listing\x18\x08 \x01(\x0b\x32\x1a.market.mass.UpdateListingH\x00\x12\x38\n\x10\x63hange_inventory\x18\t \x01(\x0b\x32\x1c.market.mass.ChangeInventoryH\x00\x12\x1f\n\x03tag\x18\n \x01(\x0b\x32\x10.market.mass.TagH\x00\x12,\n\nupdate_tag\x18\x0b \x01(\x0b\x32\x16.market.mass.UpdateTagH\x00\x12\x30\n\x0c\x63reate_order\x18\x0c \x01(\x0b\x32\x18.market.mass.CreateOrderH\x00\x12\x30\n\x0cupdate_order\x18\r \x01(\x0b\x32\x18.market.mass.UpdateOrderH\x00\x42\x07\n\x05unionb\x06proto3'
)

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, "shop_events_pb2", _globals)
if _descriptor._USE_C_DESCRIPTORS == False:
    DESCRIPTOR._options = None
    _globals["_MANIFEST"]._serialized_start = 86
    _globals["_MANIFEST"]._serialized_end = 400
    _globals["_UPDATEMANIFEST"]._serialized_start = 403
    _globals["_UPDATEMANIFEST"]._serialized_end = 949
    _globals["_ACCOUNT"]._serialized_start = 952
    _globals["_ACCOUNT"]._serialized_end = 1407
    _globals["_ACCOUNT_ONCHAINACTION"]._serialized_start = 1180
    _globals["_ACCOUNT_ONCHAINACTION"]._serialized_end = 1281
    _globals["_ACCOUNT_KEYCARDENROLL"]._serialized_start = 1283
    _globals["_ACCOUNT_KEYCARDENROLL"]._serialized_end = 1397
    _globals["_LISTING"]._serialized_start = 1410
    _globals["_LISTING"]._serialized_end = 1670
    _globals["_UPDATELISTING"]._serialized_start = 1673
    _globals["_UPDATELISTING"]._serialized_end = 2275
    _globals["_UPDATELISTING_ADDVARIATION"]._serialized_start = 2148
    _globals["_UPDATELISTING_ADDVARIATION"]._serialized_end = 2231
    _globals["_CHANGEINVENTORY"]._serialized_start = 2277
    _globals["_CHANGEINVENTORY"]._serialized_end = 2343
    _globals["_TAG"]._serialized_start = 2345
    _globals["_TAG"]._serialized_end = 2414
    _globals["_UPDATETAG"]._serialized_start = 2417
    _globals["_UPDATETAG"]._serialized_end = 2557
    _globals["_ORDER"]._serialized_start = 2560
    _globals["_ORDER"]._serialized_end = 3285
    _globals["_ORDER_STATE"]._serialized_start = 3042
    _globals["_ORDER_STATE"]._serialized_end = 3162
    _globals["_CREATEORDER"]._serialized_start = 3287
    _globals["_CREATEORDER"]._serialized_end = 3312
    _globals["_UPDATEORDER"]._serialized_start = 3315
    _globals["_UPDATEORDER"]._serialized_end = 4136
    _globals["_UPDATEORDER_CHANGEITEMS"]._serialized_start = 3802
    _globals["_UPDATEORDER_CHANGEITEMS"]._serialized_end = 3898
    _globals["_UPDATEORDER_COMMITITEMS"]._serialized_start = 3900
    _globals["_UPDATEORDER_COMMITITEMS"]._serialized_end = 3913
    _globals["_UPDATEORDER_CHOOSEPAYMENTMETHOD"]._serialized_start = 3916
    _globals["_UPDATEORDER_CHOOSEPAYMENTMETHOD"]._serialized_end = 4066
    _globals["_UPDATEORDER_CANCELED"]._serialized_start = 4068
    _globals["_UPDATEORDER_CANCELED"]._serialized_end = 4126
    _globals["_SHOPEVENT"]._serialized_start = 4139
    _globals["_SHOPEVENT"]._serialized_end = 4732
# @@protoc_insertion_point(module_scope)
