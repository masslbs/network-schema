# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from typing import Dict, List, Optional
from io import BytesIO

import cbor2

from massmarket_hash_event.cbor.uint256 import Uint256
from massmarket_hash_event.cbor.chain_address import ChainAddress

all = [
    "cbor_encode",
    # object types
    "Uint256",
    "ChainAddress",
    "ModificationAbsolute",
    "PriceModifier",
    "ShippingRegion",
    "Payee",
    "Manifest",
]


# construct default encoder
def cbor_encode(obj):
    with BytesIO() as fp:
        cbor2.CBOREncoder(
            fp,
            canonical=True,
            date_as_datetime=True,
        ).encode(obj)
        return fp.getvalue()


@dataclass
class PublicKey:
    key: bytes

    def to_cbor_dict(self) -> dict:
        return self.key

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "PublicKey":
        return cls(d)

    def __hash__(self):
        return hash(self.key)

    def __post_init__(self):
        if len(self.key) != 33:
            raise ValueError("PublicKey must be 33 bytes but got %d" % len(self.key))


@dataclass
class Account:
    keycards: List[PublicKey]
    guest: bool

    def to_cbor_dict(self) -> dict:
        return {
            "KeyCards": [k.to_cbor_dict() for k in self.keycards],
            "Guest": self.guest,
        }

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Account":
        return cls(
            keycards=[PublicKey(k) for k in d["KeyCards"]],
            guest=d["Guest"],
        )

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "Account":
        d = cbor2.loads(cbor_data)
        return cls.from_cbor_dict(d)


# manifest and friends
@dataclass
class ModificationAbsolute:
    amount: Uint256
    plus: bool

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "ModificationAbsolute":
        return cls(
            amount=Uint256(d["Amount"]),
            plus=d["Plus"],
        )

    def to_cbor_dict(self) -> dict:
        return {
            "Amount": self.amount.to_cbor_dict(),
            "Plus": self.plus,
        }


@dataclass
class PriceModifier:
    modification_percents: Optional[Uint256] = None
    modification_absolute: Optional[ModificationAbsolute] = None

    def __post_init__(self):
        if self.modification_percents is None and self.modification_absolute is None:
            raise ValueError(
                "One of modification_percents or modification_absolute must be set"
            )
        if (
            self.modification_percents is not None
            and self.modification_absolute is not None
        ):
            raise ValueError(
                "Only one of modification_percents or modification_absolute can be set"
            )

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "PriceModifier":
        mp = d.get("ModificationPrecents")
        ma = d.get("ModificationAbsolute")
        if ma is not None and isinstance(ma, dict):
            ma = ModificationAbsolute.from_cbor_dict(ma)
        return cls(
            modification_percents=mp,
            modification_absolute=ma,
        )

    def to_cbor_dict(self) -> dict:
        d = {}
        if self.modification_percents is not None:
            d["ModificationPrecents"] = self.modification_percents
        if self.modification_absolute is not None:
            d["ModificationAbsolute"] = self.modification_absolute.to_cbor_dict()
        return d


@dataclass
class ShippingRegion:
    country: str
    postcode: str
    city: str
    price_modifiers: Optional[Dict[str, PriceModifier]] = None

    def __post_init__(self):
        if self.price_modifiers is not None and not self.price_modifiers:
            raise ValueError("PriceModifiers map cannot be empty if present")

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "ShippingRegion":
        pm = d.get("PriceModifiers")
        if pm is not None:
            pm = {k: PriceModifier.from_cbor_dict(v) for k, v in pm.items()}
        return cls(
            country=d["Country"],
            postcode=d["Postcode"],
            city=d["City"],
            price_modifiers=pm,
        )

    def to_cbor_dict(self) -> dict:
        d = {
            "Country": self.country,
            "Postcode": self.postcode,
            "City": self.city,
        }
        if self.price_modifiers is not None:
            d["PriceModifiers"] = {
                k: v.to_cbor_dict() for k, v in self.price_modifiers.items()
            }
        return d


@dataclass
class Payee:
    address: ChainAddress
    call_as_contract: bool

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Payee":
        return cls(
            address=ChainAddress.from_cbor_dict(d["Address"]),
            call_as_contract=d["CallAsContract"],
        )

    def to_cbor_dict(self) -> dict:
        return {
            "Address": (
                self.address.to_cbor_dict()
                if hasattr(self.address, "to_cbor_dict")
                else self.address
            ),
            "CallAsContract": self.call_as_contract,
        }


@dataclass
class Manifest:
    shop_id: Uint256
    payees: Dict[str, Payee]
    accepted_currencies: List[ChainAddress]
    pricing_currency: ChainAddress
    shipping_regions: Optional[Dict[str, ShippingRegion]] = None

    def __post_init__(self):
        if not self.payees:
            raise ValueError("Payees map cannot be empty")
        if not self.accepted_currencies:
            raise ValueError("AcceptedCurrencies must not be empty")
        if self.shipping_regions is not None and not self.shipping_regions:
            raise ValueError("ShippingRegions map cannot be empty if present")

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Manifest":
        payees = {k: Payee.from_cbor_dict(v) for k, v in d["Payees"].items()}
        accepted_currencies = [
            (
                ChainAddress.from_cbor_dict(item)
                if isinstance(item, dict) and hasattr(ChainAddress, "from_cbor_dict")
                else item
            )
            for item in d["AcceptedCurrencies"]
        ]
        pricing_currency = (
            ChainAddress.from_cbor_dict(d["PricingCurrency"])
            if isinstance(d["PricingCurrency"], dict)
            and hasattr(ChainAddress, "from_cbor_dict")
            else d["PricingCurrency"]
        )
        shipping_regions = None
        if "ShippingRegions" in d and d["ShippingRegions"] is not None:
            shipping_regions = {
                k: ShippingRegion.from_cbor_dict(v)
                for k, v in d["ShippingRegions"].items()
            }
        return cls(
            shop_id=d["ShopID"],
            payees=payees,
            accepted_currencies=accepted_currencies,
            pricing_currency=pricing_currency,
            shipping_regions=shipping_regions,
        )

    def to_cbor_dict(self) -> dict:
        d = {
            "ShopID": self.shop_id.to_cbor_dict(),
            "Payees": {k: v.to_cbor_dict() for k, v in self.payees.items()},
            "AcceptedCurrencies": [
                item.to_cbor_dict() if hasattr(item, "to_cbor_dict") else item
                for item in self.accepted_currencies
            ],
            "PricingCurrency": (
                self.pricing_currency.to_cbor_dict()
                if hasattr(self.pricing_currency, "to_cbor_dict")
                else self.pricing_currency
            ),
        }
        if self.shipping_regions is not None:
            d["ShippingRegions"] = {
                k: v.to_cbor_dict() for k, v in self.shipping_regions.items()
            }
        return d

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "Manifest":
        d = cbor2.loads(cbor_data)
        return cls.from_cbor_dict(d)
