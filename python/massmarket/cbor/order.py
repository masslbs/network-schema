# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

from dataclasses import dataclass
from typing import Optional, List, Dict, Any
from datetime import datetime
from enum import IntEnum

import cbor2

from massmarket.cbor.base_types import Uint256, ChainAddress, Payee


class OrderState(IntEnum):
    UNSPECIFIED = 0
    OPEN = 1
    CANCELED = 2
    COMMITTED = 3
    PAYMENT_CHOSEN = 4
    UNPAID = 5
    PAID = 6


@dataclass
class OrderedItem:
    listing_id: int
    quantity: int
    variation_ids: Optional[List[str]] = None

    def __post_init__(self):
        if self.listing_id <= 0:
            raise ValueError("ListingID must be greater-or-equal to 0")
        if self.quantity < 0:
            raise ValueError("Quantity must be greater than 0")

    @classmethod
    def from_cbor_dict(cls, d: Dict[str, Any]) -> "OrderedItem":
        return cls(
            listing_id=d["ListingID"],
            quantity=d["Quantity"],
            variation_ids=d.get("VariationIDs"),
        )

    def to_cbor_dict(self) -> Dict[str, Any]:
        d = {
            "ListingID": self.listing_id,
            "Quantity": self.quantity,
        }
        if self.variation_ids is not None:
            d["VariationIDs"] = self.variation_ids
        return d


@dataclass
class AddressDetails:
    name: str
    address1: str
    city: str
    country: str
    email_address: str
    address2: Optional[str] = None
    postal_code: Optional[str] = None
    phone_number: Optional[str] = None

    @classmethod
    def from_cbor_dict(cls, d: Dict[str, Any]) -> "AddressDetails":
        return cls(
            name=d["Name"],
            address1=d["Address1"],
            address2=d.get("Address2"),
            city=d["City"],
            postal_code=d.get("PostalCode"),
            country=d["Country"],
            email_address=d["EmailAddress"],
            phone_number=d.get("PhoneNumber"),
        )

    def to_cbor_dict(self) -> Dict[str, Any]:
        d = {
            "Name": self.name,
            "Address1": self.address1,
            "City": self.city,
            "Country": self.country,
            "EmailAddress": self.email_address,
        }
        if self.address2 is not None:
            d["Address2"] = self.address2
        if self.postal_code is not None:
            d["PostalCode"] = self.postal_code
        if self.phone_number is not None:
            d["PhoneNumber"] = self.phone_number
        return d


@dataclass
class PaymentDetails:
    payment_id: bytes  # Hash
    total: Uint256
    listing_hashes: List[bytes]
    ttl: int
    shop_signature: bytes  # Signature

    def __post_init__(self):
        if not self.listing_hashes:
            raise ValueError("ListingHashes must not be empty")
        if self.ttl <= 0:
            raise ValueError("TTL must be greater than 0")
        if len(self.payment_id) != 32:  # HashSize in Go code
            raise ValueError(f"PaymentID must be 32 bytes, got {len(self.payment_id)}")
        if len(self.shop_signature) != 65:  # SignatureSize in Go code
            raise ValueError(
                f"ShopSignature must be 65 bytes, got {len(self.shop_signature)}"
            )

    @classmethod
    def from_cbor_dict(cls, d: Dict[str, Any]) -> "PaymentDetails":
        return cls(
            payment_id=d["PaymentID"],
            total=Uint256(d["Total"]),
            listing_hashes=d["ListingHashes"],
            ttl=d["TTL"],
            shop_signature=d["ShopSignature"],
        )

    def to_cbor_dict(self) -> Dict[str, Any]:
        return {
            "PaymentID": self.payment_id,
            "Total": self.total.to_cbor_dict(),
            "ListingHashes": self.listing_hashes,
            "TTL": self.ttl,
            "ShopSignature": self.shop_signature,
        }


@dataclass
class OrderPaid:
    block_hash: bytes  # Hash
    tx_hash: Optional[bytes] = None  # Hash

    def __post_init__(self):
        if len(self.block_hash) != 32:  # HashSize in Go code
            raise ValueError(f"BlockHash must be 32 bytes, got {len(self.block_hash)}")
        if self.tx_hash is not None and len(self.tx_hash) != 32:
            raise ValueError(f"TxHash must be 32 bytes, got {len(self.tx_hash)}")

    @classmethod
    def from_cbor_dict(cls, d: Dict[str, Any]) -> "OrderPaid":
        if "BlockHash" not in d:
            raise ValueError("BlockHash must be set")
        return cls(
            block_hash=d["BlockHash"],
            tx_hash=d.get("TxHash"),
        )

    def to_cbor_dict(self) -> Dict[str, Any]:
        d = {"BlockHash": self.block_hash}
        if self.tx_hash is not None:
            d["TxHash"] = self.tx_hash
        return d


@dataclass
class Order:
    id: int
    items: List[OrderedItem]
    state: OrderState
    invoice_address: Optional[AddressDetails] = None
    shipping_address: Optional[AddressDetails] = None
    canceled_at: Optional[datetime] = None
    chosen_payee: Optional[Payee] = None
    chosen_currency: Optional[ChainAddress] = None
    payment_details: Optional[PaymentDetails] = None
    tx_details: Optional[OrderPaid] = None

    def __post_init__(self):
        # Validate state-specific requirements
        if self.state == OrderState.PAID:
            if self.tx_details is None:
                raise ValueError("TxDetails is required when state is PAID")

        if self.state in (OrderState.PAID, OrderState.UNPAID):
            if self.payment_details is None:
                raise ValueError(
                    "PaymentDetails is required when state is UNPAID or PAID"
                )

        if self.state in (OrderState.PAID, OrderState.UNPAID, OrderState.COMMITTED):
            if self.chosen_payee is None:
                raise ValueError(
                    "ChosenPayee is required when state is COMMITTED, UNPAID, or PAID"
                )
            if self.chosen_currency is None:
                raise ValueError(
                    "ChosenCurrency is required when state is COMMITTED, UNPAID, or PAID"
                )
            if self.invoice_address is None and self.shipping_address is None:
                raise ValueError(
                    "Either InvoiceAddress or ShippingAddress is required for COMMITTED, UNPAID, or PAID states"
                )

        if self.state == OrderState.CANCELED:
            if self.canceled_at is None:
                raise ValueError("CanceledAt is required when state is CANCELED")

    @classmethod
    def from_cbor_dict(cls, d: Dict[str, Any]) -> "Order":
        items = [OrderedItem.from_cbor_dict(item) for item in d["Items"]]

        invoice_address = d.get("InvoiceAddress")
        if invoice_address is not None:
            invoice_address = AddressDetails.from_cbor_dict(invoice_address)

        shipping_address = d.get("ShippingAddress")
        if shipping_address is not None:
            shipping_address = AddressDetails.from_cbor_dict(shipping_address)

        chosen_payee = d.get("ChosenPayee")
        if chosen_payee is not None:
            chosen_payee = Payee.from_cbor_dict(chosen_payee)

        chosen_currency = d.get("ChosenCurrency")
        if chosen_currency is not None:
            chosen_currency = ChainAddress.from_cbor_dict(chosen_currency)

        payment_details = d.get("PaymentDetails")
        if payment_details is not None:
            payment_details = PaymentDetails.from_cbor_dict(payment_details)

        tx_details = d.get("TxDetails")
        if tx_details is not None:
            tx_details = OrderPaid.from_cbor_dict(tx_details)

        return cls(
            id=d["ID"],
            items=items,
            state=OrderState(d["State"]),
            invoice_address=invoice_address,
            shipping_address=shipping_address,
            canceled_at=d.get("CanceledAt"),
            chosen_payee=chosen_payee,
            chosen_currency=chosen_currency,
            payment_details=payment_details,
            tx_details=tx_details,
        )

    def to_cbor_dict(self) -> Dict[str, Any]:
        d = {
            "ID": self.id,
            "Items": [item.to_cbor_dict() for item in self.items],
            # TODO: why isnt this tested..?
            "State": (
                self.state.value if isinstance(self.state, OrderState) else self.state
            ),
        }

        if self.invoice_address is not None:
            d["InvoiceAddress"] = self.invoice_address.to_cbor_dict()
        if self.shipping_address is not None:
            d["ShippingAddress"] = self.shipping_address.to_cbor_dict()
        if self.canceled_at is not None:
            d["CanceledAt"] = self.canceled_at
        if self.chosen_payee is not None:
            d["ChosenPayee"] = self.chosen_payee.to_cbor_dict()
        if self.chosen_currency is not None:
            d["ChosenCurrency"] = self.chosen_currency.to_cbor_dict()
        if self.payment_details is not None:
            d["PaymentDetails"] = self.payment_details.to_cbor_dict()
        if self.tx_details is not None:
            d["TxDetails"] = self.tx_details.to_cbor_dict()

        return d

    @classmethod
    def from_cbor(cls, cbor_data: bytes) -> "Order":
        d = cbor2.loads(cbor_data)
        return cls.from_cbor_dict(d)
