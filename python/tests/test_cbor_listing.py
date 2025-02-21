import pytest
from datetime import datetime, timezone
import os
import json
import base64

import cbor2

from massmarket_hash_event.cbor.base_types import (
    Uint256,
    ModificationAbsolute,
    PriceModifier,
)
from massmarket_hash_event.cbor import cbor_encode

from massmarket_hash_event.cbor.listing import (
    Listing,
    ListingMetadata,
    ListingOption,
    ListingVariation,
    ListingStockStatus,
    ListingViewState,
)


def test_listing_metadata_roundtrip():
    metadata = ListingMetadata(
        title="Test Product",
        description="A test product description",
        images=["image1.jpg", "image2.jpg"],
    )

    encoded = cbor_encode(metadata.to_cbor_dict())
    decoded = ListingMetadata.from_cbor_dict(cbor2.loads(encoded))

    assert decoded.title == metadata.title
    assert decoded.description == metadata.description
    assert decoded.images == metadata.images


def test_listing_variation_roundtrip():
    price_mod = PriceModifier(
        modification_absolute=ModificationAbsolute(
            amount=Uint256(100),
            plus=True,
        )
    )

    variation = ListingVariation(
        variation_info=ListingMetadata(
            title="Red",
            description="Bright red color",
        ),
        price_modifier=price_mod,
        sku="RED-001",
    )

    encoded = cbor_encode(variation.to_cbor_dict())
    decoded = ListingVariation.from_cbor_dict(variation.to_cbor_dict())

    assert decoded.variation_info.title == variation.variation_info.title
    assert decoded.sku == variation.sku
    assert (
        decoded.price_modifier.modification_absolute.amount
        == variation.price_modifier.modification_absolute.amount
    )


def test_listing_option_roundtrip():
    option = ListingOption(
        title="Color",
        variations={
            "red": ListingVariation(
                variation_info=ListingMetadata(
                    title="Red",
                    description="Bright red color",
                ),
                sku="RED-001",
            ),
            "blue": ListingVariation(
                variation_info=ListingMetadata(
                    title="Blue",
                    description="Deep blue color",
                ),
                sku="BLUE-001",
            ),
        },
    )

    encoded = cbor_encode(option.to_cbor_dict())
    decoded = ListingOption.from_cbor_dict(option.to_cbor_dict())

    assert decoded.title == option.title
    assert len(decoded.variations) == len(option.variations)
    assert decoded.variations["red"].sku == option.variations["red"].sku


def test_listing_stock_status_roundtrip():
    # Test with in_stock
    status1 = ListingStockStatus(
        variation_ids=["red", "large"],
        in_stock=True,
    )

    # Test with expected_in_stock_by
    future_date = datetime(2024, 12, 31, tzinfo=timezone.utc)
    status2 = ListingStockStatus(
        variation_ids=["blue", "small"],
        expected_in_stock_by=future_date,
    )

    for status in [status1, status2]:
        encoded = cbor_encode(status.to_cbor_dict())
        decoded = ListingStockStatus.from_cbor_dict(status.to_cbor_dict())

        assert decoded.variation_ids == status.variation_ids
        assert decoded.in_stock == status.in_stock
        assert decoded.expected_in_stock_by == status.expected_in_stock_by


def test_full_listing_roundtrip():
    listing = Listing(
        id=1,
        price=Uint256(1000),
        metadata=ListingMetadata(
            title="Test Product",
            description="A complete test product",
            images=["main.jpg"],
        ),
        view_state=ListingViewState.PUBLISHED,
        options={
            "color": ListingOption(
                title="Color",
                variations={
                    "red": ListingVariation(
                        variation_info=ListingMetadata(
                            title="Red",
                            description="Red variant",
                        ),
                        sku="RED-001",
                    ),
                },
            ),
        },
        stock_statuses=[
            ListingStockStatus(
                variation_ids=["red"],
                in_stock=True,
            ),
        ],
    )

    encoded = cbor_encode(listing.to_cbor_dict())
    decoded = Listing.from_cbor_dict(listing.to_cbor_dict())

    assert decoded.id == listing.id
    assert decoded.price == listing.price
    assert decoded.metadata.title == listing.metadata.title
    assert decoded.view_state == listing.view_state
    assert len(decoded.options) == len(listing.options)
    assert len(decoded.stock_statuses) == len(listing.stock_statuses)


def test_listing_validation():
    # Test invalid ID
    with pytest.raises(ValueError, match="ID must be greater than 0"):
        Listing(
            id=0,
            price=Uint256(1000),
            metadata=ListingMetadata(
                title="Test",
                description="Test",
            ),
            view_state=ListingViewState.PUBLISHED,
        )

    # Test empty options map
    with pytest.raises(ValueError, match="Options map cannot be empty if present"):
        Listing(
            id=1,
            price=Uint256(1000),
            metadata=ListingMetadata(
                title="Test",
                description="Test",
            ),
            view_state=ListingViewState.PUBLISHED,
            options={},
        )

    # Test ListingStockStatus validation
    with pytest.raises(
        ValueError, match="One of in_stock or expected_in_stock_by must be set"
    ):
        ListingStockStatus(
            variation_ids=["red"],
        )


def test_listing_view_state():
    assert ListingViewState.UNSPECIFIED.value == 0
    assert ListingViewState.PUBLISHED.value == 1
    assert ListingViewState.DELETED.value == 2

def skip_test_listing_from_vectors_file():
    file_path = os.path.join(
        os.path.dirname(__file__), "..", "..", "vectors", "ListingOkay.json"
    )
    with open(file_path, "r") as f:
        vectors = json.load(f)

    def verify_listing_snapshot(snapshot):
        encoded_b64 = snapshot["After"]["Encoded"]
        cbor_data = base64.b64decode(encoded_b64)

        # Decode using our helper
        shop_obj = cbor2.loads(cbor_data)
        # TODO: hamt deserialize
        for listing_id in shop_obj["Listings"].keys():
            expected = shop_obj["Listings"][listing_id]
            listing_obj = Listing.from_cbor_dict(shop_obj["Listings"][listing_id])

            assert listing_obj.id == expected["ID"]
            assert listing_obj.price == Uint256(expected["Price"]) if "Price" in expected else None
            
            # Check metadata
            if "Metadata" in expected:
                assert listing_obj.metadata.title == expected["Metadata"]["Title"]
                assert listing_obj.metadata.description == expected["Metadata"]["Description"]
                assert listing_obj.metadata.images == expected["Metadata"]["Images"]

            # Check view state
            if "ViewState" in expected:
                assert listing_obj.view_state == ListingViewState(expected["ViewState"])

            # Check options
            if "Options" in expected:
                assert len(listing_obj.options) == len(expected["Options"])
                for option_key, option_value in expected["Options"].items():
                    assert option_key in listing_obj.options
                    assert listing_obj.options[option_key].title == option_value["Title"]
                    
                    # Check variations
                    if "Variations" in option_value:
                        for var_key, var_value in option_value["Variations"].items():
                            assert var_key in listing_obj.options[option_key].variations
                            variation = listing_obj.options[option_key].variations[var_key]
                            assert variation.variation_info.title == var_value["VariationInfo"]["Title"]
                            assert variation.variation_info.description == var_value["VariationInfo"]["Description"]
                            assert variation.sku == var_value["SKU"]

            # Check stock statuses
            if "StockStatuses" in expected:
                assert len(listing_obj.stock_statuses) == len(expected["StockStatuses"])
                for actual, expected_status in zip(listing_obj.stock_statuses, expected["StockStatuses"]):
                    assert actual.variation_ids == expected_status["VariationIDs"]
                    assert actual.in_stock == expected_status["InStock"]
                    if "ExpectedInStockBy" in expected_status and expected_status["ExpectedInStockBy"]:
                        assert actual.expected_in_stock_by is not None
                    else:
                        assert actual.expected_in_stock_by is None

    # Verify each snapshot in the vector file
    for snapshot in vectors["Snapshots"]:
        print(f"Verifying snapshot {snapshot['Name']}")
        verify_listing_snapshot(snapshot)

