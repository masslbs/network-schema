# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from datetime import datetime, timezone

import pytest
import cbor2

from massmarket.cbor.patch import (
    ObjectType,
    OpString,
    PatchPath,
    Patch,
    PatchSetHeader,
    SignedPatchSet,
)
from massmarket.cbor.base_types import (
    Uint256,
    EthereumAddress,
)


def test_patch_path_manifest():
    # Test manifest path (no ID needed)
    path = PatchPath(type=ObjectType.MANIFEST, fields=["pricingCurrency"])

    encoded = path.to_cbor_list()
    assert encoded[0] == "Manifest"
    assert len(encoded) == 2
    assert encoded[1] == "pricingCurrency"

    # Test invalid manifest path (with ID)
    with pytest.raises(ValueError):
        PatchPath(type=ObjectType.MANIFEST, object_id=1)


def test_patch_path_account():
    addr = EthereumAddress("0x0102030405060708090a0b0c0d0e0f1011121314")
    path = PatchPath(type=ObjectType.ACCOUNT, account_addr=addr, fields=["balance"])

    encoded = path.to_cbor_list()
    assert encoded[0] == "Accounts"
    assert encoded[1] == bytes(addr)
    assert encoded[2] == "balance"

    # Test missing account ID
    with pytest.raises(ValueError):
        PatchPath(type=ObjectType.ACCOUNT)


def test_patch_path_listing():
    path = PatchPath(type=ObjectType.LISTING, object_id=123, fields=["price"])

    encoded = path.to_cbor_list()
    assert encoded[0] == "Listings"
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
    assert decoded["Path"] == ["Manifest", "pricingCurrency"]
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
    manifest_data = ["Manifest", "pricingCurrency"]
    path = PatchPath.from_cbor(manifest_data)
    assert path.type == ObjectType.MANIFEST
    assert path.fields == ["pricingCurrency"]

    # Test account path
    addr = EthereumAddress("0x0102030405060708090a0b0c0d0e0f1011121314")
    account_data = ["Accounts", bytes(addr), "balance"]
    path = PatchPath.from_cbor(account_data)
    assert path.type == ObjectType.ACCOUNT
    assert path.account_addr == addr
    assert path.fields == ["balance"]

    # Test listing path
    listing_data = ["Listings", 123, "price"]
    path = PatchPath.from_cbor(listing_data)
    assert path.type == ObjectType.LISTING
    assert path.object_id == 123
    assert path.fields == ["price"]


import os
import cbor2


def test_patch_from_vectors_file():
    file_path = os.path.join(
        os.path.dirname(__file__), "..", "..", "vectors", "ManifestOkay.cbor"
    )
    with open(file_path, "rb") as f:
        vectors = cbor2.load(f)

    # Test the patches in the PatchSet
    patch_set = vectors["PatchSet"]
    for patch_data in patch_set["Patches"]:
        # Convert the patch data to a Patch object
        patch = Patch.from_cbor_dict(patch_data)

        # Verify the patch operation
        assert patch.op in [op for op in OpString]

        # Verify the path
        assert isinstance(patch.path, PatchPath)
        assert patch.path.type in [obj_type for obj_type in ObjectType]

        # Value should be present for most operations except REMOVE
        if patch.op != OpString.REMOVE:
            assert patch.value is not None

    # Test error patches
    file_path = os.path.join(
        os.path.dirname(__file__), "..", "..", "vectors", "ManifestError.cbor"
    )
    with open(file_path, "rb") as f:
        vectors = cbor2.load(f)

    for error_case in vectors["Patches"]:
        patch_data = error_case["Patch"]
        patch = Patch.from_cbor_dict(patch_data)

        # Verify basic structure
        assert isinstance(patch, Patch)
        assert patch.op in [op for op in OpString]
        assert isinstance(patch.path, PatchPath)

        # Check that the error message is present
        assert "Error" in error_case
        assert isinstance(error_case["Error"], str)


def test_signed_patch_set_from_vectors():
    file_path = os.path.join(
        os.path.dirname(__file__), "..", "..", "vectors", "ManifestOkay.cbor"
    )
    with open(file_path, "rb") as f:
        vectors = cbor2.load(f)

    patch_set_data = vectors["PatchSet"]

    # Convert the patch set data to a SignedPatchSet object
    patch_set = SignedPatchSet.from_cbor_dict(patch_set_data)

    # Verify the patch set structure
    assert isinstance(patch_set, SignedPatchSet)
    assert isinstance(patch_set.header, PatchSetHeader)
    assert isinstance(patch_set.signature, bytes)
    assert len(patch_set.patches) > 0

    # Verify header fields
    assert isinstance(patch_set.header.shop_id, Uint256)
    assert isinstance(patch_set.header.timestamp, datetime)
    assert isinstance(patch_set.header.root_hash, bytes)

    # Verify patches
    for patch in patch_set.patches:
        assert isinstance(patch, Patch)
        assert patch.op in [op for op in OpString]
        assert isinstance(patch.path, PatchPath)
