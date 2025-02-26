# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import json
import os
import base64
import cbor2
from pprint import pprint

from massmarket_hash_event.hamt import Trie
from massmarket_hash_event.cbor import Shop


def test_hamt_standalone_vectors():
    vectors_path = "../vectors/hamt_test.json"
    with open(vectors_path) as f:
        test_vectors = json.load(f)

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
    files = [f for f in os.listdir("../vectors") if f.endswith("Okay.json")]
    for file in files:
        with open(os.path.join("../vectors", file)) as f:
            vectors = json.load(f)

        # print(f"Testing {file}")
        for snap in vectors["Snapshots"]:
            # print(f"  Snapshot: {snap['Name']}")
            cbor_data = base64.b64decode(snap["After"]["Encoded"])

            shop_dict = cbor2.loads(cbor_data)

            # check the top-level hamts for sanity
            hamt_keys = ["Accounts", "Orders", "Listings", "Tags", "Inventory"]
            for key in hamt_keys:
                if key not in shop_dict:
                    print(f"  {key} not found in file {file}/snapshot {snap['Name']}")
                    continue

                hamt_data = shop_dict.get(key, None)
                trie = Trie.from_cbor_array(hamt_data)
                print(f"    {key}: {trie.hash().hex()}")
                # assert trie.hash().hex() == snap["After"][key]["Hash"]

            # load the full shop and compare the hashes
            shop = Shop.from_cbor_dict(shop_dict)
            assert shop.hash().hex() == snap["After"]["Hash"]

    # assert False
