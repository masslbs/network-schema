# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import json
import random
import string
import xxhash
from massmarket_hash_event.hamt import Trie
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
        entries = []

        def collect_entries(node, prefix=""):
            for e in node.entries:
                if e.node is None:
                    entries.append(
                        {
                            "type": "leaf",
                            "key": e.key.hex() if e.key else None,
                            "value": e.value,
                        }
                    )
                else:
                    entries.append(
                        {
                            "type": "branch",
                            "entries": collect_entries(e.node, prefix + "  "),
                        }
                    )
            return entries

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

    return test_vectors


if __name__ == "__main__":
    vectors = generate_test_vectors()
    with open("../vectors/hamt_test.json", "w") as f:
        json.dump(vectors, f, indent=2)
