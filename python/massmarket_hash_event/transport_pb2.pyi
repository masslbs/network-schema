# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from google.protobuf import any_pb2 as _any_pb2
from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from massmarket_hash_event import error_pb2 as _error_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SignedEvent(_message.Message):
    __slots__ = ["event", "signature"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    event: _any_pb2.Any
    signature: _base_types_pb2.Signature
    def __init__(self, event: _Optional[_Union[_any_pb2.Any, _Mapping]] = ..., signature: _Optional[_Union[_base_types_pb2.Signature, _Mapping]] = ...) -> None: ...

class EventWriteRequest(_message.Message):
    __slots__ = ["event"]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    event: SignedEvent
    def __init__(self, event: _Optional[_Union[SignedEvent, _Mapping]] = ...) -> None: ...

class EventWriteResponse(_message.Message):
    __slots__ = ["error", "success"]
    class Success(_message.Message):
        __slots__ = ["state_root"]
        STATE_ROOT_FIELD_NUMBER: _ClassVar[int]
        state_root: _base_types_pb2.Hash
        def __init__(self, state_root: _Optional[_Union[_base_types_pb2.Hash, _Mapping]] = ...) -> None: ...
    ERROR_FIELD_NUMBER: _ClassVar[int]
    SUCCESS_FIELD_NUMBER: _ClassVar[int]
    error: _error_pb2.Error
    success: EventWriteResponse.Success
    def __init__(self, error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ..., success: _Optional[_Union[EventWriteResponse.Success, _Mapping]] = ...) -> None: ...

class SyncStatusRequest(_message.Message):
    __slots__ = ["unpushed_events"]
    UNPUSHED_EVENTS_FIELD_NUMBER: _ClassVar[int]
    unpushed_events: int
    def __init__(self, unpushed_events: _Optional[int] = ...) -> None: ...

class SyncStatusResponse(_message.Message):
    __slots__ = ["error"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    error: _error_pb2.Error
    def __init__(self, error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...) -> None: ...

class PingRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class PingResponse(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...
