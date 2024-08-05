# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import error_pb2 as _error_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import (
    ClassVar as _ClassVar,
    Mapping as _Mapping,
    Optional as _Optional,
    Union as _Union,
)

DESCRIPTOR: _descriptor.FileDescriptor

class GetBlobUploadURLRequest(_message.Message):
    __slots__ = []
    def __init__(self) -> None: ...

class GetBlobUploadURLResponse(_message.Message):
    __slots__ = ["error", "url"]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    error: _error_pb2.Error
    url: str
    def __init__(
        self,
        error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...,
        url: _Optional[str] = ...,
    ) -> None: ...
