# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Manifest(_message.Message):
    __slots__ = ["token_id", "payees", "accepted_currencies", "base_currency"]
    TOKEN_ID_FIELD_NUMBER: _ClassVar[int]
    PAYEES_FIELD_NUMBER: _ClassVar[int]
    ACCEPTED_CURRENCIES_FIELD_NUMBER: _ClassVar[int]
    BASE_CURRENCY_FIELD_NUMBER: _ClassVar[int]
    token_id: _base_types_pb2.Uint256
    payees: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.Payee]
    accepted_currencies: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ShopCurrency]
    base_currency: _base_types_pb2.ShopCurrency
    def __init__(self, token_id: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ..., payees: _Optional[_Iterable[_Union[_base_types_pb2.Payee, _Mapping]]] = ..., accepted_currencies: _Optional[_Iterable[_Union[_base_types_pb2.ShopCurrency, _Mapping]]] = ..., base_currency: _Optional[_Union[_base_types_pb2.ShopCurrency, _Mapping]] = ...) -> None: ...

class UpdateManifest(_message.Message):
    __slots__ = ["add_payee", "remove_payee", "add_accepted_currencies", "remove_accepted_currencies", "set_base_currency"]
    ADD_PAYEE_FIELD_NUMBER: _ClassVar[int]
    REMOVE_PAYEE_FIELD_NUMBER: _ClassVar[int]
    ADD_ACCEPTED_CURRENCIES_FIELD_NUMBER: _ClassVar[int]
    REMOVE_ACCEPTED_CURRENCIES_FIELD_NUMBER: _ClassVar[int]
    SET_BASE_CURRENCY_FIELD_NUMBER: _ClassVar[int]
    add_payee: _base_types_pb2.Payee
    remove_payee: _base_types_pb2.Payee
    add_accepted_currencies: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ShopCurrency]
    remove_accepted_currencies: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ShopCurrency]
    set_base_currency: _base_types_pb2.ShopCurrency
    def __init__(self, add_payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ..., remove_payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ..., add_accepted_currencies: _Optional[_Iterable[_Union[_base_types_pb2.ShopCurrency, _Mapping]]] = ..., remove_accepted_currencies: _Optional[_Iterable[_Union[_base_types_pb2.ShopCurrency, _Mapping]]] = ..., set_base_currency: _Optional[_Union[_base_types_pb2.ShopCurrency, _Mapping]] = ...) -> None: ...

class Account(_message.Message):
    __slots__ = ["add", "remove", "enroll_keycard", "revoke_keycard"]
    class OnchainAction(_message.Message):
        __slots__ = ["account_address", "tx"]
        ACCOUNT_ADDRESS_FIELD_NUMBER: _ClassVar[int]
        TX_FIELD_NUMBER: _ClassVar[int]
        account_address: _base_types_pb2.EthereumAddress
        tx: _base_types_pb2.Hash
        def __init__(self, account_address: _Optional[_Union[_base_types_pb2.EthereumAddress, _Mapping]] = ..., tx: _Optional[_Union[_base_types_pb2.Hash, _Mapping]] = ...) -> None: ...
    class KeyCardEnroll(_message.Message):
        __slots__ = ["keycard_pubkey", "user_wallet"]
        KEYCARD_PUBKEY_FIELD_NUMBER: _ClassVar[int]
        USER_WALLET_FIELD_NUMBER: _ClassVar[int]
        keycard_pubkey: _base_types_pb2.PublicKey
        user_wallet: _base_types_pb2.EthereumAddress
        def __init__(self, keycard_pubkey: _Optional[_Union[_base_types_pb2.PublicKey, _Mapping]] = ..., user_wallet: _Optional[_Union[_base_types_pb2.EthereumAddress, _Mapping]] = ...) -> None: ...
    ADD_FIELD_NUMBER: _ClassVar[int]
    REMOVE_FIELD_NUMBER: _ClassVar[int]
    ENROLL_KEYCARD_FIELD_NUMBER: _ClassVar[int]
    REVOKE_KEYCARD_FIELD_NUMBER: _ClassVar[int]
    add: Account.OnchainAction
    remove: Account.OnchainAction
    enroll_keycard: Account.KeyCardEnroll
    revoke_keycard: _base_types_pb2.PublicKey
    def __init__(self, add: _Optional[_Union[Account.OnchainAction, _Mapping]] = ..., remove: _Optional[_Union[Account.OnchainAction, _Mapping]] = ..., enroll_keycard: _Optional[_Union[Account.KeyCardEnroll, _Mapping]] = ..., revoke_keycard: _Optional[_Union[_base_types_pb2.PublicKey, _Mapping]] = ...) -> None: ...

class Listing(_message.Message):
    __slots__ = ["id", "base_price", "base_info", "view_state", "options", "stock_status"]
    ID_FIELD_NUMBER: _ClassVar[int]
    BASE_PRICE_FIELD_NUMBER: _ClassVar[int]
    BASE_INFO_FIELD_NUMBER: _ClassVar[int]
    VIEW_STATE_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    STOCK_STATUS_FIELD_NUMBER: _ClassVar[int]
    id: int
    base_price: _base_types_pb2.Uint256
    base_info: _base_types_pb2.ListingMetadata
    view_state: _base_types_pb2.ListingViewState
    options: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ListingOption]
    stock_status: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ListingStockStatus]
    def __init__(self, id: _Optional[int] = ..., base_price: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ..., base_info: _Optional[_Union[_base_types_pb2.ListingMetadata, _Mapping]] = ..., view_state: _Optional[_Union[_base_types_pb2.ListingViewState, str]] = ..., options: _Optional[_Iterable[_Union[_base_types_pb2.ListingOption, _Mapping]]] = ..., stock_status: _Optional[_Iterable[_Union[_base_types_pb2.ListingStockStatus, _Mapping]]] = ...) -> None: ...

class UpdateListing(_message.Message):
    __slots__ = ["listing_id", "base_price", "base_info", "view_state", "add_options", "remove_options", "add_variations", "remove_variations", "update_variations", "stock_updates"]
    class AddVariation(_message.Message):
        __slots__ = ["option_id", "variation"]
        OPTION_ID_FIELD_NUMBER: _ClassVar[int]
        VARIATION_FIELD_NUMBER: _ClassVar[int]
        option_id: int
        variation: _base_types_pb2.ListingVariation
        def __init__(self, option_id: _Optional[int] = ..., variation: _Optional[_Union[_base_types_pb2.ListingVariation, _Mapping]] = ...) -> None: ...
    LISTING_ID_FIELD_NUMBER: _ClassVar[int]
    BASE_PRICE_FIELD_NUMBER: _ClassVar[int]
    BASE_INFO_FIELD_NUMBER: _ClassVar[int]
    VIEW_STATE_FIELD_NUMBER: _ClassVar[int]
    ADD_OPTIONS_FIELD_NUMBER: _ClassVar[int]
    REMOVE_OPTIONS_FIELD_NUMBER: _ClassVar[int]
    ADD_VARIATIONS_FIELD_NUMBER: _ClassVar[int]
    REMOVE_VARIATIONS_FIELD_NUMBER: _ClassVar[int]
    UPDATE_VARIATIONS_FIELD_NUMBER: _ClassVar[int]
    STOCK_UPDATES_FIELD_NUMBER: _ClassVar[int]
    listing_id: int
    base_price: _base_types_pb2.Uint256
    base_info: _base_types_pb2.ListingMetadata
    view_state: _base_types_pb2.ListingViewState
    add_options: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ListingOption]
    remove_options: _containers.RepeatedScalarFieldContainer[int]
    add_variations: _containers.RepeatedCompositeFieldContainer[UpdateListing.AddVariation]
    remove_variations: _containers.RepeatedScalarFieldContainer[int]
    update_variations: _containers.RepeatedCompositeFieldContainer[UpdateListing.AddVariation]
    stock_updates: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ListingStockStatus]
    def __init__(self, listing_id: _Optional[int] = ..., base_price: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ..., base_info: _Optional[_Union[_base_types_pb2.ListingMetadata, _Mapping]] = ..., view_state: _Optional[_Union[_base_types_pb2.ListingViewState, str]] = ..., add_options: _Optional[_Iterable[_Union[_base_types_pb2.ListingOption, _Mapping]]] = ..., remove_options: _Optional[_Iterable[int]] = ..., add_variations: _Optional[_Iterable[_Union[UpdateListing.AddVariation, _Mapping]]] = ..., remove_variations: _Optional[_Iterable[int]] = ..., update_variations: _Optional[_Iterable[_Union[UpdateListing.AddVariation, _Mapping]]] = ..., stock_updates: _Optional[_Iterable[_Union[_base_types_pb2.ListingStockStatus, _Mapping]]] = ...) -> None: ...

class ChangeInventory(_message.Message):
    __slots__ = ["listing_id", "variation_ids", "diff"]
    LISTING_ID_FIELD_NUMBER: _ClassVar[int]
    VARIATION_IDS_FIELD_NUMBER: _ClassVar[int]
    DIFF_FIELD_NUMBER: _ClassVar[int]
    listing_id: int
    variation_ids: _containers.RepeatedScalarFieldContainer[int]
    diff: int
    def __init__(self, listing_id: _Optional[int] = ..., variation_ids: _Optional[_Iterable[int]] = ..., diff: _Optional[int] = ...) -> None: ...

class Tag(_message.Message):
    __slots__ = ["id", "name", "item_ids", "deleted"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ITEM_IDS_FIELD_NUMBER: _ClassVar[int]
    DELETED_FIELD_NUMBER: _ClassVar[int]
    id: int
    name: str
    item_ids: _containers.RepeatedScalarFieldContainer[int]
    deleted: bool
    def __init__(self, id: _Optional[int] = ..., name: _Optional[str] = ..., item_ids: _Optional[_Iterable[int]] = ..., deleted: bool = ...) -> None: ...

class UpdateTag(_message.Message):
    __slots__ = ["tag_id", "rename", "add_item_ids", "remove_item_ids", "delete"]
    TAG_ID_FIELD_NUMBER: _ClassVar[int]
    RENAME_FIELD_NUMBER: _ClassVar[int]
    ADD_ITEM_IDS_FIELD_NUMBER: _ClassVar[int]
    REMOVE_ITEM_IDS_FIELD_NUMBER: _ClassVar[int]
    DELETE_FIELD_NUMBER: _ClassVar[int]
    tag_id: int
    rename: str
    add_item_ids: _containers.RepeatedScalarFieldContainer[int]
    remove_item_ids: _containers.RepeatedScalarFieldContainer[int]
    delete: bool
    def __init__(self, tag_id: _Optional[int] = ..., rename: _Optional[str] = ..., add_item_ids: _Optional[_Iterable[int]] = ..., remove_item_ids: _Optional[_Iterable[int]] = ..., delete: bool = ...) -> None: ...

class Order(_message.Message):
    __slots__ = ["id", "items", "state", "invoice_address", "shipping_address", "canceled_at", "chosen_payee", "chosen_currency", "payment_details", "paid"]
    class State(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        STATE_UNSPECIFIED: _ClassVar[Order.State]
        STATE_OPEN: _ClassVar[Order.State]
        STATE_CANCELED: _ClassVar[Order.State]
        STATE_COMMITED: _ClassVar[Order.State]
        STATE_UNPAID: _ClassVar[Order.State]
        STATE_PAID: _ClassVar[Order.State]
    STATE_UNSPECIFIED: Order.State
    STATE_OPEN: Order.State
    STATE_CANCELED: Order.State
    STATE_COMMITED: Order.State
    STATE_UNPAID: Order.State
    STATE_PAID: Order.State
    ID_FIELD_NUMBER: _ClassVar[int]
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    INVOICE_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    SHIPPING_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    CANCELED_AT_FIELD_NUMBER: _ClassVar[int]
    CHOSEN_PAYEE_FIELD_NUMBER: _ClassVar[int]
    CHOSEN_CURRENCY_FIELD_NUMBER: _ClassVar[int]
    PAYMENT_DETAILS_FIELD_NUMBER: _ClassVar[int]
    PAID_FIELD_NUMBER: _ClassVar[int]
    id: int
    items: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.OrderedItem]
    state: Order.State
    invoice_address: _base_types_pb2.AddressDetails
    shipping_address: _base_types_pb2.AddressDetails
    canceled_at: _timestamp_pb2.Timestamp
    chosen_payee: _base_types_pb2.Payee
    chosen_currency: _base_types_pb2.ShopCurrency
    payment_details: _base_types_pb2.PaymentDetails
    paid: _base_types_pb2.OrderPaid
    def __init__(self, id: _Optional[int] = ..., items: _Optional[_Iterable[_Union[_base_types_pb2.OrderedItem, _Mapping]]] = ..., state: _Optional[_Union[Order.State, str]] = ..., invoice_address: _Optional[_Union[_base_types_pb2.AddressDetails, _Mapping]] = ..., shipping_address: _Optional[_Union[_base_types_pb2.AddressDetails, _Mapping]] = ..., canceled_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., chosen_payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ..., chosen_currency: _Optional[_Union[_base_types_pb2.ShopCurrency, _Mapping]] = ..., payment_details: _Optional[_Union[_base_types_pb2.PaymentDetails, _Mapping]] = ..., paid: _Optional[_Union[_base_types_pb2.OrderPaid, _Mapping]] = ...) -> None: ...

class CreateOrder(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: int
    def __init__(self, id: _Optional[int] = ...) -> None: ...

class UpdateOrder(_message.Message):
    __slots__ = ["order_id", "add_items_to_order", "remove_items_from_order", "canceled", "invoice_address", "shipping_address", "commit", "payment_details", "paid"]
    class CommitItems(_message.Message):
        __slots__ = ["currency", "payee", "commited_at"]
        CURRENCY_FIELD_NUMBER: _ClassVar[int]
        PAYEE_FIELD_NUMBER: _ClassVar[int]
        COMMITED_AT_FIELD_NUMBER: _ClassVar[int]
        currency: _base_types_pb2.ShopCurrency
        payee: _base_types_pb2.Payee
        commited_at: _timestamp_pb2.Timestamp
        def __init__(self, currency: _Optional[_Union[_base_types_pb2.ShopCurrency, _Mapping]] = ..., payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ..., commited_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
    class Canceled(_message.Message):
        __slots__ = ["canceld_at"]
        CANCELD_AT_FIELD_NUMBER: _ClassVar[int]
        canceld_at: _timestamp_pb2.Timestamp
        def __init__(self, canceld_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    ADD_ITEMS_TO_ORDER_FIELD_NUMBER: _ClassVar[int]
    REMOVE_ITEMS_FROM_ORDER_FIELD_NUMBER: _ClassVar[int]
    CANCELED_FIELD_NUMBER: _ClassVar[int]
    INVOICE_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    SHIPPING_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    COMMIT_FIELD_NUMBER: _ClassVar[int]
    PAYMENT_DETAILS_FIELD_NUMBER: _ClassVar[int]
    PAID_FIELD_NUMBER: _ClassVar[int]
    order_id: int
    add_items_to_order: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.OrderedItem]
    remove_items_from_order: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.OrderedItem]
    canceled: UpdateOrder.Canceled
    invoice_address: _base_types_pb2.AddressDetails
    shipping_address: _base_types_pb2.AddressDetails
    commit: UpdateOrder.CommitItems
    payment_details: _base_types_pb2.PaymentDetails
    paid: _base_types_pb2.OrderPaid
    def __init__(self, order_id: _Optional[int] = ..., add_items_to_order: _Optional[_Iterable[_Union[_base_types_pb2.OrderedItem, _Mapping]]] = ..., remove_items_from_order: _Optional[_Iterable[_Union[_base_types_pb2.OrderedItem, _Mapping]]] = ..., canceled: _Optional[_Union[UpdateOrder.Canceled, _Mapping]] = ..., invoice_address: _Optional[_Union[_base_types_pb2.AddressDetails, _Mapping]] = ..., shipping_address: _Optional[_Union[_base_types_pb2.AddressDetails, _Mapping]] = ..., commit: _Optional[_Union[UpdateOrder.CommitItems, _Mapping]] = ..., payment_details: _Optional[_Union[_base_types_pb2.PaymentDetails, _Mapping]] = ..., paid: _Optional[_Union[_base_types_pb2.OrderPaid, _Mapping]] = ...) -> None: ...

class ShopEvent(_message.Message):
    __slots__ = ["nonce", "shop_id", "timestamp", "manifest", "update_manifest", "account", "listing", "update_listing", "change_inventory", "tag", "update_tag", "create_order", "update_order"]
    NONCE_FIELD_NUMBER: _ClassVar[int]
    SHOP_ID_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    MANIFEST_FIELD_NUMBER: _ClassVar[int]
    UPDATE_MANIFEST_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_FIELD_NUMBER: _ClassVar[int]
    LISTING_FIELD_NUMBER: _ClassVar[int]
    UPDATE_LISTING_FIELD_NUMBER: _ClassVar[int]
    CHANGE_INVENTORY_FIELD_NUMBER: _ClassVar[int]
    TAG_FIELD_NUMBER: _ClassVar[int]
    UPDATE_TAG_FIELD_NUMBER: _ClassVar[int]
    CREATE_ORDER_FIELD_NUMBER: _ClassVar[int]
    UPDATE_ORDER_FIELD_NUMBER: _ClassVar[int]
    nonce: int
    shop_id: _base_types_pb2.Uint256
    timestamp: _timestamp_pb2.Timestamp
    manifest: Manifest
    update_manifest: UpdateManifest
    account: Account
    listing: Listing
    update_listing: UpdateListing
    change_inventory: ChangeInventory
    tag: Tag
    update_tag: UpdateTag
    create_order: CreateOrder
    update_order: UpdateOrder
    def __init__(self, nonce: _Optional[int] = ..., shop_id: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ..., timestamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ..., manifest: _Optional[_Union[Manifest, _Mapping]] = ..., update_manifest: _Optional[_Union[UpdateManifest, _Mapping]] = ..., account: _Optional[_Union[Account, _Mapping]] = ..., listing: _Optional[_Union[Listing, _Mapping]] = ..., update_listing: _Optional[_Union[UpdateListing, _Mapping]] = ..., change_inventory: _Optional[_Union[ChangeInventory, _Mapping]] = ..., tag: _Optional[_Union[Tag, _Mapping]] = ..., update_tag: _Optional[_Union[UpdateTag, _Mapping]] = ..., create_order: _Optional[_Union[CreateOrder, _Mapping]] = ..., update_order: _Optional[_Union[UpdateOrder, _Mapping]] = ...) -> None: ...
