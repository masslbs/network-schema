# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

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

class ListingViewState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    LISTING_VIEW_STATE_UNSPECIFIED: _ClassVar[ListingViewState]
    LISTING_VIEW_STATE_PUBLISHED: _ClassVar[ListingViewState]
    LISTING_VIEW_STATE_DELETED: _ClassVar[ListingViewState]

LISTING_VIEW_STATE_UNSPECIFIED: ListingViewState
LISTING_VIEW_STATE_PUBLISHED: ListingViewState
LISTING_VIEW_STATE_DELETED: ListingViewState

class RequestId(_message.Message):
    __slots__ = ["raw"]
    RAW_FIELD_NUMBER: _ClassVar[int]
    raw: int
    def __init__(self, raw: _Optional[int] = ...) -> None: ...

class ObjectId(_message.Message):
    __slots__ = ["raw"]
    RAW_FIELD_NUMBER: _ClassVar[int]
    raw: bytes
    def __init__(self, raw: _Optional[bytes] = ...) -> None: ...

class Signature(_message.Message):
    __slots__ = ["raw"]
    RAW_FIELD_NUMBER: _ClassVar[int]
    raw: bytes
    def __init__(self, raw: _Optional[bytes] = ...) -> None: ...

class PublicKey(_message.Message):
    __slots__ = ["raw"]
    RAW_FIELD_NUMBER: _ClassVar[int]
    raw: bytes
    def __init__(self, raw: _Optional[bytes] = ...) -> None: ...

class Hash(_message.Message):
    __slots__ = ["raw"]
    RAW_FIELD_NUMBER: _ClassVar[int]
    raw: bytes
    def __init__(self, raw: _Optional[bytes] = ...) -> None: ...

class EthereumAddress(_message.Message):
    __slots__ = ["raw"]
    RAW_FIELD_NUMBER: _ClassVar[int]
    raw: bytes
    def __init__(self, raw: _Optional[bytes] = ...) -> None: ...

class IPFSAddress(_message.Message):
    __slots__ = ["cid"]
    CID_FIELD_NUMBER: _ClassVar[int]
    cid: str
    def __init__(self, cid: _Optional[str] = ...) -> None: ...

class Uint256(_message.Message):
    __slots__ = ["raw"]
    RAW_FIELD_NUMBER: _ClassVar[int]
    raw: bytes
    def __init__(self, raw: _Optional[bytes] = ...) -> None: ...

class ShopCurrency(_message.Message):
    __slots__ = ["chain_id", "address"]
    CHAIN_ID_FIELD_NUMBER: _ClassVar[int]
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    chain_id: int
    address: EthereumAddress
    def __init__(
        self,
        chain_id: _Optional[int] = ...,
        address: _Optional[_Union[EthereumAddress, _Mapping]] = ...,
    ) -> None: ...

class Payee(_message.Message):
    __slots__ = ["name", "address", "chain_id", "call_as_contract"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ADDRESS_FIELD_NUMBER: _ClassVar[int]
    CHAIN_ID_FIELD_NUMBER: _ClassVar[int]
    CALL_AS_CONTRACT_FIELD_NUMBER: _ClassVar[int]
    name: str
    address: EthereumAddress
    chain_id: int
    call_as_contract: bool
    def __init__(
        self,
        name: _Optional[str] = ...,
        address: _Optional[_Union[EthereumAddress, _Mapping]] = ...,
        chain_id: _Optional[int] = ...,
        call_as_contract: bool = ...,
    ) -> None: ...

class ShippingRegion(_message.Message):
    __slots__ = ["name", "country", "postal_code", "city", "order_price_modifier_ids"]
    NAME_FIELD_NUMBER: _ClassVar[int]
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    POSTAL_CODE_FIELD_NUMBER: _ClassVar[int]
    CITY_FIELD_NUMBER: _ClassVar[int]
    ORDER_PRICE_MODIFIER_IDS_FIELD_NUMBER: _ClassVar[int]
    name: str
    country: str
    postal_code: str
    city: str
    order_price_modifier_ids: _containers.RepeatedCompositeFieldContainer[ObjectId]
    def __init__(
        self,
        name: _Optional[str] = ...,
        country: _Optional[str] = ...,
        postal_code: _Optional[str] = ...,
        city: _Optional[str] = ...,
        order_price_modifier_ids: _Optional[
            _Iterable[_Union[ObjectId, _Mapping]]
        ] = ...,
    ) -> None: ...

class OrderPriceModifier(_message.Message):
    __slots__ = ["id", "title", "percentage", "absolute"]
    ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    PERCENTAGE_FIELD_NUMBER: _ClassVar[int]
    ABSOLUTE_FIELD_NUMBER: _ClassVar[int]
    id: ObjectId
    title: str
    percentage: Uint256
    absolute: PlusMinus
    def __init__(
        self,
        id: _Optional[_Union[ObjectId, _Mapping]] = ...,
        title: _Optional[str] = ...,
        percentage: _Optional[_Union[Uint256, _Mapping]] = ...,
        absolute: _Optional[_Union[PlusMinus, _Mapping]] = ...,
    ) -> None: ...

class PlusMinus(_message.Message):
    __slots__ = ["plus_sign", "diff"]
    PLUS_SIGN_FIELD_NUMBER: _ClassVar[int]
    DIFF_FIELD_NUMBER: _ClassVar[int]
    plus_sign: bool
    diff: Uint256
    def __init__(
        self, plus_sign: bool = ..., diff: _Optional[_Union[Uint256, _Mapping]] = ...
    ) -> None: ...

class ListingMetadata(_message.Message):
    __slots__ = ["title", "description", "images"]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    IMAGES_FIELD_NUMBER: _ClassVar[int]
    title: str
    description: str
    images: _containers.RepeatedScalarFieldContainer[str]
    def __init__(
        self,
        title: _Optional[str] = ...,
        description: _Optional[str] = ...,
        images: _Optional[_Iterable[str]] = ...,
    ) -> None: ...

class ListingOption(_message.Message):
    __slots__ = ["id", "title", "variations"]
    ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    VARIATIONS_FIELD_NUMBER: _ClassVar[int]
    id: ObjectId
    title: str
    variations: _containers.RepeatedCompositeFieldContainer[ListingVariation]
    def __init__(
        self,
        id: _Optional[_Union[ObjectId, _Mapping]] = ...,
        title: _Optional[str] = ...,
        variations: _Optional[_Iterable[_Union[ListingVariation, _Mapping]]] = ...,
    ) -> None: ...

class ListingVariation(_message.Message):
    __slots__ = ["id", "variation_info", "diff"]
    ID_FIELD_NUMBER: _ClassVar[int]
    VARIATION_INFO_FIELD_NUMBER: _ClassVar[int]
    DIFF_FIELD_NUMBER: _ClassVar[int]
    id: ObjectId
    variation_info: ListingMetadata
    diff: PlusMinus
    def __init__(
        self,
        id: _Optional[_Union[ObjectId, _Mapping]] = ...,
        variation_info: _Optional[_Union[ListingMetadata, _Mapping]] = ...,
        diff: _Optional[_Union[PlusMinus, _Mapping]] = ...,
    ) -> None: ...

class ListingStockStatus(_message.Message):
    __slots__ = ["variation_ids", "in_stock", "expected_in_stock_by"]
    VARIATION_IDS_FIELD_NUMBER: _ClassVar[int]
    IN_STOCK_FIELD_NUMBER: _ClassVar[int]
    EXPECTED_IN_STOCK_BY_FIELD_NUMBER: _ClassVar[int]
    variation_ids: _containers.RepeatedCompositeFieldContainer[ObjectId]
    in_stock: bool
    expected_in_stock_by: _timestamp_pb2.Timestamp
    def __init__(
        self,
        variation_ids: _Optional[_Iterable[_Union[ObjectId, _Mapping]]] = ...,
        in_stock: bool = ...,
        expected_in_stock_by: _Optional[
            _Union[_timestamp_pb2.Timestamp, _Mapping]
        ] = ...,
    ) -> None: ...

class AddressDetails(_message.Message):
    __slots__ = [
        "name",
        "address1",
        "address2",
        "city",
        "postal_code",
        "country",
        "email_address",
        "phone_number",
    ]
    NAME_FIELD_NUMBER: _ClassVar[int]
    ADDRESS1_FIELD_NUMBER: _ClassVar[int]
    ADDRESS2_FIELD_NUMBER: _ClassVar[int]
    CITY_FIELD_NUMBER: _ClassVar[int]
    POSTAL_CODE_FIELD_NUMBER: _ClassVar[int]
    COUNTRY_FIELD_NUMBER: _ClassVar[int]
    EMAIL_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    PHONE_NUMBER_FIELD_NUMBER: _ClassVar[int]
    name: str
    address1: str
    address2: str
    city: str
    postal_code: str
    country: str
    email_address: str
    phone_number: str
    def __init__(
        self,
        name: _Optional[str] = ...,
        address1: _Optional[str] = ...,
        address2: _Optional[str] = ...,
        city: _Optional[str] = ...,
        postal_code: _Optional[str] = ...,
        country: _Optional[str] = ...,
        email_address: _Optional[str] = ...,
        phone_number: _Optional[str] = ...,
    ) -> None: ...

class PaymentDetails(_message.Message):
    __slots__ = [
        "payment_id",
        "total",
        "listing_hashes",
        "ttl",
        "shop_signature",
        "shipping_region",
    ]
    PAYMENT_ID_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FIELD_NUMBER: _ClassVar[int]
    LISTING_HASHES_FIELD_NUMBER: _ClassVar[int]
    TTL_FIELD_NUMBER: _ClassVar[int]
    SHOP_SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    SHIPPING_REGION_FIELD_NUMBER: _ClassVar[int]
    payment_id: Hash
    total: Uint256
    listing_hashes: _containers.RepeatedCompositeFieldContainer[IPFSAddress]
    ttl: str
    shop_signature: Signature
    shipping_region: ShippingRegion
    def __init__(
        self,
        payment_id: _Optional[_Union[Hash, _Mapping]] = ...,
        total: _Optional[_Union[Uint256, _Mapping]] = ...,
        listing_hashes: _Optional[_Iterable[_Union[IPFSAddress, _Mapping]]] = ...,
        ttl: _Optional[str] = ...,
        shop_signature: _Optional[_Union[Signature, _Mapping]] = ...,
        shipping_region: _Optional[_Union[ShippingRegion, _Mapping]] = ...,
    ) -> None: ...

class OrderPaid(_message.Message):
    __slots__ = ["tx_hash", "block_hash"]
    TX_HASH_FIELD_NUMBER: _ClassVar[int]
    BLOCK_HASH_FIELD_NUMBER: _ClassVar[int]
    tx_hash: Hash
    block_hash: Hash
    def __init__(
        self,
        tx_hash: _Optional[_Union[Hash, _Mapping]] = ...,
        block_hash: _Optional[_Union[Hash, _Mapping]] = ...,
    ) -> None: ...

class OrderedItem(_message.Message):
    __slots__ = ["listing_id", "variation_ids", "quantity"]
    LISTING_ID_FIELD_NUMBER: _ClassVar[int]
    VARIATION_IDS_FIELD_NUMBER: _ClassVar[int]
    QUANTITY_FIELD_NUMBER: _ClassVar[int]
    listing_id: ObjectId
    variation_ids: _containers.RepeatedCompositeFieldContainer[ObjectId]
    quantity: int
    def __init__(
        self,
        listing_id: _Optional[_Union[ObjectId, _Mapping]] = ...,
        variation_ids: _Optional[_Iterable[_Union[ObjectId, _Mapping]]] = ...,
        quantity: _Optional[int] = ...,
    ) -> None: ...
