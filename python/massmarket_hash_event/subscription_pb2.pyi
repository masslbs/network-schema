# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from massmarket_hash_event import error_pb2 as _error_pb2
from massmarket_hash_event import transport_pb2 as _transport_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ObjectType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    OBJECT_TYPE_UNSPECIFIED: _ClassVar[ObjectType]
    OBJECT_TYPE_LISTING: _ClassVar[ObjectType]
    OBJECT_TYPE_TAG: _ClassVar[ObjectType]
    OBJECT_TYPE_ORDER: _ClassVar[ObjectType]
    OBJECT_TYPE_ACCOUNT: _ClassVar[ObjectType]
    OBJECT_TYPE_MANIFEST: _ClassVar[ObjectType]
    OBJECT_TYPE_INVENTORY: _ClassVar[ObjectType]
OBJECT_TYPE_UNSPECIFIED: ObjectType
OBJECT_TYPE_LISTING: ObjectType
OBJECT_TYPE_TAG: ObjectType
OBJECT_TYPE_ORDER: ObjectType
OBJECT_TYPE_ACCOUNT: ObjectType
OBJECT_TYPE_MANIFEST: ObjectType
OBJECT_TYPE_INVENTORY: ObjectType

class SubscriptionRequest(_message.Message):
    __slots__ = ["start_shop_seq_no", "shop_id", "filters"]
    class Filter(_message.Message):
        __slots__ = ["object_type", "object_id"]
        OBJECT_TYPE_FIELD_NUMBER: _ClassVar[int]
        OBJECT_ID_FIELD_NUMBER: _ClassVar[int]
        object_type: ObjectType
        object_id: bytes
        def __init__(self, object_type: _Optional[_Union[ObjectType, str]] = ..., object_id: _Optional[bytes] = ...) -> None: ...
    START_SHOP_SEQ_NO_FIELD_NUMBER: _ClassVar[int]
    SHOP_ID_FIELD_NUMBER: _ClassVar[int]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    start_shop_seq_no: int
    shop_id: _base_types_pb2.Uint256
    filters: _containers.RepeatedCompositeFieldContainer[SubscriptionRequest.Filter]
    def __init__(self, start_shop_seq_no: _Optional[int] = ..., shop_id: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ..., filters: _Optional[_Iterable[_Union[SubscriptionRequest.Filter, _Mapping]]] = ...) -> None: ...

class SubscriptionResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: _error_pb2.Error
    def __init__(self, error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...) -> None: ...

class SubscriptionPush(_message.Message):
    __slots__ = ["events"]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[_transport_pb2.SignedEvent]
    def __init__(self, events: _Optional[_Iterable[_Union[_transport_pb2.SignedEvent, _Mapping]]] = ...) -> None: ...

class SubscriptionPushResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class SubscriptionCancel(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: _error_pb2.Error
    def __init__(self, error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...) -> None: ...
