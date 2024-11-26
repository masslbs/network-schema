# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import pytest
import os
import json
import base64

import cbor2

from massmarket.cbor.base_types import (
    Uint256,
    ChainAddress,
    Payee,
)
from massmarket.cbor import (
    cbor_encode,
    Shop,
)

from massmarket.cbor.order import (
    OrderState,
    OrderedItem,
    AddressDetails,
    PaymentDetails,
    OrderPaid,
    Order,
)


def test_ordered_item_roundtrip():
    item = OrderedItem(
        listing_id=5555,
        quantity=3,
        variation_ids=["red", "large"],
    )

    encoded = cbor_encode(item.to_cbor_dict())
    decoded = OrderedItem.from_cbor_dict(cbor2.loads(encoded))

    assert decoded.listing_id == item.listing_id
    assert decoded.quantity == item.quantity
    assert decoded.variation_ids == item.variation_ids


def test_address_details_roundtrip():
    address = AddressDetails(
        name="John Doe",
        address1="123 Main St",
        city="Anytown",
        country="US",
        email_address="john.doe@example.com",
        address2="Apt 101",
        postal_code="12345",
        phone_number="555-123-4567",
    )

    encoded = cbor_encode(address.to_cbor_dict())
    decoded = AddressDetails.from_cbor_dict(cbor2.loads(encoded))

    assert decoded.name == address.name
    assert decoded.address1 == address.address1
    assert decoded.address2 == address.address2
    assert decoded.city == address.city
    assert decoded.postal_code == address.postal_code
    assert decoded.country == address.country
    assert decoded.email_address == address.email_address
    assert decoded.phone_number == address.phone_number


def test_payment_details_roundtrip():
    payment_details = PaymentDetails(
        payment_id=b"\x01" * 32,
        total=Uint256(12345),
        listing_hashes=[b"\x02" * 32, b"\x03" * 32],
        ttl=3600,
        shop_signature=b"\x04" * 65,
    )

    encoded = cbor_encode(payment_details.to_cbor_dict())
    decoded = PaymentDetails.from_cbor_dict(cbor2.loads(encoded))

    assert decoded.payment_id == payment_details.payment_id
    assert decoded.total == payment_details.total
    assert decoded.listing_hashes == payment_details.listing_hashes
    assert decoded.ttl == payment_details.ttl
    assert decoded.shop_signature == payment_details.shop_signature


def test_order_paid_roundtrip():
    order_paid = OrderPaid(
        block_hash=b"\x05" * 32,
        tx_hash=b"\x06" * 32,
    )

    encoded = cbor_encode(order_paid.to_cbor_dict())
    decoded = OrderPaid.from_cbor_dict(cbor2.loads(encoded))

    assert decoded.block_hash == order_paid.block_hash
    assert decoded.tx_hash == order_paid.tx_hash

    # Test with tx_hash as None
    order_paid_no_tx = OrderPaid(
        block_hash=b"\x05" * 32,
    )
    encoded = cbor_encode(order_paid_no_tx.to_cbor_dict())
    decoded = OrderPaid.from_cbor_dict(cbor2.loads(encoded))

    assert decoded.block_hash == order_paid_no_tx.block_hash
    assert decoded.tx_hash is None


def test_full_order_roundtrip():
    address = AddressDetails(
        name="John Doe",
        address1="123 Main St",
        city="Anytown",
        country="US",
        email_address="john.doe@example.com",
    )

    payee = Payee(
        address=ChainAddress(
            chain_id=1337,
            address=b"\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f\x10\x11\x12\x13\x14",
        ),
        call_as_contract=False,
    )

    currency = ChainAddress(chain_id=1337, address=b"\x00" * 20)

    payment_details = PaymentDetails(
        payment_id=b"\x01" * 32,
        total=Uint256(12345),
        listing_hashes=[b"\x02" * 32],
        ttl=3600,
        shop_signature=b"\x04" * 65,
    )

    order = Order(
        id=666,
        items=[
            OrderedItem(
                listing_id=5555,
                quantity=3,
                variation_ids=["red"],
            )
        ],
        state=OrderState.COMMITTED,
        invoice_address=address,
        chosen_payee=payee,
        chosen_currency=currency,
        payment_details=payment_details,
    )

    encoded = cbor_encode(order.to_cbor_dict())
    decoded = Order.from_cbor_dict(cbor2.loads(encoded))

    assert decoded.id == order.id
    assert len(decoded.items) == len(order.items)
    assert decoded.items[0].listing_id == order.items[0].listing_id
    assert decoded.items[0].quantity == order.items[0].quantity
    assert decoded.state == order.state
    assert decoded.invoice_address.name == order.invoice_address.name
    assert decoded.chosen_payee.address.chain_id == order.chosen_payee.address.chain_id
    assert decoded.chosen_currency.chain_id == order.chosen_currency.chain_id
    assert decoded.payment_details.total == order.payment_details.total


def test_order_validation():
    # Test missing chosen_payee for COMMITTED state
    with pytest.raises(
        ValueError, match="ChosenPayee is required when state is COMMITTED"
    ):
        Order(
            id=1,
            items=[OrderedItem(listing_id=5555, quantity=1)],
            state=OrderState.COMMITTED,
            chosen_currency=ChainAddress(chain_id=1337, address=b"\x00" * 20),
            invoice_address=AddressDetails(
                name="John",
                address1="123",
                city="City",
                country="US",
                email_address="john@example.com",
            ),
        )

    # Test missing invoice_address or shipping_address for COMMITTED state
    with pytest.raises(
        ValueError, match="Either InvoiceAddress or ShippingAddress is required"
    ):
        Order(
            id=1,
            items=[OrderedItem(listing_id=5555, quantity=1)],
            state=OrderState.COMMITTED,
            chosen_payee=Payee(
                address=ChainAddress(chain_id=1337, address=b"\x00" * 20),
                call_as_contract=False,
            ),
            chosen_currency=ChainAddress(chain_id=1337, address=b"\x00" * 20),
        )

    # Test missing canceled_at for CANCELED state
    with pytest.raises(
        ValueError, match="CanceledAt is required when state is CANCELED"
    ):
        Order(
            id=1,
            items=[OrderedItem(listing_id=5555, quantity=1)],
            state=OrderState.CANCELED,
        )

    # Test missing payment_details for UNPAID state
    with pytest.raises(
        ValueError, match="PaymentDetails is required when state is UNPAID"
    ):
        Order(
            id=1,
            items=[OrderedItem(listing_id=5555, quantity=1)],
            state=OrderState.UNPAID,
            chosen_payee=Payee(
                address=ChainAddress(chain_id=1337, address=b"\x00" * 20),
                call_as_contract=False,
            ),
            chosen_currency=ChainAddress(chain_id=1337, address=b"\x00" * 20),
            invoice_address=AddressDetails(
                name="John",
                address1="123",
                city="City",
                country="US",
                email_address="john@example.com",
            ),
        )

    # Test missing tx_details for PAID state
    with pytest.raises(ValueError, match="TxDetails is required when state is PAID"):
        Order(
            id=1,
            items=[OrderedItem(listing_id=5555, quantity=1)],
            state=OrderState.PAID,
            chosen_payee=Payee(
                address=ChainAddress(chain_id=1337, address=b"\x00" * 20),
                call_as_contract=False,
            ),
            chosen_currency=ChainAddress(chain_id=1337, address=b"\x00" * 20),
            invoice_address=AddressDetails(
                name="John",
                address1="123",
                city="City",
                country="US",
                email_address="john@example.com",
            ),
            payment_details=PaymentDetails(
                payment_id=b"\x01" * 32,
                total=Uint256(100),
                listing_hashes=[b"\x02" * 32],
                ttl=3600,
                shop_signature=b"\x04" * 65,
            ),
        )


def test_order_state():
    assert OrderState.UNSPECIFIED.value == 0
    assert OrderState.OPEN.value == 1
    assert OrderState.CANCELED.value == 2
    assert OrderState.COMMITTED.value == 3
    assert OrderState.PAYMENT_CHOSEN.value == 4
    assert OrderState.UNPAID.value == 5
    assert OrderState.PAID.value == 6


# this does not test the patching logic, just the roundtrip from the _after_ state
def test_order_from_vectors_file():
    file_path = os.path.join(
        os.path.dirname(__file__), "..", "..", "vectors", "OrderOkay.json"
    )
    with open(file_path, "r") as f:
        vectors = json.load(f)

    for snap in vectors["Snapshots"]:
        print(f"Testing {snap['Name']}")
        encoded_b64 = snap["Before"]["Encoded"]
        cbor_data = base64.b64decode(encoded_b64)

        # Decode using our helper
        shop = Shop.from_cbor(cbor_data)

        assert shop.orders.size == len(snap["Before"]["Value"]["Orders"])

        def check_order(order_id, order_dict):
            print(f"Checking order {order_id}")
            from pprint import pprint

            pprint(order_dict)
            order_obj = Order.from_cbor_dict(order_dict)

            expected_orders = snap["Before"]["Value"]["Orders"]
            json_order_id = order_id.hex()
            assert json_order_id in expected_orders
            expected = expected_orders[json_order_id]
            verify_order(order_obj, expected)

            # roundtrip
            cbor_data = cbor_encode(order_obj)
            order_obj_roundtrip = Order.from_cbor(cbor_data)
            assert order_obj_roundtrip.id == order_obj.id
            assert order_obj_roundtrip == order_obj

        shop.orders.all(check_order)
        want_hash = base64.b64decode(snap["Before"]["Hash"])
        got_hash = shop.hash()
        assert want_hash == got_hash


def verify_order(order_obj: Order, expected: dict):
    assert order_obj.id == expected["ID"]

    # Check items
    if "Items" in expected and expected["Items"] is not None:
        if len(expected["Items"]) == 0:
            assert order_obj.items == []
        else:
            assert len(order_obj.items) == len(expected["Items"])
            for i, expected_item in enumerate(expected["Items"]):
                assert order_obj.items[i].listing_id == expected_item["ListingID"]
                assert order_obj.items[i].quantity == expected_item["Quantity"]

                # Check variation_ids if present
                if (
                    "VariationIDs" in expected_item
                    and expected_item["VariationIDs"] is not None
                ):
                    assert (
                        order_obj.items[i].variation_ids
                        == expected_item["VariationIDs"]
                    )
                else:
                    assert order_obj.items[i].variation_ids is None

    # Check state
    if "State" in expected and expected["State"] is not None:
        assert order_obj.state == OrderState(expected["State"])

    # Check invoice address
    if "InvoiceAddress" in expected and expected["InvoiceAddress"] is not None:
        assert order_obj.invoice_address is not None
        assert order_obj.invoice_address.name == expected["InvoiceAddress"]["Name"]
        assert (
            order_obj.invoice_address.address1 == expected["InvoiceAddress"]["Address1"]
        )
        assert order_obj.invoice_address.city == expected["InvoiceAddress"]["City"]
        assert (
            order_obj.invoice_address.country == expected["InvoiceAddress"]["Country"]
        )
        assert (
            order_obj.invoice_address.email_address
            == expected["InvoiceAddress"]["EmailAddress"]
        )

        # Check optional fields
        if (
            "Address2" in expected["InvoiceAddress"]
            and expected["InvoiceAddress"]["Address2"] is not None
            and expected["InvoiceAddress"]["Address2"]
        ):
            assert (
                order_obj.invoice_address.address2
                == expected["InvoiceAddress"]["Address2"]
            )
        else:
            assert order_obj.invoice_address.address2 is None

        if (
            "PostalCode" in expected["InvoiceAddress"]
            and expected["InvoiceAddress"]["PostalCode"] is not None
            and expected["InvoiceAddress"]["PostalCode"]
        ):
            assert (
                order_obj.invoice_address.postal_code
                == expected["InvoiceAddress"]["PostalCode"]
            )
        else:
            assert order_obj.invoice_address.postal_code is None

        if (
            "PhoneNumber" in expected["InvoiceAddress"]
            and expected["InvoiceAddress"]["PhoneNumber"] is not None
            and expected["InvoiceAddress"]["PhoneNumber"]
        ):
            assert (
                order_obj.invoice_address.phone_number
                == expected["InvoiceAddress"]["PhoneNumber"]
            )
        else:
            assert order_obj.invoice_address.phone_number is None
    else:
        assert order_obj.invoice_address is None

    # Check shipping address (similar to invoice address)
    if "ShippingAddress" in expected and expected["ShippingAddress"] is not None:
        assert order_obj.shipping_address is not None
        assert order_obj.shipping_address.name == expected["ShippingAddress"]["Name"]
        assert (
            order_obj.shipping_address.address1
            == expected["ShippingAddress"]["Address1"]
        )
        assert order_obj.shipping_address.city == expected["ShippingAddress"]["City"]
        assert (
            order_obj.shipping_address.country == expected["ShippingAddress"]["Country"]
        )
        assert (
            order_obj.shipping_address.email_address
            == expected["ShippingAddress"]["EmailAddress"]
        )

        # Check optional fields (same as invoice address)
        if (
            "Address2" in expected["ShippingAddress"]
            and expected["ShippingAddress"]["Address2"] is not None
            and expected["ShippingAddress"]["Address2"]
        ):
            assert (
                order_obj.shipping_address.address2
                == expected["ShippingAddress"]["Address2"]
            )
        else:
            assert order_obj.shipping_address.address2 is None

        if (
            "PostalCode" in expected["ShippingAddress"]
            and expected["ShippingAddress"]["PostalCode"] is not None
            and expected["ShippingAddress"]["PostalCode"]
        ):
            assert (
                order_obj.shipping_address.postal_code
                == expected["ShippingAddress"]["PostalCode"]
            )
        else:
            assert order_obj.shipping_address.postal_code is None

        if (
            "PhoneNumber" in expected["ShippingAddress"]
            and expected["ShippingAddress"]["PhoneNumber"] is not None
            and expected["ShippingAddress"]["PhoneNumber"]
        ):
            assert (
                order_obj.shipping_address.phone_number
                == expected["ShippingAddress"]["PhoneNumber"]
            )
        else:
            assert order_obj.shipping_address.phone_number is None
    else:
        assert order_obj.shipping_address is None

    # Check chosen payee
    if "ChosenPayee" in expected and expected["ChosenPayee"] is not None:
        assert order_obj.chosen_payee is not None
        assert (
            order_obj.chosen_payee.address.chain_id
            == expected["ChosenPayee"]["Address"]["ChainID"]
        )
        # Note: Address is binary in our model but hex string in JSON
        if (
            "CallAsContract" in expected["ChosenPayee"]
            and expected["ChosenPayee"]["CallAsContract"] is not None
        ):
            assert (
                order_obj.chosen_payee.call_as_contract
                == expected["ChosenPayee"]["CallAsContract"]
            )
    else:
        assert order_obj.chosen_payee is None

    # Check chosen currency
    if "ChosenCurrency" in expected and expected["ChosenCurrency"] is not None:
        assert order_obj.chosen_currency is not None
        assert (
            order_obj.chosen_currency.chain_id == expected["ChosenCurrency"]["ChainID"]
        )
        # Note: Address is binary in our model but hex string in JSON
    else:
        assert order_obj.chosen_currency is None

    # Check payment details
    if "PaymentDetails" in expected and expected["PaymentDetails"] is not None:
        assert order_obj.payment_details is not None
        assert order_obj.payment_details.ttl == expected["PaymentDetails"]["TTL"]
        # Note: PaymentID, Total, ListingHashes, and ShopSignature would need special handling
        # as they are binary in our model but represented differently in JSON
    else:
        assert order_obj.payment_details is None

    # Check tx details
    if "TxDetails" in expected and expected["TxDetails"] is not None:
        assert order_obj.tx_details is not None
        # Note: BlockHash and TxHash are binary in our model but represented differently in JSON
    else:
        assert order_obj.tx_details is None

    # Check canceled_at
    if (
        "CanceledAt" in expected
        and expected["CanceledAt"] is not None
        and expected["CanceledAt"]
    ):
        assert order_obj.canceled_at is not None
    else:
        assert order_obj.canceled_at is None
