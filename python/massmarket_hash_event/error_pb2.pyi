# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class ErrorCodes(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = []
    unknown: _ClassVar[ErrorCodes]
    notFound: _ClassVar[ErrorCodes]
    invalid: _ClassVar[ErrorCodes]
    notAuthenticated: _ClassVar[ErrorCodes]
    alreadyAuthenticated: _ClassVar[ErrorCodes]
    alreadyConnected: _ClassVar[ErrorCodes]
    tooManyConcurrentRequests: _ClassVar[ErrorCodes]
    unlinkedKeyCard: _ClassVar[ErrorCodes]
    minumumVersionNotReached: _ClassVar[ErrorCodes]
    outOfStock: _ClassVar[ErrorCodes]
    simulated: _ClassVar[ErrorCodes]
unknown: ErrorCodes
notFound: ErrorCodes
invalid: ErrorCodes
notAuthenticated: ErrorCodes
alreadyAuthenticated: ErrorCodes
alreadyConnected: ErrorCodes
tooManyConcurrentRequests: ErrorCodes
unlinkedKeyCard: ErrorCodes
minumumVersionNotReached: ErrorCodes
outOfStock: ErrorCodes
simulated: ErrorCodes

class Error(_message.Message):
    __slots__ = ["code", "message"]
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    code: ErrorCodes
    message: str
    def __init__(self, code: _Optional[_Union[ErrorCodes, str]] = ..., message: _Optional[str] = ...) -> None: ...
