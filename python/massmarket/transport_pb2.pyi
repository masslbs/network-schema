# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class PatchSetWriteRequest(_message.Message):
    __slots__ = ("patch_set",)
    PATCH_SET_FIELD_NUMBER: _ClassVar[int]
    patch_set: bytes
    def __init__(self, patch_set: _Optional[bytes] = ...) -> None: ...

class SyncStatusRequest(_message.Message):
    __slots__ = ("subscription_id", "unpushed_patches")
    SUBSCRIPTION_ID_FIELD_NUMBER: _ClassVar[int]
    UNPUSHED_PATCHES_FIELD_NUMBER: _ClassVar[int]
    subscription_id: int
    unpushed_patches: int
    def __init__(
        self,
        subscription_id: _Optional[int] = ...,
        unpushed_patches: _Optional[int] = ...,
    ) -> None: ...

class PingRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
