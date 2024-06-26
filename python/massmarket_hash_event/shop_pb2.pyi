# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class ShopCurrency(_message.Message):
    __slots__ = ["chain_id", "token_addr"]
    CHAIN_ID_FIELD_NUMBER: _ClassVar[int]
    TOKEN_ADDR_FIELD_NUMBER: _ClassVar[int]
    chain_id: int
    token_addr: bytes
    def __init__(self, chain_id: _Optional[int] = ..., token_addr: _Optional[bytes] = ...) -> None: ...
