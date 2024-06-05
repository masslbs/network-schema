# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from google.protobuf import any_pb2 as _any_pb2
from massmarket_hash_event import error_pb2 as _error_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EventWriteRequest(_message.Message):
    __slots__ = ["request_id", "event"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    event: _any_pb2.Any
    def __init__(self, request_id: _Optional[bytes] = ..., event: _Optional[_Union[_any_pb2.Any, _Mapping]] = ...) -> None: ...

class EventWriteResponse(_message.Message):
    __slots__ = ["request_id", "error", "new_store_hash", "event_sequence_no"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    NEW_STORE_HASH_FIELD_NUMBER: _ClassVar[int]
    EVENT_SEQUENCE_NO_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: _error_pb2.Error
    new_store_hash: bytes
    event_sequence_no: int
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ..., new_store_hash: _Optional[bytes] = ..., event_sequence_no: _Optional[int] = ...) -> None: ...

class EventPushRequest(_message.Message):
    __slots__ = ["request_id", "events"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    events: _containers.RepeatedCompositeFieldContainer[_any_pb2.Any]
    def __init__(self, request_id: _Optional[bytes] = ..., events: _Optional[_Iterable[_Union[_any_pb2.Any, _Mapping]]] = ...) -> None: ...

class EventPushResponse(_message.Message):
    __slots__ = ["request_id", "error"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: _error_pb2.Error
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...) -> None: ...

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
    error: _error_pb2.Error
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...) -> None: ...

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
    error: _error_pb2.Error
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...) -> None: ...
