# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket import base_types_pb2 as _base_types_pb2
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

class ObjectType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
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
    __slots__ = ("start_shop_seq_no", "shop_id", "filters")

    class Filter(_message.Message):
        __slots__ = ("object_type", "object_id")
        OBJECT_TYPE_FIELD_NUMBER: _ClassVar[int]
        OBJECT_ID_FIELD_NUMBER: _ClassVar[int]
        object_type: ObjectType
        object_id: _base_types_pb2.ObjectId
        def __init__(
            self,
            object_type: _Optional[_Union[ObjectType, str]] = ...,
            object_id: _Optional[_Union[_base_types_pb2.ObjectId, _Mapping]] = ...,
        ) -> None: ...

    START_SHOP_SEQ_NO_FIELD_NUMBER: _ClassVar[int]
    SHOP_ID_FIELD_NUMBER: _ClassVar[int]
    FILTERS_FIELD_NUMBER: _ClassVar[int]
    start_shop_seq_no: int
    shop_id: _base_types_pb2.Uint256
    filters: _containers.RepeatedCompositeFieldContainer[SubscriptionRequest.Filter]
    def __init__(
        self,
        start_shop_seq_no: _Optional[int] = ...,
        shop_id: _Optional[_Union[_base_types_pb2.Uint256, _Mapping]] = ...,
        filters: _Optional[
            _Iterable[_Union[SubscriptionRequest.Filter, _Mapping]]
        ] = ...,
    ) -> None: ...

class SubscriptionPushRequest(_message.Message):
    __slots__ = ("subscription_id", "sets")

    class SequencedPartialPatchSet(_message.Message):
        __slots__ = (
            "shop_seq_no",
            "patch_leaf_index",
            "header",
            "signature",
            "patches",
            "proofs",
        )
        SHOP_SEQ_NO_FIELD_NUMBER: _ClassVar[int]
        PATCH_LEAF_INDEX_FIELD_NUMBER: _ClassVar[int]
        HEADER_FIELD_NUMBER: _ClassVar[int]
        SIGNATURE_FIELD_NUMBER: _ClassVar[int]
        PATCHES_FIELD_NUMBER: _ClassVar[int]
        PROOFS_FIELD_NUMBER: _ClassVar[int]
        shop_seq_no: int
        patch_leaf_index: int
        header: bytes
        signature: bytes
        patches: _containers.RepeatedScalarFieldContainer[bytes]
        proofs: _containers.RepeatedScalarFieldContainer[bytes]
        def __init__(
            self,
            shop_seq_no: _Optional[int] = ...,
            patch_leaf_index: _Optional[int] = ...,
            header: _Optional[bytes] = ...,
            signature: _Optional[bytes] = ...,
            patches: _Optional[_Iterable[bytes]] = ...,
            proofs: _Optional[_Iterable[bytes]] = ...,
        ) -> None: ...

    SUBSCRIPTION_ID_FIELD_NUMBER: _ClassVar[int]
    SETS_FIELD_NUMBER: _ClassVar[int]
    subscription_id: bytes
    sets: _containers.RepeatedCompositeFieldContainer[
        SubscriptionPushRequest.SequencedPartialPatchSet
    ]
    def __init__(
        self,
        subscription_id: _Optional[bytes] = ...,
        sets: _Optional[
            _Iterable[
                _Union[SubscriptionPushRequest.SequencedPartialPatchSet, _Mapping]
            ]
        ] = ...,
    ) -> None: ...

class SubscriptionCancelRequest(_message.Message):
    __slots__ = ("subscription_id",)
    SUBSCRIPTION_ID_FIELD_NUMBER: _ClassVar[int]
    subscription_id: bytes
    def __init__(self, subscription_id: _Optional[bytes] = ...) -> None: ...
