# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from massmarket_hash_event import error_pb2 as _error_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class AuthenticateRequest(_message.Message):
    __slots__ = ["request_id", "public_key"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    PUBLIC_KEY_FIELD_NUMBER: _ClassVar[int]
    request_id: _base_types_pb2.RequestId
    public_key: _base_types_pb2.PublicKey
    def __init__(self, request_id: _Optional[_Union[_base_types_pb2.RequestId, _Mapping]] = ..., public_key: _Optional[_Union[_base_types_pb2.PublicKey, _Mapping]] = ...) -> None: ...

class AuthenticateResponse(_message.Message):
    __slots__ = ["request_id", "error", "challenge"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    CHALLENGE_FIELD_NUMBER: _ClassVar[int]
    request_id: _base_types_pb2.RequestId
    error: _error_pb2.Error
    challenge: bytes
    def __init__(self, request_id: _Optional[_Union[_base_types_pb2.RequestId, _Mapping]] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ..., challenge: _Optional[bytes] = ...) -> None: ...

class ChallengeSolvedRequest(_message.Message):
    __slots__ = ["request_id", "signature"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    request_id: _base_types_pb2.RequestId
    signature: _base_types_pb2.Signature
    def __init__(self, request_id: _Optional[_Union[_base_types_pb2.RequestId, _Mapping]] = ..., signature: _Optional[_Union[_base_types_pb2.Signature, _Mapping]] = ...) -> None: ...

class ChallengeSolvedResponse(_message.Message):
    __slots__ = ["request_id", "error"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    request_id: _base_types_pb2.RequestId
    error: _error_pb2.Error
    def __init__(self, request_id: _Optional[_Union[_base_types_pb2.RequestId, _Mapping]] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ...) -> None: ...
