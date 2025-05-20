# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket import base_types_pb2 as _base_types_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
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

class Order(_message.Message):
    __slots__ = (
        "id",
        "items",
        "shipping_status",
        "canceled_at",
        "commited_at",
        "invoice_address",
        "shipping_address",
        "address_updated_at",
        "chosen_payee",
        "chosen_currency",
        "payment_details",
        "payment_details_created_at",
        "payment_transactions",
    )
    ID_FIELD_NUMBER: _ClassVar[int]
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    SHIPPING_STATUS_FIELD_NUMBER: _ClassVar[int]
    CANCELED_AT_FIELD_NUMBER: _ClassVar[int]
    COMMITED_AT_FIELD_NUMBER: _ClassVar[int]
    INVOICE_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    SHIPPING_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    ADDRESS_UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    CHOSEN_PAYEE_FIELD_NUMBER: _ClassVar[int]
    CHOSEN_CURRENCY_FIELD_NUMBER: _ClassVar[int]
    PAYMENT_DETAILS_FIELD_NUMBER: _ClassVar[int]
    PAYMENT_DETAILS_CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    PAYMENT_TRANSACTIONS_FIELD_NUMBER: _ClassVar[int]
    id: _base_types_pb2.ObjectId
    items: _containers.RepeatedCompositeFieldContainer[_base_types_pb2.OrderedItem]
    shipping_status: str
    canceled_at: _timestamp_pb2.Timestamp
    commited_at: _timestamp_pb2.Timestamp
    invoice_address: _base_types_pb2.AddressDetails
    shipping_address: _base_types_pb2.AddressDetails
    address_updated_at: _timestamp_pb2.Timestamp
    chosen_payee: _base_types_pb2.Payee
    chosen_currency: _base_types_pb2.ShopCurrency
    payment_details: _base_types_pb2.PaymentDetails
    payment_details_created_at: _timestamp_pb2.Timestamp
    payment_transactions: _containers.RepeatedCompositeFieldContainer[
        _base_types_pb2.OrderTransaction
    ]
    def __init__(
        self,
        id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        items: _Optional[
            _Iterable[_Union[_base_types_pb2.OrderedItem, _Mapping]]
        ] = ...,
        shipping_status: _Optional[str] = ...,
        canceled_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...,
        commited_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...,
        invoice_address: _Optional[
            _Union[_base_types_pb2.AddressDetails, _Mapping]
        ] = ...,
        shipping_address: _Optional[
            _Union[_base_types_pb2.AddressDetails, _Mapping]
        ] = ...,
        address_updated_at: _Optional[_Union[_timestamp_pb2.Timestamp, _Mapping]] = ...,
        chosen_payee: _Optional[_Union[_base_types_pb2.Payee, _Mapping]] = ...,
        chosen_currency: _Optional[
            _Union[_base_types_pb2.ShopCurrency, _Mapping]
        ] = ...,
        payment_details: _Optional[
            _Union[_base_types_pb2.PaymentDetails, _Mapping]
        ] = ...,
        payment_details_created_at: _Optional[
            _Union[_timestamp_pb2.Timestamp, _Mapping]
        ] = ...,
        payment_transactions: _Optional[
            _Iterable[_Union[_base_types_pb2.OrderTransaction, _Mapping]]
        ] = ...,
    ) -> None: ...
