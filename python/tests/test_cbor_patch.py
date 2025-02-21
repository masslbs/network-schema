# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from datetime import datetime, timezone

import pytest
import cbor2

from massmarket_hash_event.cbor.patch import (
    ObjectType,
    OpString,
    PatchPath,
    Patch,
    PatchSetHeader,
    SignedPatchSet,
)
from massmarket_hash_event.cbor.base_types import (
    Uint256,
    EthereumAddress,
)


def test_patch_path_manifest():
    # Test manifest path (no ID needed)
    path = PatchPath(type=ObjectType.MANIFEST, fields=["pricingCurrency"])

    encoded = path.to_cbor_list()
    assert encoded[0] == "manifest"
    assert len(encoded) == 2
    assert encoded[1] == "pricingCurrency"

    # Test invalid manifest path (with ID)
    with pytest.raises(ValueError):
        PatchPath(type=ObjectType.MANIFEST, object_id=1)


def test_patch_path_account():
    addr = EthereumAddress("0x0102030405060708090a0b0c0d0e0f1011121314")
    path = PatchPath(type=ObjectType.ACCOUNT, account_addr=addr, fields=["balance"])

    encoded = path.to_cbor_list()
    assert encoded[0] == "account"
    assert encoded[1] == bytes(addr)
    assert encoded[2] == "balance"

    # Test missing account ID
    with pytest.raises(ValueError):
        PatchPath(type=ObjectType.ACCOUNT)


def test_patch_path_listing():
    path = PatchPath(type=ObjectType.LISTING, object_id=123, fields=["price"])

    encoded = path.to_cbor_list()
    assert encoded[0] == "listing"
    assert encoded[1] == 123
    assert encoded[2] == "price"

    # Test missing object ID
    with pytest.raises(ValueError):
        PatchPath(type=ObjectType.LISTING)


def test_patch():
    path = PatchPath(type=ObjectType.MANIFEST, fields=["pricingCurrency"])

    patch = Patch(op=OpString.REPLACE, path=path, value=b"some_cbor_data")

    # Use new to_cbor() method to get the serialized bytes
    serialized = cbor2.dumps(patch.to_cbor_dict())
    assert isinstance(serialized, bytes)

    decoded = cbor2.loads(serialized)
    assert decoded["Op"] == "replace"
    assert decoded["Path"] == ["manifest", "pricingCurrency"]
    assert decoded["Value"] == b"some_cbor_data"


def test_signed_patch_set():
    path = PatchPath(type=ObjectType.MANIFEST, fields=["pricingCurrency"])

    patch = Patch(op=OpString.REPLACE, path=path, value=b"some_cbor_data")

    header = PatchSetHeader(
        key_card_nonce=1,
        shop_id=Uint256(123),
        timestamp=datetime.now(timezone.utc),
        root_hash=b"some_hash",
    )

    patch_set = SignedPatchSet(
        header=header, signature=b"some_signature", patches=[patch]
    )

    # Test CBOR dict conversion
    cbor_dict = patch_set.to_cbor_dict()
    assert "Header" in cbor_dict
    assert "Signature" in cbor_dict
    assert "Patches" in cbor_dict

    # Test empty patches list
    with pytest.raises(ValueError):
        SignedPatchSet(header=header, signature=b"some_signature", patches=[])


def test_patch_path_from_cbor():
    # Test manifest path
    manifest_data = ["manifest", "pricingCurrency"]
    path = PatchPath.from_cbor(manifest_data)
    assert path.type == ObjectType.MANIFEST
    assert path.fields == ["pricingCurrency"]

    # Test account path
    addr = EthereumAddress("0x0102030405060708090a0b0c0d0e0f1011121314")
    account_data = ["account", bytes(addr), "balance"]
    path = PatchPath.from_cbor(account_data)
    assert path.type == ObjectType.ACCOUNT
    assert path.account_addr == addr
    assert path.fields == ["balance"]

    # Test listing path
    listing_data = ["listing", 123, "price"]
    path = PatchPath.from_cbor(listing_data)
    assert path.type == ObjectType.LISTING
    assert path.object_id == 123
    assert path.fields == ["price"]
