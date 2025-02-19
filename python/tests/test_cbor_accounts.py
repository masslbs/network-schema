# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import cbor2

from massmarket_hash_event.cbor import Account, PublicKey

def test_cbor_account():
    pk_data = b"\xac\xab"*11 + b"\x00"*11
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
