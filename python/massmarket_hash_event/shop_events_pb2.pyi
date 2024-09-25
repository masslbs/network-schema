# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import (
    ClassVar as _ClassVar,
    Iterable as _Iterable,
    Mapping as _Mapping,
    Optional as _Optional,
    Union as _Union,
)

DESCRIPTOR: _descriptor.FileDescriptor

class Manifest(_message.Message):
    __slots__ = [
        "token_id",
        "payees",
        "accepted_currencies",
        "pricing_currency",
        "shipping_regions",
        "order_price_modifiers",
    ]
    TOKEN_ID_FIELD_NUMBER: _ClassVar[int]
    PAYEES_FIELD_NUMBER: _ClassVar[int]
    ACCEPTED_CURRENCIES_FIELD_NUMBER: _ClassVar[int]
    PRICING_CURRENCY_FIELD_NUMBER: _ClassVar[int]
    SHIPPING_REGIONS_FIELD_NUMBER: _ClassVar[int]
    ORDER_PRICE_MODIFIERS_FIELD_NUMBER: _ClassVar[int]
    token_id: _base_types_pb2.Uint256
    payees: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.Payee]
    accepted_currencies: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ShopCurrency
    ]
    pricing_currency: _base_types_pb2.ShopCurrency
    shipping_regions: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ShippingRegion
    ]
    order_price_modifiers: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.OrderPriceModifier
    ]
    def __init__(
        self,
        token_id: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ...,
        payees: _Optional[_Iterable[_Union[_base_types_pb2.Payee, _Mapping]]] = ...,
        accepted_currencies: _Optional[
            _Iterable[_Union[_base_types_pb2.ShopCurrency, _Mapping]]
        ] = ...,
        pricing_currency: _Optional[
            _Union[_base_types_pb2.ShopCurrency, _Mapping]
        ] = ...,
        shipping_regions: _Optional[
            _Iterable[_Union[_base_types_pb2.ShippingRegion, _Mapping]]
        ] = ...,
        order_price_modifiers: _Optional[
            _Iterable[_Union[_base_types_pb2.OrderPriceModifier, _Mapping]]
        ] = ...,
    ) -> None: ...

class UpdateManifest(_message.Message):
    __slots__ = [
        "add_payee",
        "remove_payee",
        "add_accepted_currencies",
        "remove_accepted_currencies",
        "set_pricing_currency",
        "add_order_price_modifiers",
        "remove_order_price_modifier_ids",
        "add_shipping_regions",
        "remove_shipping_regions",
    ]
    ADD_PAYEE_FIELD_NUMBER: _ClassVar[int]
    REMOVE_PAYEE_FIELD_NUMBER: _ClassVar[int]
    ADD_ACCEPTED_CURRENCIES_FIELD_NUMBER: _ClassVar[int]
    REMOVE_ACCEPTED_CURRENCIES_FIELD_NUMBER: _ClassVar[int]
    SET_PRICING_CURRENCY_FIELD_NUMBER: _ClassVar[int]
    ADD_ORDER_PRICE_MODIFIERS_FIELD_NUMBER: _ClassVar[int]
    REMOVE_ORDER_PRICE_MODIFIER_IDS_FIELD_NUMBER: _ClassVar[int]
    ADD_SHIPPING_REGIONS_FIELD_NUMBER: _ClassVar[int]
    REMOVE_SHIPPING_REGIONS_FIELD_NUMBER: _ClassVar[int]
    add_payee: _base_types_pb2.Payee
    remove_payee: _base_types_pb2.Payee
    add_accepted_currencies: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ShopCurrency
    ]
    remove_accepted_currencies: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ShopCurrency
    ]
    set_pricing_currency: _base_types_pb2.ShopCurrency
    add_order_price_modifiers: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.OrderPriceModifier
    ]
    remove_order_price_modifier_ids: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ObjectId
    ]
    add_shipping_regions: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ShippingRegion
    ]
    remove_shipping_regions: _containers.RepeatedScalarFieldContainer[str]
    def __init__(
        self,
        add_payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ...,
        remove_payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ...,
        add_accepted_currencies: _Optional[
            _Iterable[_Union[_base_types_pb2.ShopCurrency, _Mapping]]
        ] = ...,
        remove_accepted_currencies: _Optional[
            _Iterable[_Union[_base_types_pb2.ShopCurrency, _Mapping]]
        ] = ...,
        set_pricing_currency: _Optional[
            _Union[_base_types_pb2.ShopCurrency, _Mapping]
        ] = ...,
        add_order_price_modifiers: _Optional[
            _Iterable[_Union[_base_types_pb2.OrderPriceModifier, _Mapping]]
        ] = ...,
        remove_order_price_modifier_ids: _Optional[
            _Iterable[_Union[_base_types_pb2.ObjectId, _Mapping]]
        ] = ...,
        add_shipping_regions: _Optional[
            _Iterable[_Union[_base_types_pb2.ShippingRegion, _Mapping]]
        ] = ...,
        remove_shipping_regions: _Optional[_Iterable[str]] = ...,
    ) -> None: ...

class Account(_message.Message):
    __slots__ = ["add", "remove", "enroll_keycard", "revoke_keycard"]

    class OnchainAction(_message.Message):
        __slots__ = ["account_address", "tx"]
        ACCOUNT_ADDRESS_FIELD_NUMBER: _ClassVar[int]
        TX_FIELD_NUMBER: _ClassVar[int]
        account_address: _base_types_pb2.EthereumAddress
        tx: _base_types_pb2.Hash
        def __init__(
            self,
            account_address: _Optional[
                _Union[_base_types_pb2.EthereumAddress, _Mapping]
            ] = ...,
            tx: _Optional[_Union[_base_types_pb2.Hash, _Mapping]] = ...,
        ) -> None: ...

    class KeyCardEnroll(_message.Message):
        __slots__ = ["keycard_pubkey", "user_wallet"]
        KEYCARD_PUBKEY_FIELD_NUMBER: _ClassVar[int]
        USER_WALLET_FIELD_NUMBER: _ClassVar[int]
        keycard_pubkey: _base_types_pb2.PublicKey
        user_wallet: _base_types_pb2.EthereumAddress
        def __init__(
            self,
            keycard_pubkey: _Optional[
                _Union[_base_types_pb2.PublicKey, _Mapping]
            ] = ...,
            user_wallet: _Optional[
                _Union[_base_types_pb2.EthereumAddress, _Mapping]
            ] = ...,
        ) -> None: ...

    ADD_FIELD_NUMBER: _ClassVar[int]
    REMOVE_FIELD_NUMBER: _ClassVar[int]
    ENROLL_KEYCARD_FIELD_NUMBER: _ClassVar[int]
    REVOKE_KEYCARD_FIELD_NUMBER: _ClassVar[int]
    add: Account.OnchainAction
    remove: Account.OnchainAction
    enroll_keycard: Account.KeyCardEnroll
    revoke_keycard: _base_types_pb2.PublicKey
    def __init__(
        self,
        add: _Optional[_Union[Account.OnchainAction, _Mapping]] = ...,
        remove: _Optional[_Union[Account.OnchainAction, _Mapping]] = ...,
        enroll_keycard: _Optional[_Union[Account.KeyCardEnroll, _Mapping]] = ...,
        revoke_keycard: _Optional[_Union[_base_types_pb2.PublicKey, _Mapping]] = ...,
    ) -> None: ...

class Listing(_message.Message):
    __slots__ = ["id", "price", "metadata", "view_state", "options", "stock_statuses"]
    ID_FIELD_NUMBER: _ClassVar[int]
    PRICE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    VIEW_STATE_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    STOCK_STATUSES_FIELD_NUMBER: _ClassVar[int]
    id: _base_types_pb2.ObjectId
    price: _base_types_pb2.Uint256
    metadata: _base_types_pb2.ListingMetadata
    view_state: _base_types_pb2.ListingViewState
    options: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ListingOption]
    stock_statuses: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ListingStockStatus
    ]
    def __init__(
        self,
        id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        price: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ...,
        metadata: _Optional[_Union[_base_types_pb2.ListingMetadata, _Mapping]] = ...,
        view_state: _Optional[_Union[_base_types_pb2.ListingViewState, str]] = ...,
        options: _Optional[
            _Iterable[_Union[_base_types_pb2.ListingOption, _Mapping]]
        ] = ...,
        stock_statuses: _Optional[
            _Iterable[_Union[_base_types_pb2.ListingStockStatus, _Mapping]]
        ] = ...,
    ) -> None: ...

class UpdateListing(_message.Message):
    __slots__ = [
        "id",
        "price",
        "metadata",
        "view_state",
        "add_options",
        "remove_option_ids",
        "add_variations",
        "remove_variation_ids",
        "update_variations",
        "stock_updates",
    ]

    class AddVariation(_message.Message):
        __slots__ = ["option_id", "variation"]
        OPTION_ID_FIELD_NUMBER: _ClassVar[int]
        VARIATION_FIELD_NUMBER: _ClassVar[int]
        option_id: _base_types_pb2.ObjectId
        variation: _base_types_pb2.ListingVariation
        def __init__(
            self,
            option_id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
            variation: _Optional[
                _Union[_base_types_pb2.ListingVariation, _Mapping]
            ] = ...,
        ) -> None: ...

    ID_FIELD_NUMBER: _ClassVar[int]
    PRICE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    VIEW_STATE_FIELD_NUMBER: _ClassVar[int]
    ADD_OPTIONS_FIELD_NUMBER: _ClassVar[int]
    REMOVE_OPTION_IDS_FIELD_NUMBER: _ClassVar[int]
    ADD_VARIATIONS_FIELD_NUMBER: _ClassVar[int]
    REMOVE_VARIATION_IDS_FIELD_NUMBER: _ClassVar[int]
    UPDATE_VARIATIONS_FIELD_NUMBER: _ClassVar[int]
    STOCK_UPDATES_FIELD_NUMBER: _ClassVar[int]
    id: _base_types_pb2.ObjectId
    price: _base_types_pb2.Uint256
    metadata: _base_types_pb2.ListingMetadata
    view_state: _base_types_pb2.ListingViewState
    add_options: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ListingOption
    ]
    remove_option_ids: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ObjectId
    ]
    add_variations: _containers.RepeatedCompositeFieldContainer[
        UpdateListing.AddVariation
    ]
    remove_variation_ids: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ObjectId
    ]
    update_variations: _containers.RepeatedCompositeFieldContainer[
        UpdateListing.AddVariation
    ]
    stock_updates: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ListingStockStatus
    ]
    def __init__(
        self,
        id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        price: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ...,
        metadata: _Optional[_Union[_base_types_pb2.ListingMetadata, _Mapping]] = ...,
        view_state: _Optional[_Union[_base_types_pb2.ListingViewState, str]] = ...,
        add_options: _Optional[
            _Iterable[_Union[_base_types_pb2.ListingOption, _Mapping]]
        ] = ...,
        remove_option_ids: _Optional[
            _Iterable[_Union[_base_types_pb2.ObjectId, _Mapping]]
        ] = ...,
        add_variations: _Optional[
            _Iterable[_Union[UpdateListing.AddVariation, _Mapping]]
        ] = ...,
        remove_variation_ids: _Optional[
            _Iterable[_Union[_base_types_pb2.ObjectId, _Mapping]]
        ] = ...,
        update_variations: _Optional[
            _Iterable[_Union[UpdateListing.AddVariation, _Mapping]]
        ] = ...,
        stock_updates: _Optional[
            _Iterable[_Union[_base_types_pb2.ListingStockStatus, _Mapping]]
        ] = ...,
    ) -> None: ...

class ChangeInventory(_message.Message):
    __slots__ = ["id", "variation_ids", "diff"]
    ID_FIELD_NUMBER: _ClassVar[int]
    VARIATION_IDS_FIELD_NUMBER: _ClassVar[int]
    DIFF_FIELD_NUMBER: _ClassVar[int]
    id: _base_types_pb2.ObjectId
    variation_ids: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ObjectId]
    diff: int
    def __init__(
        self,
        id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        variation_ids: _Optional[
            _Iterable[_Union[_base_types_pb2.ObjectId, _Mapping]]
        ] = ...,
        diff: _Optional[int] = ...,
    ) -> None: ...

class Tag(_message.Message):
    __slots__ = ["id", "name", "listing_ids", "deleted"]
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    LISTING_IDS_FIELD_NUMBER: _ClassVar[int]
    DELETED_FIELD_NUMBER: _ClassVar[int]
    id: _base_types_pb2.ObjectId
    name: str
    listing_ids: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.ObjectId]
    deleted: bool
    def __init__(
        self,
        id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        name: _Optional[str] = ...,
        listing_ids: _Optional[
            _Iterable[_Union[_base_types_pb2.ObjectId, _Mapping]]
        ] = ...,
        deleted: bool = ...,
    ) -> None: ...

class UpdateTag(_message.Message):
    __slots__ = ["id", "rename", "add_listing_ids", "remove_listing_ids", "delete"]
    ID_FIELD_NUMBER: _ClassVar[int]
    RENAME_FIELD_NUMBER: _ClassVar[int]
    ADD_LISTING_IDS_FIELD_NUMBER: _ClassVar[int]
    REMOVE_LISTING_IDS_FIELD_NUMBER: _ClassVar[int]
    DELETE_FIELD_NUMBER: _ClassVar[int]
    id: _base_types_pb2.ObjectId
    rename: str
    add_listing_ids: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ObjectId
    ]
    remove_listing_ids: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.ObjectId
    ]
    delete: bool
    def __init__(
        self,
        id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        rename: _Optional[str] = ...,
        add_listing_ids: _Optional[
            _Iterable[_Union[_base_types_pb2.ObjectId, _Mapping]]
        ] = ...,
        remove_listing_ids: _Optional[
            _Iterable[_Union[_base_types_pb2.ObjectId, _Mapping]]
        ] = ...,
        delete: bool = ...,
    ) -> None: ...

class Order(_message.Message):
    __slots__ = [
        "id",
        "items",
        "state",
        "invoice_address",
        "shipping_address",
        "canceled_at",
        "chosen_payee",
        "chosen_currency",
        "payment_details",
        "paid",
    ]

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
    id: _base_types_pb2.ObjectId
    items: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.OrderedItem]
    state: Order.State
    invoice_address: _base_types_pb2.AddressDetails
    shipping_address: _base_types_pb2.AddressDetails
    canceled_at: _timestamp_pb2.Timestamp
    chosen_payee: _base_types_pb2.Payee
    chosen_currency: _base_types_pb2.ShopCurrency
    payment_details: _base_types_pb2.PaymentDetails
    paid: _base_types_pb2.OrderPaid
    def __init__(
        self,
        id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        items: _Optional[
            _Iterable[_Union[_base_types_pb2.OrderedItem, _Mapping]]
        ] = ...,
        state: _Optional[_Union[Order.State, str]] = ...,
        invoice_address: _Optional[
            _Union[_base_types_pb2.AddressDetails, _Mapping]
        ] = ...,
        shipping_address: _Optional[
            _Union[_base_types_pb2.AddressDetails, _Mapping]
        ] = ...,
        canceled_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...,
        chosen_payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ...,
        chosen_currency: _Optional[
            _Union[_base_types_pb2.ShopCurrency, _Mapping]
        ] = ...,
        payment_details: _Optional[
            _Union[_base_types_pb2.PaymentDetails, _Mapping]
        ] = ...,
        paid: _Optional[_Union[_base_types_pb2.OrderPaid, _Mapping]] = ...,
    ) -> None: ...

class CreateOrder(_message.Message):
    __slots__ = ["id"]
    ID_FIELD_NUMBER: _ClassVar[int]
    id: _base_types_pb2.ObjectId
    def __init__(
        self, id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...
    ) -> None: ...

class UpdateOrder(_message.Message):
    __slots__ = [
        "id",
        "canceled",
        "change_items",
        "commit_items",
        "invoice_address",
        "shipping_address",
        "choose_payment",
        "payment_details",
        "paid",
    ]

    class ChangeItems(_message.Message):
        __slots__ = ["adds", "removes"]
        ADDS_FIELD_NUMBER: _ClassVar[int]
        REMOVES_FIELD_NUMBER: _ClassVar[int]
        adds: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.OrderedItem]
        removes: _containers.RepeatedCompositeFieldContainer[
            _base_types_pb2.OrderedItem
        ]
        def __init__(
            self,
            adds: _Optional[
                _Iterable[_Union[_base_types_pb2.OrderedItem, _Mapping]]
            ] = ...,
            removes: _Optional[
                _Iterable[_Union[_base_types_pb2.OrderedItem, _Mapping]]
            ] = ...,
        ) -> None: ...

    class CommitItems(_message.Message):
        __slots__ = []
        def __init__(self) -> None: ...

    class ChoosePaymentMethod(_message.Message):
        __slots__ = ["currency", "payee", "commited_at"]
        CURRENCY_FIELD_NUMBER: _ClassVar[int]
        PAYEE_FIELD_NUMBER: _ClassVar[int]
        COMMITED_AT_FIELD_NUMBER: _ClassVar[int]
        currency: _base_types_pb2.ShopCurrency
        payee: _base_types_pb2.Payee
        commited_at: _timestamp_pb2.Timestamp
        def __init__(
            self,
            currency: _Optional[_Union[_base_types_pb2.ShopCurrency, _Mapping]] = ...,
            payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ...,
            commited_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...,
        ) -> None: ...

    class Canceled(_message.Message):
        __slots__ = ["canceld_at"]
        CANCELD_AT_FIELD_NUMBER: _ClassVar[int]
        canceld_at: _timestamp_pb2.Timestamp
        def __init__(
            self,
            canceld_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...,
        ) -> None: ...

    ID_FIELD_NUMBER: _ClassVar[int]
    CANCELED_FIELD_NUMBER: _ClassVar[int]
    CHANGE_ITEMS_FIELD_NUMBER: _ClassVar[int]
    COMMIT_ITEMS_FIELD_NUMBER: _ClassVar[int]
    INVOICE_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    SHIPPING_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    CHOOSE_PAYMENT_FIELD_NUMBER: _ClassVar[int]
    PAYMENT_DETAILS_FIELD_NUMBER: _ClassVar[int]
    PAID_FIELD_NUMBER: _ClassVar[int]
    id: _base_types_pb2.ObjectId
    canceled: UpdateOrder.Canceled
    change_items: UpdateOrder.ChangeItems
    commit_items: UpdateOrder.CommitItems
    invoice_address: _base_types_pb2.AddressDetails
    shipping_address: _base_types_pb2.AddressDetails
    choose_payment: UpdateOrder.ChoosePaymentMethod
    payment_details: _base_types_pb2.PaymentDetails
    paid: _base_types_pb2.OrderPaid
    def __init__(
        self,
        id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        canceled: _Optional[_Union[UpdateOrder.Canceled, _Mapping]] = ...,
        change_items: _Optional[_Union[UpdateOrder.ChangeItems, _Mapping]] = ...,
        commit_items: _Optional[_Union[UpdateOrder.CommitItems, _Mapping]] = ...,
        invoice_address: _Optional[
            _Union[_base_types_pb2.AddressDetails, _Mapping]
        ] = ...,
        shipping_address: _Optional[
            _Union[_base_types_pb2.AddressDetails, _Mapping]
        ] = ...,
        choose_payment: _Optional[
            _Union[UpdateOrder.ChoosePaymentMethod, _Mapping]
        ] = ...,
        payment_details: _Optional[
            _Union[_base_types_pb2.PaymentDetails, _Mapping]
        ] = ...,
        paid: _Optional[_Union[_base_types_pb2.OrderPaid, _Mapping]] = ...,
    ) -> None: ...

class ShopEvent(_message.Message):
    __slots__ = [
        "nonce",
        "shop_id",
        "timestamp",
        "manifest",
        "update_manifest",
        "account",
        "listing",
        "update_listing",
        "change_inventory",
        "tag",
        "update_tag",
        "create_order",
        "update_order",
    ]
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
    def __init__(
        self,
        nonce: _Optional[int] = ...,
        shop_id: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ...,
        timestamp: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...,
        manifest: _Optional[_Union[Manifest, _Mapping]] = ...,
        update_manifest: _Optional[_Union[UpdateManifest, _Mapping]] = ...,
        account: _Optional[_Union[Account, _Mapping]] = ...,
        listing: _Optional[_Union[Listing, _Mapping]] = ...,
        update_listing: _Optional[_Union[UpdateListing, _Mapping]] = ...,
        change_inventory: _Optional[_Union[ChangeInventory, _Mapping]] = ...,
        tag: _Optional[_Union[Tag, _Mapping]] = ...,
        update_tag: _Optional[_Union[UpdateTag, _Mapping]] = ...,
        create_order: _Optional[_Union[CreateOrder, _Mapping]] = ...,
        update_order: _Optional[_Union[UpdateOrder, _Mapping]] = ...,
    ) -> None: ...
