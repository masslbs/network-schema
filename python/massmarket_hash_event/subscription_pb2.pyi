from massmarket_hash_event import base_types_pb2 as _base_types_pb2
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
    __slots__ = ("subscription_id", "patches", "patch_set_meta")

    class SequencedPatch(_message.Message):
        __slots__ = ("shop_seq_no", "patch_leaf_index", "patch_data", "mmr_proof")
        SHOP_SEQ_NO_FIELD_NUMBER: _ClassVar[int]
        PATCH_LEAF_INDEX_FIELD_NUMBER: _ClassVar[int]
        PATCH_DATA_FIELD_NUMBER: _ClassVar[int]
        MMR_PROOF_FIELD_NUMBER: _ClassVar[int]
        shop_seq_no: int
        patch_leaf_index: int
        patch_data: bytes
        mmr_proof: bytes
        def __init__(
            self,
            shop_seq_no: _Optional[int] = ...,
            patch_leaf_index: _Optional[int] = ...,
            patch_data: _Optional[bytes] = ...,
            mmr_proof: _Optional[bytes] = ...,
        ) -> None: ...

    class PatchSetMetaEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: int
        value: SubscriptionPushRequest.PatchSetMeta
        def __init__(
            self,
            key: _Optional[int] = ...,
            value: _Optional[
                _Union[SubscriptionPushRequest.PatchSetMeta, _Mapping]
            ] = ...,
        ) -> None: ...

    class PatchSetMeta(_message.Message):
        __slots__ = ("header", "signature")
        HEADER_FIELD_NUMBER: _ClassVar[int]
        SIGNATURE_FIELD_NUMBER: _ClassVar[int]
        header: bytes
        signature: bytes
        def __init__(
            self, header: _Optional[bytes] = ..., signature: _Optional[bytes] = ...
        ) -> None: ...

    SUBSCRIPTION_ID_FIELD_NUMBER: _ClassVar[int]
    PATCHES_FIELD_NUMBER: _ClassVar[int]
    PATCH_SET_META_FIELD_NUMBER: _ClassVar[int]
    subscription_id: bytes
    patches: _containers.RepeatedCompositeFieldContainer[
        SubscriptionPushRequest.SequencedPatch
    ]
    patch_set_meta: _containers.MessageMap[int, SubscriptionPushRequest.PatchSetMeta]
    def __init__(
        self,
        subscription_id: _Optional[bytes] = ...,
        patches: _Optional[
            _Iterable[_Union[SubscriptionPushRequest.SequencedPatch, _Mapping]]
        ] = ...,
        patch_set_meta: _Optional[
            _Mapping[int, SubscriptionPushRequest.PatchSetMeta]
        ] = ...,
    ) -> None: ...

class SubscriptionCancelRequest(_message.Message):
    __slots__ = ("subscription_id",)
    SUBSCRIPTION_ID_FIELD_NUMBER: _ClassVar[int]
    subscription_id: bytes
    def __init__(self, subscription_id: _Optional[bytes] = ...) -> None: ...
