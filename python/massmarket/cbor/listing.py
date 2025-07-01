# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from typing import Optional, List, Dict
from datetime import datetime
from enum import IntEnum

import cbor2

from massmarket.cbor.base_types import Uint256, PriceModifier


@dataclass
class ListingMetadata:
    title: str
    description: str
    images: Optional[List[str]] = None

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "ListingMetadata":
        return cls(
            title=d["Title"],
            description=d["Description"],
            images=d.get("Images"),
        )

    def to_cbor_dict(self) -> dict:
        d = {
            "Title": self.title,
            "Description": self.description,
        }
        if self.images is not None:
            d["Images"] = self.images
        return d


@dataclass
class ListingVariation:
    variation_info: ListingMetadata
    price_modifier: Optional[PriceModifier] = None
    sku: Optional[str] = None

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "ListingVariation":
        pm = d.get("PriceModifier")
        if pm is not None:
            pm = PriceModifier.from_cbor_dict(pm)
        return cls(
            variation_info=ListingMetadata.from_cbor_dict(d["VariationInfo"]),
            price_modifier=pm,
            sku=d.get("SKU"),
        )

    def to_cbor_dict(self) -> dict:
        d = {
            "VariationInfo": self.variation_info.to_cbor_dict(),
        }
        if self.price_modifier is not None:
            d["PriceModifier"] = self.price_modifier.to_cbor_dict()
        if self.sku is not None:
            d["SKU"] = self.sku
        return d


@dataclass
class ListingOption:
    title: str
    variations: Optional[Dict[str, ListingVariation]] = None

    def __post_init__(self):
        if self.variations is not None and not self.variations:
            raise ValueError("Variations map cannot be empty if present")

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "ListingOption":
        variations = d.get("Variations")
        if variations is not None:
            variations = {
                k: ListingVariation.from_cbor_dict(v) for k, v in variations.items()
            }
        return cls(
            title=d["Title"],
            variations=variations,
        )

    def to_cbor_dict(self) -> dict:
        d = {"Title": self.title}
        if self.variations is not None:
            d["Variations"] = {k: v.to_cbor_dict() for k, v in self.variations.items()}
        return d


class ListingViewState(IntEnum):
    UNSPECIFIED = 0
    PUBLISHED = 1
    DELETED = 2


@dataclass
class Listing:
    id: int
    price: Uint256
    metadata: ListingMetadata
    view_state: ListingViewState = ListingViewState.UNSPECIFIED
    options: Optional[Dict[str, ListingOption]] = None

    def __post_init__(self):
        if self.options is not None and not self.options:
            raise ValueError("Options map cannot be empty if present")

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Listing":
        options = d.get("Options")
        if options is not None:
            options = {k: ListingOption.from_cbor_dict(v) for k, v in options.items()}

        return cls(
            id=d["ID"],
            price=Uint256(d["Price"]),
            metadata=ListingMetadata.from_cbor_dict(d["Metadata"]),
            view_state=ListingViewState(d["ViewState"]),
            options=options,
        )

    def to_cbor_dict(self) -> dict:
        d = {
            "ID": self.id,
            "Price": self.price.to_cbor_dict(),
            "Metadata": self.metadata.to_cbor_dict(),
            "ViewState": self.view_state.value,
        }
        if self.options is not None:
            d["Options"] = {k: v.to_cbor_dict() for k, v in self.options.items()}
        return d

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "Listing":
        d = cbor2.loads(cbor_data)
        return cls.from_cbor_dict(d)
