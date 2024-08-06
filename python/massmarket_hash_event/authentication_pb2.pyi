# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import base_types_pb2 as _base_types_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import (
    ClassVar as _ClassVar,
    Mapping as _Mapping,
    Optional as _Optional,
    Union as _Union,
)

DESCRIPTOR: _descriptor.FileDescriptor

class AuthenticateRequest(_message.Message):
    __slots__ = ["public_key"]
    PUBLIC_KEY_FIELD_NUMBER: _ClassVar[int]
    public_key: _base_types_pb2.PublicKey
    def __init__(
        self, public_key: _Optional[_Union[_base_types_pb2.PublicKey, _Mapping]] = ...
    ) -> None: ...

class ChallengeSolvedRequest(_message.Message):
    __slots__ = ["signature"]
    SIGNATURE_FIELD_NUMBER: _ClassVar[int]
    signature: _base_types_pb2.Signature
    def __init__(
        self, signature: _Optional[_Union[_base_types_pb2.Signature, _Mapping]] = ...
    ) -> None: ...
