from typing import Dict, Optional, List
from dataclasses import dataclass

import cbor2

from massmarket_hash_event.cbor.uint256 import Uint256
from massmarket_hash_event.cbor.chain_address import ChainAddress
from massmarket_hash_event.cbor.base_types import ShippingRegion


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
