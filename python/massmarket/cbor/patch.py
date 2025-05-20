# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import List, Optional, Any

import cbor2

from massmarket.cbor.base_types import (
    Uint256,
    EthereumAddress,
)


class ObjectType(str, Enum):
    """Represents the type of object being patched"""

    SCHEMA_VERSION = "SchemaVersion"
    MANIFEST = "Manifest"
    ACCOUNT = "Accounts"
    LISTING = "Listings"
    ORDER = "Orders"
    TAG = "Tags"
    INVENTORY = "Inventory"

    @classmethod
    def is_valid(cls, value: str) -> bool:
        try:
            cls(value)
            return True
        except ValueError:
            return False

    def __eq__(self, other) -> bool:
        if isinstance(other, str):
            return self.value == other
        elif isinstance(other, ObjectType):
            return self.value == other.value
        else:
            return False


class OpString(str, Enum):
    """Represents the type of patch operation"""

    ADD = "add"
    APPEND = "append"
    REPLACE = "replace"
    REMOVE = "remove"
    INCREMENT = "increment"
    DECREMENT = "decrement"


@dataclass
class PatchPath:
    """Represents a path in a patch operation"""

    type: ObjectType
    object_id: Optional[int] = None  # for listing/order
    account_addr: Optional[EthereumAddress] = None  # for account
    tag_name: Optional[str] = None  # for tag
    fields: List[Any] = None

    def __post_init__(self):
        if self.fields is None:
            self.fields = []

        # Validate type-specific requirements
        if self.type == ObjectType.MANIFEST:
            if any([self.object_id, self.account_addr, self.tag_name]):
                raise ValueError("manifest patch should not have an id")
        elif self.type == ObjectType.ACCOUNT:
            if not self.account_addr:
                raise ValueError("account patch needs an id")
            if any([self.object_id, self.tag_name]):
                raise ValueError("account patch should not have object_id or tag_name")
        elif self.type in [ObjectType.LISTING, ObjectType.ORDER, ObjectType.INVENTORY]:
            if self.object_id is None:
                raise ValueError(f"{self.type} patch needs an id")
            if any([self.account_addr, self.tag_name]):
                raise ValueError(
                    f"{self.type} patch should not have account_addr or tag_name"
                )
        elif self.type == ObjectType.TAG:
            if not self.tag_name:
                raise ValueError("tag patch needs a tag name")
            if any([self.object_id, self.account_addr]):
                raise ValueError("tag patch should not have object_id or account_addr")

    def to_cbor_list(self) -> List[Any]:
        """Encode to CBOR array format"""
        path = [self.type]

        if self.type == ObjectType.MANIFEST:
            pass  # no id needed
        elif self.type == ObjectType.ACCOUNT:
            path.append(bytes(self.account_addr))
        elif self.type in [ObjectType.LISTING, ObjectType.ORDER, ObjectType.INVENTORY]:
            path.append(self.object_id)
        elif self.type == ObjectType.TAG:
            path.append(self.tag_name)

        path.extend(self.fields)
        return path

    @classmethod
    def from_cbor(cls, data: List[Any]) -> "PatchPath":
        """Create from CBOR array format"""
        if not data or not isinstance(data[0], str):
            raise ValueError("Invalid patch path format")

        obj_type = ObjectType(data[0])
        fields = []
        object_id = None
        account_addr = None
        tag_name = None

        if obj_type != ObjectType.MANIFEST:
            if len(data) < 2:
                raise ValueError(f"{obj_type} patch needs an id")

            if obj_type == ObjectType.ACCOUNT:
                if (
                    not isinstance(data[1], bytes)
                    or len(data[1]) != EthereumAddress.SIZE
                ):
                    raise ValueError("Invalid ethereum address")
                account_addr = EthereumAddress(data[1])
            elif obj_type in [
                ObjectType.LISTING,
                ObjectType.ORDER,
                ObjectType.INVENTORY,
            ]:
                if not isinstance(data[1], int):
                    raise ValueError("Invalid object id")
                object_id = data[1]
            elif obj_type == ObjectType.TAG:
                if not isinstance(data[1], str):
                    raise ValueError("Invalid tag name")
                tag_name = data[1]

            fields = data[2:]
        else:
            fields = data[1:]

        return cls(
            type=obj_type,
            object_id=object_id,
            account_addr=account_addr,
            tag_name=tag_name,
            fields=fields,
        )


@dataclass
class Patch:
    """Represents a single patch operation"""

    op: OpString
    path: PatchPath
    value: any  # cbor.RawMessage equivalent

    def __post_init__(self):
        if not isinstance(self.op, OpString):
            self.op = OpString(self.op)

    def to_cbor_dict(self) -> dict:
        v = self.value
        if hasattr(v, "to_cbor_dict"):
            v = v.to_cbor_dict()
        return {"Op": self.op.value, "Path": self.path.to_cbor_list(), "Value": v}

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "Patch":
        """Create a Patch from CBOR bytes"""
        d = cbor2.loads(cbor_data)
        return cls.from_cbor_dict(d)

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Patch":
        return cls(
            op=OpString(d["Op"]), path=PatchPath.from_cbor(d["Path"]), value=d["Value"]
        )


@dataclass
class PatchSetHeader:
    """Header information for a patch set"""

    key_card_nonce: int
    shop_id: Uint256
    timestamp: datetime
    root_hash: bytes  # Hash type

    def __post_init__(self):
        if self.key_card_nonce <= 0:
            raise ValueError("KeyCardNonce must be greater than 0")
        if len(self.root_hash) == 0:
            raise ValueError("RootHash cannot be empty")

    def to_cbor_dict(self) -> dict:
        return {
            "KeyCardNonce": self.key_card_nonce,
            "ShopID": self.shop_id.to_cbor_dict(),
            "Timestamp": self.timestamp,
            "RootHash": self.root_hash,
        }

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "PatchSetHeader":
        d = cbor2.loads(cbor_data)
        return cls.from_cbor_dict(d)

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "PatchSetHeader":
        return cls(
            key_card_nonce=d["KeyCardNonce"],
            shop_id=Uint256(d["ShopID"]),
            timestamp=d["Timestamp"],
            root_hash=d["RootHash"],
        )


@dataclass
class SignedPatchSet:
    """A complete signed set of patches"""

    header: PatchSetHeader
    signature: bytes  # Signature type
    patches: List[Patch]

    def __post_init__(self):
        if not self.patches:
            raise ValueError("Patches cannot be empty")
        if len(self.signature) == 0:
            raise ValueError("Signature cannot be empty")

    def to_cbor_dict(self) -> dict:
        return {
            "Header": self.header.to_cbor_dict(),
            "Signature": self.signature,
            "Patches": [p.to_cbor_dict() for p in self.patches],
        }

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "SignedPatchSet":
        """Create a SignedPatchSet from CBOR bytes"""
        d = cbor2.loads(cbor_data)
        return cls.from_cbor_dict(d)

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "SignedPatchSet":
        header = PatchSetHeader(
            key_card_nonce=d["Header"]["KeyCardNonce"],
            shop_id=Uint256(d["Header"]["ShopID"]),
            timestamp=d["Header"]["Timestamp"],
            root_hash=d["Header"]["RootHash"],
        )

        patches = [
            Patch(
                op=OpString(p["Op"]),
                path=(
                    PatchPath.from_cbor(p["Path"])
                    if isinstance(p["Path"], list)
                    else p["Path"]
                ),
                value=p["Value"],
            )
            for p in d["Patches"]
        ]

        return cls(header=header, signature=d["Signature"], patches=patches)
