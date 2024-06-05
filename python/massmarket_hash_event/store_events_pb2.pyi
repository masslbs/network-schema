# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class StoreManifest(_message.Message):
    __slots__ = ["event_id", "store_token_id", "domain", "published_tag_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    STORE_TOKEN_ID_FIELD_NUMBER: _ClassVar[int]
    DOMAIN_FIELD_NUMBER: _ClassVar[int]
    PUBLISHED_TAG_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    store_token_id: bytes
    domain: str
    published_tag_id: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., store_token_id: _Optional[bytes] = ..., domain: _Optional[str] = ..., published_tag_id: _Optional[bytes] = ...) -> None: ...

class UpdateStoreManifest(_message.Message):
    __slots__ = ["event_id", "field", "string", "tag_id", "erc20_addr"]
    class ManifestField(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        MANIFEST_FIELD_UNSPECIFIED: _ClassVar[UpdateStoreManifest.ManifestField]
        MANIFEST_FIELD_DOMAIN: _ClassVar[UpdateStoreManifest.ManifestField]
        MANIFEST_FIELD_PUBLISHED_TAG: _ClassVar[UpdateStoreManifest.ManifestField]
        MANIFEST_FIELD_ADD_ERC20: _ClassVar[UpdateStoreManifest.ManifestField]
        MANIFEST_FIELD_REMOVE_ERC20: _ClassVar[UpdateStoreManifest.ManifestField]
    MANIFEST_FIELD_UNSPECIFIED: UpdateStoreManifest.ManifestField
    MANIFEST_FIELD_DOMAIN: UpdateStoreManifest.ManifestField
    MANIFEST_FIELD_PUBLISHED_TAG: UpdateStoreManifest.ManifestField
    MANIFEST_FIELD_ADD_ERC20: UpdateStoreManifest.ManifestField
    MANIFEST_FIELD_REMOVE_ERC20: UpdateStoreManifest.ManifestField
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    FIELD_FIELD_NUMBER: _ClassVar[int]
    STRING_FIELD_NUMBER: _ClassVar[int]
    TAG_ID_FIELD_NUMBER: _ClassVar[int]
    ERC20_ADDR_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    field: UpdateStoreManifest.ManifestField
    string: str
    tag_id: bytes
    erc20_addr: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., field: _Optional[_Union[UpdateStoreManifest.ManifestField, str]] = ..., string: _Optional[str] = ..., tag_id: _Optional[bytes] = ..., erc20_addr: _Optional[bytes] = ...) -> None: ...

class CreateItem(_message.Message):
    __slots__ = ["event_id", "price", "metadata"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    PRICE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    price: str
    metadata: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., price: _Optional[str] = ..., metadata: _Optional[bytes] = ...) -> None: ...

class UpdateItem(_message.Message):
    __slots__ = ["event_id", "item_id", "field", "price", "metadata"]
    class ItemField(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        ITEM_FIELD_UNSPECIFIED: _ClassVar[UpdateItem.ItemField]
        ITEM_FIELD_PRICE: _ClassVar[UpdateItem.ItemField]
        ITEM_FIELD_METADATA: _ClassVar[UpdateItem.ItemField]
    ITEM_FIELD_UNSPECIFIED: UpdateItem.ItemField
    ITEM_FIELD_PRICE: UpdateItem.ItemField
    ITEM_FIELD_METADATA: UpdateItem.ItemField
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    ITEM_ID_FIELD_NUMBER: _ClassVar[int]
    FIELD_FIELD_NUMBER: _ClassVar[int]
    PRICE_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    item_id: bytes
    field: UpdateItem.ItemField
    price: str
    metadata: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., item_id: _Optional[bytes] = ..., field: _Optional[_Union[UpdateItem.ItemField, str]] = ..., price: _Optional[str] = ..., metadata: _Optional[bytes] = ...) -> None: ...

class CreateTag(_message.Message):
    __slots__ = ["event_id", "name"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    name: str
    def __init__(self, event_id: _Optional[bytes] = ..., name: _Optional[str] = ...) -> None: ...

class UpdateTag(_message.Message):
    __slots__ = ["event_id", "tag_id", "action", "item_id", "new_name", "delete"]
    class TagAction(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        TAG_ACTION_UNSPECIFIED: _ClassVar[UpdateTag.TagAction]
        TAG_ACTION_ADD_ITEM: _ClassVar[UpdateTag.TagAction]
        TAG_ACTION_REMOVE_ITEM: _ClassVar[UpdateTag.TagAction]
        TAG_ACTION_RENAME: _ClassVar[UpdateTag.TagAction]
        TAG_ACTION_DELETE_TAG: _ClassVar[UpdateTag.TagAction]
    TAG_ACTION_UNSPECIFIED: UpdateTag.TagAction
    TAG_ACTION_ADD_ITEM: UpdateTag.TagAction
    TAG_ACTION_REMOVE_ITEM: UpdateTag.TagAction
    TAG_ACTION_RENAME: UpdateTag.TagAction
    TAG_ACTION_DELETE_TAG: UpdateTag.TagAction
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    TAG_ID_FIELD_NUMBER: _ClassVar[int]
    ACTION_FIELD_NUMBER: _ClassVar[int]
    ITEM_ID_FIELD_NUMBER: _ClassVar[int]
    NEW_NAME_FIELD_NUMBER: _ClassVar[int]
    DELETE_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    tag_id: bytes
    action: UpdateTag.TagAction
    item_id: bytes
    new_name: str
    delete: bool
    def __init__(self, event_id: _Optional[bytes] = ..., tag_id: _Optional[bytes] = ..., action: _Optional[_Union[UpdateTag.TagAction, str]] = ..., item_id: _Optional[bytes] = ..., new_name: _Optional[str] = ..., delete: bool = ...) -> None: ...

class ChangeStock(_message.Message):
    __slots__ = ["event_id", "item_ids", "diffs", "order_id", "tx_hash"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    ITEM_IDS_FIELD_NUMBER: _ClassVar[int]
    DIFFS_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    TX_HASH_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    item_ids: _containers.RepeatedScalarFieldContainer[bytes]
    diffs: _containers.RepeatedScalarFieldContainer[int]
    order_id: bytes
    tx_hash: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., item_ids: _Optional[_Iterable[bytes]] = ..., diffs: _Optional[_Iterable[int]] = ..., order_id: _Optional[bytes] = ..., tx_hash: _Optional[bytes] = ...) -> None: ...

class NewKeyCard(_message.Message):
    __slots__ = ["event_id", "user_wallet_addr", "card_public_key", "is_guest"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_WALLET_ADDR_FIELD_NUMBER: _ClassVar[int]
    CARD_PUBLIC_KEY_FIELD_NUMBER: _ClassVar[int]
    IS_GUEST_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    user_wallet_addr: bytes
    card_public_key: bytes
    is_guest: bool
    def __init__(self, event_id: _Optional[bytes] = ..., user_wallet_addr: _Optional[bytes] = ..., card_public_key: _Optional[bytes] = ..., is_guest: bool = ...) -> None: ...

class CreateOrder(_message.Message):
    __slots__ = ["event_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    def __init__(self, event_id: _Optional[bytes] = ...) -> None: ...

class UpdateOrder(_message.Message):
    __slots__ = ["event_id", "order_id", "change_items", "items_finalized", "order_canceled"]
    class ChangeItems(_message.Message):
        __slots__ = ["item_id", "quantity"]
        ITEM_ID_FIELD_NUMBER: _ClassVar[int]
        QUANTITY_FIELD_NUMBER: _ClassVar[int]
        item_id: bytes
        quantity: int
        def __init__(self, item_id: _Optional[bytes] = ..., quantity: _Optional[int] = ...) -> None: ...
    class ItemsFinalized(_message.Message):
        __slots__ = ["payment_id", "sub_total", "sales_tax", "total", "ttl", "order_hash", "currency_addr", "total_in_crypto", "payee_addr", "is_payment_endpoint", "shop_signature"]
        PAYMENT_ID_FIELD_NUMBER: _ClassVar[int]
        SUB_TOTAL_FIELD_NUMBER: _ClassVar[int]
        SALES_TAX_FIELD_NUMBER: _ClassVar[int]
        TOTAL_FIELD_NUMBER: _ClassVar[int]
        TTL_FIELD_NUMBER: _ClassVar[int]
        ORDER_HASH_FIELD_NUMBER: _ClassVar[int]
        CURRENCY_ADDR_FIELD_NUMBER: _ClassVar[int]
        TOTAL_IN_CRYPTO_FIELD_NUMBER: _ClassVar[int]
        PAYEE_ADDR_FIELD_NUMBER: _ClassVar[int]
        IS_PAYMENT_ENDPOINT_FIELD_NUMBER: _ClassVar[int]
        SHOP_SIGNATURE_FIELD_NUMBER: _ClassVar[int]
        payment_id: bytes
        sub_total: str
        sales_tax: str
        total: str
        ttl: str
        order_hash: bytes
        currency_addr: bytes
        total_in_crypto: str
        payee_addr: bytes
        is_payment_endpoint: bool
        shop_signature: bytes
        def __init__(self, payment_id: _Optional[bytes] = ..., sub_total: _Optional[str] = ..., sales_tax: _Optional[str] = ..., total: _Optional[str] = ..., ttl: _Optional[str] = ..., order_hash: _Optional[bytes] = ..., currency_addr: _Optional[bytes] = ..., total_in_crypto: _Optional[str] = ..., payee_addr: _Optional[bytes] = ..., is_payment_endpoint: bool = ..., shop_signature: _Optional[bytes] = ...) -> None: ...
    class OrderCanceled(_message.Message):
        __slots__ = ["timestamp"]
        TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
        timestamp: int
        def __init__(self, timestamp: _Optional[int] = ...) -> None: ...
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    CHANGE_ITEMS_FIELD_NUMBER: _ClassVar[int]
    ITEMS_FINALIZED_FIELD_NUMBER: _ClassVar[int]
    ORDER_CANCELED_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    order_id: bytes
    change_items: UpdateOrder.ChangeItems
    items_finalized: UpdateOrder.ItemsFinalized
    order_canceled: UpdateOrder.OrderCanceled
    def __init__(self, event_id: _Optional[bytes] = ..., order_id: _Optional[bytes] = ..., change_items: _Optional[_Union[UpdateOrder.ChangeItems, _Mapping]] = ..., items_finalized: _Optional[_Union[UpdateOrder.ItemsFinalized, _Mapping]] = ..., order_canceled: _Optional[_Union[UpdateOrder.OrderCanceled, _Mapping]] = ...) -> None: ...

class StoreEvent(_message.Message):
    __slots__ = ["signature", "store_manifest", "update_store_manifest", "create_item", "update_item", "create_tag", "update_tag", "create_order", "update_order", "change_stock", "new_key_card"]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    STORE_MANIFEST_FIELD_NUMBER: _ClassVar[int]
    UPDATE_STORE_MANIFEST_FIELD_NUMBER: _ClassVar[int]
    CREATE_ITEM_FIELD_NUMBER: _ClassVar[int]
    UPDATE_ITEM_FIELD_NUMBER: _ClassVar[int]
    CREATE_TAG_FIELD_NUMBER: _ClassVar[int]
    UPDATE_TAG_FIELD_NUMBER: _ClassVar[int]
    CREATE_ORDER_FIELD_NUMBER: _ClassVar[int]
    UPDATE_ORDER_FIELD_NUMBER: _ClassVar[int]
    CHANGE_STOCK_FIELD_NUMBER: _ClassVar[int]
    NEW_KEY_CARD_FIELD_NUMBER: _ClassVar[int]
    signature: bytes
    store_manifest: StoreManifest
    update_store_manifest: UpdateStoreManifest
    create_item: CreateItem
    update_item: UpdateItem
    create_tag: CreateTag
    update_tag: UpdateTag
    create_order: CreateOrder
    update_order: UpdateOrder
    change_stock: ChangeStock
    new_key_card: NewKeyCard
    def __init__(self, signature: _Optional[bytes] = ..., store_manifest: _Optional[_Union[StoreManifest, _Mapping]] = ..., update_store_manifest: _Optional[_Union[UpdateStoreManifest, _Mapping]] = ..., create_item: _Optional[_Union[CreateItem, _Mapping]] = ..., update_item: _Optional[_Union[UpdateItem, _Mapping]] = ..., create_tag: _Optional[_Union[CreateTag, _Mapping]] = ..., update_tag: _Optional[_Union[UpdateTag, _Mapping]] = ..., create_order: _Optional[_Union[CreateOrder, _Mapping]] = ..., update_order: _Optional[_Union[UpdateOrder, _Mapping]] = ..., change_stock: _Optional[_Union[ChangeStock, _Mapping]] = ..., new_key_card: _Optional[_Union[NewKeyCard, _Mapping]] = ...) -> None: ...
