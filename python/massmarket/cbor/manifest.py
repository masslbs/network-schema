# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from typing import Dict, Optional, Set
from dataclasses import dataclass

import cbor2

from massmarket.cbor.base_types import (
    Uint256,
    ChainAddress,
    ShippingRegion,
    PayeeMetadata,
    EthereumAddress,
)


@dataclass
class Manifest:
    shop_id: Uint256
    payees: Dict[int, Dict[EthereumAddress, PayeeMetadata]]
    accepted_currencies: Dict[int, Set[EthereumAddress]]
    pricing_currency: ChainAddress
    shipping_regions: Optional[Dict[str, ShippingRegion]] = None

    @classmethod
    def from_cbor_dict(cls, d: dict) -> "Manifest":
        payees = {
            chainId: {
                EthereumAddress.from_bytes(a): PayeeMetadata.from_cbor_dict(v)
                for a, v in addrs.items()
            }
            for chainId, addrs in d["Payees"].items()
        }
        accepted_currencies = {
            chainId: {EthereumAddress.from_cbor(a) for a in addrs}
            for chainId, addrs in d["AcceptedCurrencies"].items()
        }
        pricing_currency = (
            ChainAddress.from_cbor_dict(d["PricingCurrency"])
            if isinstance(d["PricingCurrency"], dict)
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
            "ShopID": int(self.shop_id),
            "Payees": {
                chainId: {
                    a.to_bytes(): meta.to_cbor_dict() for a, meta in addrs.items()
                }
                for chainId, addrs in self.payees.items()
            },
            "AcceptedCurrencies": {
                chainId: {a.to_bytes(): {} for a in addrs}
                for chainId, addrs in self.accepted_currencies.items()
            },
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
