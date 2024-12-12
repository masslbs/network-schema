# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from google.protobuf import any_pb2 as _any_pb2
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

class SignedEvent(_message.Message):
    __slots__ = ["event", "signature"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    event: _any_pb2.Any
    signature: _base_types_pb2.Signature
    def __init__(
        self,
        event: _Optional[_Union[_any_pb2.Any, _Mapping]] = ...,
        signature: _Optional[_Union[_base_types_pb2.Signature, _Mapping]] = ...,
    ) -> None: ...

class EventWriteRequest(_message.Message):
    __slots__ = ["events"]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[SignedEvent]
    def __init__(
        self, events: _Optional[_Iterable[_Union[SignedEvent, _Mapping]]] = ...
    ) -> None: ...

class SyncStatusRequest(_message.Message):
    __slots__ = ["unpushed_events"]
    UNPUSHED_EVENTS_FIELD_NUMBER: _ClassVar[int]
    unpushed_events: int
    def __init__(self, unpushed_events: _Optional[int] = ...) -> None: ...

class PingRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
