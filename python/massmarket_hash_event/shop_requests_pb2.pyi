# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

from massmarket_hash_event import error_pb2 as _error_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CommitItemsToOrderRequest(_message.Message):
    __slots__ = ["request_id", "order_id", "erc20_addr", "chain_id"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ORDER_ID_FIELD_NUMBER: _ClassVar[int]
    ERC20_ADDR_FIELD_NUMBER: _ClassVar[int]
    CHAIN_ID_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    order_id: bytes
    erc20_addr: bytes
    chain_id: int
    def __init__(self, request_id: _Optional[bytes] = ..., order_id: _Optional[bytes] = ..., erc20_addr: _Optional[bytes] = ..., chain_id: _Optional[int] = ...) -> None: ...

class CommitItemsToOrderResponse(_message.Message):
    __slots__ = ["request_id", "error", "order_finalized_id"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    ORDER_FINALIZED_ID_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: _error_pb2.Error
    order_finalized_id: bytes
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ..., order_finalized_id: _Optional[bytes] = ...) -> None: ...

class GetBlobUploadURLRequest(_message.Message):
    __slots__ = ["request_id"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    def __init__(self, request_id: _Optional[bytes] = ...) -> None: ...

class GetBlobUploadURLResponse(_message.Message):
    __slots__ = ["request_id", "error", "url"]
    REQUEST_ID_FIELD_NUMBER: _ClassVar[int]
    ERROR_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    request_id: bytes
    error: _error_pb2.Error
    url: str
    def __init__(self, request_id: _Optional[bytes] = ..., error: _Optional[_Union[_error_pb2.Error, _Mapping]] = ..., url: _Optional[str] = ...) -> None: ...
