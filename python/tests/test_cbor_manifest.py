# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import base64
import json
import os
import pytest
import cbor2

from massmarket_hash_event.cbor.base_types import (
    Uint256,
    PriceModifier,
    ModificationAbsolute,
)
from massmarket_hash_event.cbor.chain_address import ChainAddress
from massmarket_hash_event.cbor.manifest import (
    Manifest,
    Payee,
    ShippingRegion,
)


def test_cbor_manifest_cbor_keys():
    # Create a simple Manifest instance matching a valid Go schema.
    manifest = Manifest(
        shop_id=Uint256.random(),
        payees={
            "default": Payee(
                address="0x0102030405060708090a0b0c0d0e0f1011121314",  # or a ChainAddress if implemented
                call_as_contract=False,
            )
        },
        accepted_currencies=[
            ChainAddress(
                chain_id=1337, address="0x0000000000000000000000000000000000000000"
            ),
            ChainAddress(
                chain_id=1337, address="0xffffffffffffffffffffffffffffffffffffffff"
            ),
        ],
        pricing_currency=ChainAddress(
            chain_id=1337, address="0xff00ff00ff00ff00ff00ff00ff00ff00ff00ff00"
        ),
        shipping_regions={
            "default": ShippingRegion(
                country="DE",
                postcode="",
                city="",
                price_modifiers={
                    "some_modifier": PriceModifier(
                        modification_percents=Uint256(5),
                    )
                },
            )
        },
    )

    # When we convert to a CBOR dict, check that we use CamelCase keys.
    cbor_dict = manifest.to_cbor_dict()
    assert "ShopID" in cbor_dict
    assert "Payees" in cbor_dict
    assert "AcceptedCurrencies" in cbor_dict
    assert "PricingCurrency" in cbor_dict
    # Optional ShippingRegions should appear if non-null.
    assert "ShippingRegions" in cbor_dict

    # Also check that the payee subdict used the Go naming
    payee_dict = cbor_dict["Payees"]["default"]
    assert "Address" in payee_dict
    assert "CallAsContract" in payee_dict


def test_cbor_manifest_roundtrip():
    # Build an instance and then round-trip it through CBOR encoding and decoding.
    original = Manifest(
        shop_id=Uint256.random(),
        payees={
            "default": Payee(
                address=ChainAddress(
                    chain_id=1337, address="0x0102030405060708090a0b0c0d0e0f1011121314"
                ),
                call_as_contract=False,
            )
        },
        accepted_currencies=[
            ChainAddress(
                chain_id=1337, address="0x0000000000000000000000000000000000000000"
            ),
            ChainAddress(
                chain_id=1337, address="0xffffffffffffffffffffffffffffffffffffffff"
            ),
        ],
        pricing_currency=ChainAddress(
            chain_id=1337, address="0x0000000000000000000000000000000000000000"
        ),
        shipping_regions={
            "default": ShippingRegion(
                country="DE",
                postcode="",
                city="",
                price_modifiers={
                    "mod": PriceModifier(
                        modification_absolute=ModificationAbsolute(
                            amount=Uint256(1000),
                            plus=False,
                        ),
                    )
                },
            )
        },
    )
    cbor_bytes = cbor2.dumps(original.to_cbor_dict())
    decoded = Manifest.from_cbor(cbor_bytes)
    # Depending on your Uint256/ChainAddress objects' implementations, you may need to adjust the equality check.
    assert original == decoded


def test_cbor_price_modifier_validation():
    # PriceModifier without any set values should raise an error on initialization.
    with pytest.raises(ValueError):
        PriceModifier()


def test_cbor_manifest_from_vectors_file():
    # Use the provided vector file (for example ManifestOkay.json) to drive assertions.
    file_path = os.path.join(
        os.path.dirname(__file__), "..", "..", "vectors", "ManifestOkay.json"
    )
    with open(file_path, "r") as f:
        vectors = json.load(f)

    def verify_manifest_snapshot(snapshot):
        encoded_b64 = snapshot["Before"]["Encoded"]
        cbor_data = base64.b64decode(encoded_b64)

        # Decode using our helper.
        shop_obj = cbor2.loads(cbor_data)
        manifest_cbor = cbor2.dumps(shop_obj["Manifest"])
        manifest_obj = Manifest.from_cbor(manifest_cbor)

        # Compare key properties. Note that the JSON "Value" in the vector file holds the manifest
        # under the "Manifest" key – which uses CamelCase keys.
        expected = snapshot["Before"]["Value"]["Manifest"]
        # Convert the expected dictionary to a ChainAddress object for comparison
        expected_pricing_currency = ChainAddress(
            chain_id=expected["PricingCurrency"]["ChainID"],
            address=expected["PricingCurrency"]["Address"],
        )
        assert manifest_obj.pricing_currency == expected_pricing_currency

        # Check payees
        expected_payee = expected["Payees"]["default"]
        expected_payee_addr = ChainAddress(
            chain_id=expected_payee["Address"]["ChainID"],
            address=expected_payee["Address"]["Address"],
        )
        assert manifest_obj.payees["default"].address == expected_payee_addr
        assert (
            manifest_obj.payees["default"].call_as_contract
            == expected_payee["CallAsContract"]
        )

        # Check accepted currencies
        assert len(manifest_obj.accepted_currencies) == len(
            expected["AcceptedCurrencies"]
        )
        for actual, expected_curr in zip(
            manifest_obj.accepted_currencies, expected["AcceptedCurrencies"]
        ):
            expected_chain_addr = ChainAddress(
                chain_id=expected_curr["ChainID"], address=expected_curr["Address"]
            )
            assert actual == expected_chain_addr

        # Check shop ID
        assert manifest_obj.shop_id == Uint256(expected["ShopID"])

        # Check shipping regions
        assert len(manifest_obj.shipping_regions) == len(expected["ShippingRegions"])
        expected_region = expected["ShippingRegions"]["default"]
        actual_region = manifest_obj.shipping_regions["default"]
        assert actual_region.country == expected_region["Country"]
        assert actual_region.postcode == expected_region["Postcode"]
        assert actual_region.city == expected_region["City"]
        assert (
            actual_region.price_modifiers == expected_region["PriceModifiers"]
        )  # Both should be None

    # Verify each snapshot in the vector file
    for snapshot in vectors["Snapshots"]:
        print(f"Verifying snapshot {snapshot['Name']}")
        verify_manifest_snapshot(snapshot)
