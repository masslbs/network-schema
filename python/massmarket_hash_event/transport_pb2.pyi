# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SignedEvent(_message.Message):
    __slots__ = ["cbor_patch_set", "signature"]
    CBOR_PATCH_SET_FIELD_NUMBER: _ClassVar[int]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    cbor_patch_set: bytes
    signature: _base_types_pb2.Signature
    def __init__(self, cbor_patch_set: _Optional[bytes] = ..., signature: _Optional[_Union[_base_types_pb2.Signature, _Mapping]] = ...) -> None: ...

class EventWriteRequest(_message.Message):
    __slots__ = ["events"]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[SignedEvent]
    def __init__(self, events: _Optional[_Iterable[_Union[SignedEvent, _Mapping]]] = ...) -> None: ...

class SyncStatusRequest(_message.Message):
    __slots__ = ["unpushed_events"]
    UNPUSHED_EVENTS_FIELD_NUMBER: _ClassVar[int]
    unpushed_events: int
    def __init__(self, unpushed_events: _Optional[int] = ...) -> None: ...

class PingRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
