# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import cbor2
import random
import string
from massmarket.hamt import Trie
from typing import List, Dict, Any


def random_bytes(length: int) -> bytes:
    return "".join(
        random.choices(string.ascii_letters + string.digits, k=length)
    ).encode()


def generate_test_vector(operations: List[Dict[str, Any]]) -> Dict[str, Any]:
    trie = Trie.new()
    hashes = []
    states = []

    for op in operations:
        if op["type"] == "insert":
            trie.insert(bytes.fromhex(op["key"]), op["value"])
        elif op["type"] == "delete":
            trie.delete(bytes.fromhex(op["key"]))

        # Collect state information
        def collect_entries(node, prefix=""):
            collected = []  # Create new list for each recursion level
            for e in node.entries:
                if e.node is None:
                    collected.append(
                        {
                            "type": "leaf",
                            "key": e.key.hex() if e.key else None,
                            "value": e.value,
                        }
                    )
                else:
                    # Create a new dictionary for each branch
                    collected.append(
                        {
                            "type": "branch",
                            "entries": collect_entries(e.node, prefix + "  "),
                        }
                    )
            return collected  # Return new list instead of modifying global entries

        states.append(
            {
                "bitmap": trie.root.bitmap,
                "entries": collect_entries(trie.root),
            }
        )
        hashes.append(trie.hash().hex())

    return {"operations": operations, "hashes": hashes, "states": states}


def generate_test_vectors() -> List[Dict[str, Any]]:
    test_vectors = []

    # Test vector 1: Simple insertions
    ops = [
        {"type": "insert", "key": b"key1".hex(), "value": "value1"},
        {"type": "insert", "key": b"key2".hex(), "value": "value2"},
        {"type": "insert", "key": b"key3".hex(), "value": "value3"},
    ]
    test_vectors.append(generate_test_vector(ops))

    # Test vector 2: Insertions and deletions
    ops = [
        {"type": "insert", "key": b"key1".hex(), "value": "value1"},
        {"type": "insert", "key": b"key2".hex(), "value": "value2"},
        {"type": "delete", "key": b"key1".hex()},
        {"type": "insert", "key": b"key3".hex(), "value": "value3"},
    ]
    test_vectors.append(generate_test_vector(ops))

    # Test vector 3: Overwrite values
    ops = [
        {"type": "insert", "key": b"key1".hex(), "value": "value1"},
        {"type": "insert", "key": b"key1".hex(), "value": "value2"},
        {"type": "insert", "key": b"key1".hex(), "value": "value3"},
    ]
    test_vectors.append(generate_test_vector(ops))

    # Test vector 4: Random operations
    ops = []
    keys = [random_bytes(4).hex() for _ in range(5)]
    for _ in range(10):
        if random.random() < 0.7:  # 70% chance of insert
            key = random.choice(keys)
            ops.append(
                {
                    "type": "insert",
                    "key": key,
                    "value": f"value{random.randint(1, 100)}",
                }
            )
        else:  # 30% chance of delete
            key = random.choice(keys)
            ops.append({"type": "delete", "key": key})
    test_vectors.append(generate_test_vector(ops))

    # Test vector 5: Large set of insertions to produce some branching and depth
    ops = [
        {
            "type": "insert",
            "key": random_bytes(4).hex(),
            "value": f"value{random.randint(1, 1000)}",
        }
        for _ in range(1000)
    ]
    test_vectors.append(generate_test_vector(ops))

    # Test vector 6: Mixed operations with large set of data
    ops = []
    keys = [random_bytes(4).hex() for _ in range(1000)]
    for _ in range(100):
        if random.random() < 0.7:  # 70% chance of insert
            key = random.choice(keys)
            ops.append(
                {
                    "type": "insert",
                    "key": key,
                    "value": f"value{random.randint(1, 1000)}",
                }
            )
        else:  # 30% chance of delete
            key = random.choice(keys)
            ops.append({"type": "delete", "key": key})
    test_vectors.append(generate_test_vector(ops))

    return test_vectors


if __name__ == "__main__":
    vectors = generate_test_vectors()
    import os

    test_data_out = os.getenv("TEST_DATA_OUT")
    if test_data_out is None:
        test_data_out = "../vectors"

    os.makedirs(test_data_out, exist_ok=True)
    fname = os.path.join(test_data_out, "hamt_test.cbor")
    with open(fname, "wb") as f:
        cbor2.dump(vectors, f)
