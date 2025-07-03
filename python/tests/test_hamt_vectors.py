# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import json
import base64
import cbor2

from massmarket.hamt import Trie
from massmarket.cbor import Shop
from utils import get_vector_file, list_vector_files


def test_hamt_standalone_vectors():
    vectors_path = get_vector_file("hamt_test.cbor")
    with open(vectors_path, "rb") as f:
        test_vectors = cbor2.load(f)

    for i, vector in enumerate(test_vectors):
        trie = Trie.new()
        operations = vector["operations"]
        expected_hashes = vector["hashes"]

        for j, op in enumerate(operations):
            if op["type"] == "insert":
                trie.insert(bytes.fromhex(op["key"]), op["value"])
            elif op["type"] == "delete":
                trie.delete(bytes.fromhex(op["key"]))

            actual_hash = trie.hash().hex()
            assert (
                actual_hash == expected_hashes[j]
            ), f"Test vector {i}, operation {j}: hash mismatch"


def test_hamt_shop_vectors():
    files = list_vector_files("Okay.json")
    assert len(files) > 0
    for file in files:
        with open(get_vector_file(file)) as f:
            vectors = json.load(f)

        # print(f"Testing {file}")
        for snap in vectors["Snapshots"]:
            # print(f"  Snapshot: {snap['Name']}")
            print(f"  Snapshot: {snap['Name']}")
            cbor_data = base64.b64decode(snap["After"]["Encoded"])

            shop_dict = cbor2.loads(cbor_data)

            # load the full shop and compare the hashes
            shop = Shop.from_cbor_dict(shop_dict)
            expected_hash = base64.b64decode(snap["After"]["Hash"])
            assert shop.hash().hex() == expected_hash.hex()
