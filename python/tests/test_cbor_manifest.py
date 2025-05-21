# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import base64
import json
import os
import pytest
import cbor2

from massmarket.cbor.base_types import (
    ChainAddress,
    Uint256,
    PriceModifier,
    ModificationAbsolute,
    EthereumAddress,
)
from massmarket.cbor.manifest import (
    Manifest,
    PayeeMetadata,
    ShippingRegion,
)
from massmarket.cbor import Shop


def test_cbor_manifest_cbor_keys():
    # Create a simple Manifest instance matching a valid Go schema.
    test_addr = EthereumAddress("0x0102030405060708090a0b0c0d0e0f1011121314")
    manifest = Manifest(
        shop_id=Uint256.random(),
        payees={
            1337: {
                test_addr: PayeeMetadata(call_as_contract=False),
            }
        },
        accepted_currencies={
            1337: {
                EthereumAddress("0x0000000000000000000000000000000000000000"),
                EthereumAddress("0xff00ff00ff00ff00ff00ff00ff00ff00ff00ff00"),
            }
        },
        pricing_currency=ChainAddress(
            chain_id=1337, address="0xff00ff00ff00ff00ff00ff00ff00ff00ff00ff00"
        ),
        shipping_regions={
            "default": ShippingRegion(
                country="DE",
                postal_code="",
                city="",
                price_modifiers={
                    "some_modifier": PriceModifier(
                        modification_percent=Uint256(5),
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
    payee_dict = cbor_dict["Payees"][1337]
    assert test_addr in payee_dict
    assert "CallAsContract" in payee_dict[test_addr]


def test_cbor_manifest_roundtrip():
    # Build an instance and then round-trip it through CBOR encoding and decoding.
    original = Manifest(
        shop_id=Uint256.random(),
        payees={
            1337: {
                EthereumAddress(
                    "0x0102030405060708090a0b0c0d0e0f1011121314"
                ): PayeeMetadata(call_as_contract=True),
                EthereumAddress(
                    "0xffffffffffffffffffffffffffffffffffffffff"
                ): PayeeMetadata(call_as_contract=False),
            }
        },
        accepted_currencies={
            1337: {
                EthereumAddress("0x0000000000000000000000000000000000000000"),
                EthereumAddress("0xffffffffffffffffffffffffffffffffffffffff"),
            }
        },
        pricing_currency=ChainAddress(
            chain_id=1337, address="0x0000000000000000000000000000000000000000"
        ),
        shipping_regions={
            "default": ShippingRegion(
                country="DE",
                postal_code="",
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
    file_path = os.path.join(
        os.path.dirname(__file__), "..", "..", "vectors", "ManifestOkay.json"
    )
    with open(file_path, "r") as f:
        vectors = json.load(f)

    for a_or_b in ["After", "Before"]:
        print(f"Testing {a_or_b}")

        for snap in vectors["Snapshots"]:
            print(f"Testing {snap['Name']}")
            encoded_b64 = snap[a_or_b]["Encoded"]
            cbor_data = base64.b64decode(encoded_b64)

            # Decode using our helper
            shop = Shop.from_cbor(cbor_data)
            expected = snap[a_or_b]["Value"]["Manifest"]

            # Verify manifest properties
            verify_manifest(shop.manifest, expected)

            want_hash = base64.b64decode(snap[a_or_b]["Hash"])
            got_hash = shop.hash()
            assert want_hash == got_hash


def verify_manifest(manifest_obj: Manifest, expected: dict):
    # Check shop ID
    # TODO: Uint256/big.Int printed as {} in json
    # assert manifest_obj.shop_id == expected["ShopID"]

    # Check pricing currency
    if "PricingCurrency" in expected and expected["PricingCurrency"] is not None:
        print(expected["PricingCurrency"])
        expected_pricing_currency = ChainAddress(
            chain_id=expected["PricingCurrency"]["ChainID"],
            address=expected["PricingCurrency"]["Address"],
        )
        assert manifest_obj.pricing_currency == expected_pricing_currency
    else:
        assert manifest_obj.pricing_currency is None

    # Check payees
    if "Payees" in expected and expected["Payees"] is not None:
        assert len(manifest_obj.payees) == len(expected["Payees"])
        for chain_id_str, expected_payees in expected["Payees"].items():
            chain_id = int(chain_id_str)
            assert chain_id in manifest_obj.payees
            assert len(manifest_obj.payees[chain_id]) == len(expected_payees)

            for eth_addr_hex, expected_metadata in expected_payees.items():
                eth_addr = EthereumAddress(eth_addr_hex)
                assert eth_addr in manifest_obj.payees[chain_id]
                assert (
                    manifest_obj.payees[chain_id][eth_addr].call_as_contract
                    == expected_metadata["CallAsContract"]
                )
    else:
        assert manifest_obj.payees is None or len(manifest_obj.payees) == 0

    # Check accepted currencies
    if "AcceptedCurrencies" in expected and expected["AcceptedCurrencies"] is not None:
        assert len(manifest_obj.accepted_currencies) == len(
            expected["AcceptedCurrencies"]
        )
        for chain_id_str, expected_currencies in expected["AcceptedCurrencies"].items():
            chain_id = int(chain_id_str)
            assert chain_id in manifest_obj.accepted_currencies
            assert len(manifest_obj.accepted_currencies[chain_id]) == len(
                expected_currencies
            )

            for eth_addr_hex in expected_currencies:
                eth_addr = EthereumAddress(eth_addr_hex)
                assert eth_addr in manifest_obj.accepted_currencies[chain_id]
    else:
        assert (
            manifest_obj.accepted_currencies is None
            or len(manifest_obj.accepted_currencies) == 0
        )

    # Check shipping regions
    if "ShippingRegions" in expected and len(expected["ShippingRegions"]) > 0:
        assert len(manifest_obj.shipping_regions) == len(expected["ShippingRegions"])
        for key, expected_region in expected["ShippingRegions"].items():
            assert key in manifest_obj.shipping_regions
            actual_region = manifest_obj.shipping_regions[key]
            assert actual_region.country == expected_region["Country"]
            assert actual_region.postal_code == expected_region["PostalCode"]
            assert actual_region.city == expected_region["City"]

            # Check price modifiers if present
            if (
                "PriceModifiers" in expected_region
                and expected_region["PriceModifiers"]
            ):
                assert len(actual_region.price_modifiers) == len(
                    expected_region["PriceModifiers"]
                )
                # Additional verification of price modifiers could be added here
            else:
                assert (
                    actual_region.price_modifiers is None
                    or len(actual_region.price_modifiers) == 0
                )
    else:
        assert (
            manifest_obj.shipping_regions is None
            or len(manifest_obj.shipping_regions) == 0
        )
