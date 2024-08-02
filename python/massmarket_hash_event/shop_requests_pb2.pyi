# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from massmarket_hash_event import error_pb2 as _error_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class GetBlobUploadURLRequest(_message.Message):
    __slots__ = ["request_id"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    request_id: _base_types_pb2.RequestId
    def __init__(self, request_id: _Optional[_Union[_base_types_pb2.RequestId, _Mapping]] = ...) -> None: ...

class GetBlobUploadURLResponse(_message.Message):
    __slots__ = ["request_id", "error", "url"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    request_id: _base_types_pb2.RequestId
    error: _error_pb2.Error
    url: str
    def __init__(self, request_id: _Optional[_Union[_base_types_pb2.RequestId, _Mapping]] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ..., url: _Optional[str] = ...) -> None: ...
