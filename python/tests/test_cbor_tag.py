# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import cbor2

from massmarket.cbor.base_types import Tag


def test_cbor_tag_roundtrip():
    tag = Tag(name="shoes", listings=[1, 2, 3])
    assert tag.to_cbor_dict() == {
        "Name": "shoes",
        "Listings": [1, 2, 3],
    }

    cbor_data = cbor2.dumps(tag.to_cbor_dict())
    tag2 = Tag.from_cbor_dict(cbor2.loads(cbor_data))

    assert tag2.name == "shoes"
    assert tag2.listings == [1, 2, 3]
