# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import json
from massmarket_hash_event.hamt import Trie


def test_hamt_vectors():
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
