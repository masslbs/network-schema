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
    ERROR_CODES_UNSPECIFIED: _ClassVar[ErrorCodes]
    ERROR_CODES_NOT_FOUND: _ClassVar[ErrorCodes]
    ERROR_CODES_INVALID: _ClassVar[ErrorCodes]
    ERROR_CODES_NOT_AUTHENTICATED: _ClassVar[ErrorCodes]
    ERROR_CODES_ALREADY_AUTHENTICATED: _ClassVar[ErrorCodes]
    ERROR_CODES_ALREADY_CONNECTED: _ClassVar[ErrorCodes]
    ERROR_CODES_TOO_MANY_CONCURRENT_REQUESTS: _ClassVar[ErrorCodes]
    ERROR_CODES_UNLINKED_KEYCARD: _ClassVar[ErrorCodes]
    ERROR_CODES_MINUMUM_VERSION_NOT_REACHED: _ClassVar[ErrorCodes]
    ERROR_CODES_OUT_OF_STOCK: _ClassVar[ErrorCodes]
    ERROR_CODES_SIMULATED: _ClassVar[ErrorCodes]
ERROR_CODES_UNSPECIFIED: ErrorCodes
ERROR_CODES_NOT_FOUND: ErrorCodes
ERROR_CODES_INVALID: ErrorCodes
ERROR_CODES_NOT_AUTHENTICATED: ErrorCodes
ERROR_CODES_ALREADY_AUTHENTICATED: ErrorCodes
ERROR_CODES_ALREADY_CONNECTED: ErrorCodes
ERROR_CODES_TOO_MANY_CONCURRENT_REQUESTS: ErrorCodes
ERROR_CODES_UNLINKED_KEYCARD: ErrorCodes
ERROR_CODES_MINUMUM_VERSION_NOT_REACHED: ErrorCodes
ERROR_CODES_OUT_OF_STOCK: ErrorCodes
ERROR_CODES_SIMULATED: ErrorCodes

class Error(_message.Message):
    __slots__ = ["code", "message"]
    CODE_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_FIELD_NUMBER: _ClassVar[int]
    code: ErrorCodes
    message: str
    def __init__(self, code: _Optional[_Union[ErrorCodes, str]] = ..., message: _Optional[str] = ...) -> None: ...
