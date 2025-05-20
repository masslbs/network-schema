# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import cbor2

from massmarket.cbor.base_types import Account, PublicKey


def test_cbor_account():
    pk_data = b"\xac\xab" * 11 + b"\x00" * 11
    account = Account(
        keycards=[PublicKey(pk_data)],
        guest=False,
    )
    cbor_dict = account.to_cbor_dict()
    assert cbor_dict["KeyCards"] == [pk_data]
    assert cbor_dict["Guest"] == False

    account2 = Account.from_cbor_dict(cbor_dict)
    assert account == account2

    # test round trip
    cbor_data = cbor2.dumps(account.to_cbor_dict())
    account3 = Account.from_cbor_dict(cbor2.loads(cbor_data))
    assert account == account3


def test_cbor_public_key():
    pk_data = b"\xac\xab" * 11 + b"\x00" * 11
    pk = PublicKey(pk_data)
    assert pk.key == pk_data
    assert pk.to_cbor_dict() == pk_data
    assert PublicKey.from_cbor_dict(pk.to_cbor_dict()) == pk

    # Test using PublicKey as dict key
    pk1_data = b"\xac\xab" * 11 + b"\x00" * 11
    pk2_data = b"\xcd\xef" * 11 + b"\x11" * 11

    pk1 = PublicKey(pk1_data)
    pk2 = PublicKey(pk2_data)

    # Create dict with PublicKey keys
    pk_dict = {pk1: "value1", pk2: "value2"}

    # Test lookup works
    assert pk_dict[pk1] == "value1"
    assert pk_dict[pk2] == "value2"

    # Test same key data maps to same value
    pk1_copy = PublicKey(pk1_data)
    assert pk_dict[pk1_copy] == "value1"
