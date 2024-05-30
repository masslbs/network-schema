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

class UpdateManifest(_message.Message):
    __slots__ = ["event_id", "field", "string", "tag_id", "erc20_addr"]
    class ManifestField(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = []
        MANIFEST_FIELD_UNSPECIFIED: _ClassVar[UpdateManifest.ManifestField]
        MANIFEST_FIELD_DOMAIN: _ClassVar[UpdateManifest.ManifestField]
        MANIFEST_FIELD_PUBLISHED_TAG: _ClassVar[UpdateManifest.ManifestField]
        MANIFEST_FIELD_ADD_ERC20: _ClassVar[UpdateManifest.ManifestField]
        MANIFEST_FIELD_REMOVE_ERC20: _ClassVar[UpdateManifest.ManifestField]
    MANIFEST_FIELD_UNSPECIFIED: UpdateManifest.ManifestField
    MANIFEST_FIELD_DOMAIN: UpdateManifest.ManifestField
    MANIFEST_FIELD_PUBLISHED_TAG: UpdateManifest.ManifestField
    MANIFEST_FIELD_ADD_ERC20: UpdateManifest.ManifestField
    MANIFEST_FIELD_REMOVE_ERC20: UpdateManifest.ManifestField
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    FIELD_FIELD_NUMBER: _ClassVar[int]
    STRING_FIELD_NUMBER: _ClassVar[int]
    TAG_ID_FIELD_NUMBER: _ClassVar[int]
    ERC20_ADDR_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    field: UpdateManifest.ManifestField
    string: str
    tag_id: bytes
    erc20_addr: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., field: _Optional[_Union[UpdateManifest.ManifestField, str]] = ..., string: _Optional[str] = ..., tag_id: _Optional[bytes] = ..., erc20_addr: _Optional[bytes] = ...) -> None: ...

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

class AddToTag(_message.Message):
    __slots__ = ["event_id", "tag_id", "item_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    TAG_ID_FIELD_NUMBER: _ClassVar[int]
    ITEM_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    tag_id: bytes
    item_id: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., tag_id: _Optional[bytes] = ..., item_id: _Optional[bytes] = ...) -> None: ...

class RemoveFromTag(_message.Message):
    __slots__ = ["event_id", "tag_id", "item_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    TAG_ID_FIELD_NUMBER: _ClassVar[int]
    ITEM_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    tag_id: bytes
    item_id: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., tag_id: _Optional[bytes] = ..., item_id: _Optional[bytes] = ...) -> None: ...

class RenameTag(_message.Message):
    __slots__ = ["event_id", "tag_id", "name"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    TAG_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    tag_id: bytes
    name: str
    def __init__(self, event_id: _Optional[bytes] = ..., tag_id: _Optional[bytes] = ..., name: _Optional[str] = ...) -> None: ...

class DeleteTag(_message.Message):
    __slots__ = ["event_id", "tag_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    TAG_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    tag_id: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., tag_id: _Optional[bytes] = ...) -> None: ...

class ChangeStock(_message.Message):
    __slots__ = ["event_id", "item_ids", "diffs", "cart_id", "tx_hash"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    ITEM_IDS_FIELD_NUMBER: _ClassVar[int]
    DIFFS_FIELD_NUMBER: _ClassVar[int]
    CART_ID_FIELD_NUMBER: _ClassVar[int]
    TX_HASH_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    item_ids: _containers.RepeatedScalarFieldContainer[bytes]
    diffs: _containers.RepeatedScalarFieldContainer[int]
    cart_id: bytes
    tx_hash: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., item_ids: _Optional[_Iterable[bytes]] = ..., diffs: _Optional[_Iterable[int]] = ..., cart_id: _Optional[bytes] = ..., tx_hash: _Optional[bytes] = ...) -> None: ...

class NewKeyCard(_message.Message):
    __slots__ = ["event_id", "user_wallet_addr", "card_public_key"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    USER_WALLET_ADDR_FIELD_NUMBER: _ClassVar[int]
    CARD_PUBLIC_KEY_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    user_wallet_addr: bytes
    card_public_key: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., user_wallet_addr: _Optional[bytes] = ..., card_public_key: _Optional[bytes] = ...) -> None: ...

class CreateCart(_message.Message):
    __slots__ = ["event_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    def __init__(self, event_id: _Optional[bytes] = ...) -> None: ...

class ChangeCart(_message.Message):
    __slots__ = ["event_id", "cart_id", "item_id", "quantity"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    CART_ID_FIELD_NUMBER: _ClassVar[int]
    ITEM_ID_FIELD_NUMBER: _ClassVar[int]
    QUANTITY_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    cart_id: bytes
    item_id: bytes
    quantity: int
    def __init__(self, event_id: _Optional[bytes] = ..., cart_id: _Optional[bytes] = ..., item_id: _Optional[bytes] = ..., quantity: _Optional[int] = ...) -> None: ...

class CartFinalized(_message.Message):
    __slots__ = ["event_id", "cart_id", "purchase_addr", "erc20_addr", "sub_total", "sales_tax", "total", "total_in_crypto", "payment_id", "payment_ttl"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    CART_ID_FIELD_NUMBER: _ClassVar[int]
    PURCHASE_ADDR_FIELD_NUMBER: _ClassVar[int]
    ERC20_ADDR_FIELD_NUMBER: _ClassVar[int]
    SUB_TOTAL_FIELD_NUMBER: _ClassVar[int]
    SALES_TAX_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FIELD_NUMBER: _ClassVar[int]
    TOTAL_IN_CRYPTO_FIELD_NUMBER: _ClassVar[int]
    PAYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    PAYMENT_TTL_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    cart_id: bytes
    purchase_addr: bytes
    erc20_addr: bytes
    sub_total: str
    sales_tax: str
    total: str
    total_in_crypto: str
    payment_id: bytes
    payment_ttl: str
    def __init__(self, event_id: _Optional[bytes] = ..., cart_id: _Optional[bytes] = ..., purchase_addr: _Optional[bytes] = ..., erc20_addr: _Optional[bytes] = ..., sub_total: _Optional[str] = ..., sales_tax: _Optional[str] = ..., total: _Optional[str] = ..., total_in_crypto: _Optional[str] = ..., payment_id: _Optional[bytes] = ..., payment_ttl: _Optional[str] = ...) -> None: ...

class CartAbandoned(_message.Message):
    __slots__ = ["event_id", "cart_id"]
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    CART_ID_FIELD_NUMBER: _ClassVar[int]
    event_id: bytes
    cart_id: bytes
    def __init__(self, event_id: _Optional[bytes] = ..., cart_id: _Optional[bytes] = ...) -> None: ...

class Event(_message.Message):
    __slots__ = ["signature", "store_manifest", "update_manifest", "create_item", "update_item", "create_tag", "add_to_tag", "remove_from_tag", "rename_tag", "delete_tag", "create_cart", "change_cart", "cart_finalized", "cart_abandoned", "change_stock", "new_key_card"]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    STORE_MANIFEST_FIELD_NUMBER: _ClassVar[int]
    UPDATE_MANIFEST_FIELD_NUMBER: _ClassVar[int]
    CREATE_ITEM_FIELD_NUMBER: _ClassVar[int]
    UPDATE_ITEM_FIELD_NUMBER: _ClassVar[int]
    CREATE_TAG_FIELD_NUMBER: _ClassVar[int]
    ADD_TO_TAG_FIELD_NUMBER: _ClassVar[int]
    REMOVE_FROM_TAG_FIELD_NUMBER: _ClassVar[int]
    RENAME_TAG_FIELD_NUMBER: _ClassVar[int]
    DELETE_TAG_FIELD_NUMBER: _ClassVar[int]
    CREATE_CART_FIELD_NUMBER: _ClassVar[int]
    CHANGE_CART_FIELD_NUMBER: _ClassVar[int]
    CART_FINALIZED_FIELD_NUMBER: _ClassVar[int]
    CART_ABANDONED_FIELD_NUMBER: _ClassVar[int]
    CHANGE_STOCK_FIELD_NUMBER: _ClassVar[int]
    NEW_KEY_CARD_FIELD_NUMBER: _ClassVar[int]
    signature: bytes
    store_manifest: StoreManifest
    update_manifest: UpdateManifest
    create_item: CreateItem
    update_item: UpdateItem
    create_tag: CreateTag
    add_to_tag: AddToTag
    remove_from_tag: RemoveFromTag
    rename_tag: RenameTag
    delete_tag: DeleteTag
    create_cart: CreateCart
    change_cart: ChangeCart
    cart_finalized: CartFinalized
    cart_abandoned: CartAbandoned
    change_stock: ChangeStock
    new_key_card: NewKeyCard
    def __init__(self, signature: _Optional[bytes] = ..., store_manifest: _Optional[_Union[StoreManifest, _Mapping]] = ..., update_manifest: _Optional[_Union[UpdateManifest, _Mapping]] = ..., create_item: _Optional[_Union[CreateItem, _Mapping]] = ..., update_item: _Optional[_Union[UpdateItem, _Mapping]] = ..., create_tag: _Optional[_Union[CreateTag, _Mapping]] = ..., add_to_tag: _Optional[_Union[AddToTag, _Mapping]] = ..., remove_from_tag: _Optional[_Union[RemoveFromTag, _Mapping]] = ..., rename_tag: _Optional[_Union[RenameTag, _Mapping]] = ..., delete_tag: _Optional[_Union[DeleteTag, _Mapping]] = ..., create_cart: _Optional[_Union[CreateCart, _Mapping]] = ..., change_cart: _Optional[_Union[ChangeCart, _Mapping]] = ..., cart_finalized: _Optional[_Union[CartFinalized, _Mapping]] = ..., cart_abandoned: _Optional[_Union[CartAbandoned, _Mapping]] = ..., change_stock: _Optional[_Union[ChangeStock, _Mapping]] = ..., new_key_card: _Optional[_Union[NewKeyCard, _Mapping]] = ...) -> None: ...

class AuthenticateRequest(_message.Message):
    __slots__ = ["request_id", "public_key"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    PUBLIC_KEY_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    public_key: bytes
    def __init__(self, request_id: _Optional[bytes] = ..., public_key: _Optional[bytes] = ...) -> None: ...

class AuthenticateResponse(_message.Message):
    __slots__ = ["request_id", "error", "challenge"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    CHALLENGE_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: Error
    challenge: bytes
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[Error, _Mapping]] = ..., challenge: _Optional[bytes] = ...) -> None: ...

class ChallengeSolvedRequest(_message.Message):
    __slots__ = ["request_id", "signature"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    signature: bytes
    def __init__(self, request_id: _Optional[bytes] = ..., signature: _Optional[bytes] = ...) -> None: ...

class ChallengeSolvedResponse(_message.Message):
    __slots__ = ["request_id", "error"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: Error
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class CommitCartRequest(_message.Message):
    __slots__ = ["request_id", "cart_id", "erc20_addr", "escrow_addr"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    CART_ID_FIELD_NUMBER: _ClassVar[int]
    ERC20_ADDR_FIELD_NUMBER: _ClassVar[int]
    ESCROW_ADDR_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    cart_id: bytes
    erc20_addr: bytes
    escrow_addr: bytes
    def __init__(self, request_id: _Optional[bytes] = ..., cart_id: _Optional[bytes] = ..., erc20_addr: _Optional[bytes] = ..., escrow_addr: _Optional[bytes] = ...) -> None: ...

class CommitCartResponse(_message.Message):
    __slots__ = ["request_id", "error", "cart_finalized_id"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    CART_FINALIZED_ID_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: Error
    cart_finalized_id: bytes
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[Error, _Mapping]] = ..., cart_finalized_id: _Optional[bytes] = ...) -> None: ...

class GetBlobUploadURLRequest(_message.Message):
    __slots__ = ["request_id"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    def __init__(self, request_id: _Optional[bytes] = ...) -> None: ...

class GetBlobUploadURLResponse(_message.Message):
    __slots__ = ["request_id", "error", "url"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: Error
    url: str
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[Error, _Mapping]] = ..., url: _Optional[str] = ...) -> None: ...

class EventWriteRequest(_message.Message):
    __slots__ = ["request_id", "event"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    event: Event
    def __init__(self, request_id: _Optional[bytes] = ..., event: _Optional[_Union[Event, _Mapping]] = ...) -> None: ...

class EventWriteResponse(_message.Message):
    __slots__ = ["request_id", "error", "new_store_hash", "event_sequence_no"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    NEW_STORE_HASH_FIELD_NUMBER: _ClassVar[int]
    EVENT_SEQUENCE_NO_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: Error
    new_store_hash: bytes
    event_sequence_no: int
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[Error, _Mapping]] = ..., new_store_hash: _Optional[bytes] = ..., event_sequence_no: _Optional[int] = ...) -> None: ...

class SyncStatusRequest(_message.Message):
    __slots__ = ["request_id", "unpushed_events"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    UNPUSHED_EVENTS_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    unpushed_events: int
    def __init__(self, request_id: _Optional[bytes] = ..., unpushed_events: _Optional[int] = ...) -> None: ...

class SyncStatusResponse(_message.Message):
    __slots__ = ["request_id", "error"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: Error
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class EventPushRequest(_message.Message):
    __slots__ = ["request_id", "events"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    events: _containers.RepeatedCompositeFieldContainer[Event]
    def __init__(self, request_id: _Optional[bytes] = ..., events: _Optional[_Iterable[_Union[Event, _Mapping]]] = ...) -> None: ...

class EventPushResponse(_message.Message):
    __slots__ = ["request_id", "error"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: Error
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class PingRequest(_message.Message):
    __slots__ = ["request_id"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    def __init__(self, request_id: _Optional[bytes] = ...) -> None: ...

class PingResponse(_message.Message):
    __slots__ = ["request_id", "error"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: Error
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[Error, _Mapping]] = ...) -> None: ...

class Error(_message.Message):
    __slots__ = ["code", "message"]
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    code: str
    message: str
    def __init__(self, code: _Optional[str] = ..., message: _Optional[str] = ...) -> None: ...
