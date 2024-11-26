# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import cbor2
import random

from massmarket.hamt import Trie


def test_basic_hamt():
    trie = Trie.new()

    # Insert some values
    trie.insert(b"name", "Alice")
    trie1 = trie.copy()

    trie1.insert(b"age", "Bob")
    assert trie1.size == 2

    # Verify insertions
    val = trie1.get(b"name")
    assert val is not None
    assert val == "Alice"

    val = trie1.get(b"age")
    assert val is not None
    assert val == "Bob"

    # Original trie should be unchanged
    val = trie.get(b"age")
    assert val is None

    # Should also work with literal types, like int
    trie2 = Trie.new()
    trie2.insert(b"age", 1)
    val_int = trie2.get(b"age")
    assert val_int is not None
    assert val_int == 1


def test_complex_operations():
    values = [cbor2.dumps(i) for i in range(5)]
    new_values = [cbor2.dumps(f"new-{i}") for i in range(5)]

    # Create initial trie with multiple values
    trie = Trie.new()
    trie.insert(b"a", values[0])
    trie.insert(b"b", values[1])
    trie.insert(b"c", values[2])
    trie.insert(b"d", values[3])
    assert trie.size == 4

    # Test replacing existing values
    trie2 = trie.copy()
    trie2.insert(b"b", new_values[1])
    assert trie2.size == 4

    # Verify original is unchanged
    val = trie.get(b"b")
    assert val is not None
    assert val == values[1]

    # Verify new value in new trie
    val = trie2.get(b"b")
    assert val is not None
    assert val == new_values[1]

    # Test deleting values
    trie3 = trie2.copy()
    trie3.delete(b"a")
    assert trie3.size == 3

    # Verify deletion
    val = trie3.get(b"a")
    assert val is None

    # Other values should remain
    val = trie3.get(b"b")
    assert val is not None
    assert val == new_values[1]

    val = trie3.get(b"c")
    assert val is not None
    assert val == values[2]


def test_trie_hash():
    # Empty trie should have consistent hash
    trie1 = Trie.new()
    hash1 = trie1.hash()
    assert hash1 is not None

    # Same empty trie should have same hash
    trie2 = Trie.new()
    hash2 = trie2.hash()
    assert hash1 == hash2

    # Adding elements should change hash
    trie1.insert(b"a", "1")
    hash3 = trie1.hash()
    assert hash1 != hash3

    # Same elements added in same order should have same hash
    trie1.insert(b"a", "1")
    hash4 = trie1.hash()
    assert hash3 == hash4

    # Different elements should have different hashes
    trie5 = Trie.new()
    trie5.insert(b"a", "1")
    trie5.insert(b"b", "2")
    hash5 = trie5.hash()
    assert hash3 != hash5

    # Order of insertion shouldn't matter
    trie6 = Trie.new()
    trie6.insert(b"a", "1")
    trie6.insert(b"b", "2")

    trie7 = Trie.new()
    trie7.insert(b"b", "2")
    trie7.insert(b"a", "1")

    hash6 = trie6.hash()
    hash7 = trie7.hash()
    assert hash6 == hash7


def test_trie_size_tracking():
    trie = Trie.new()
    assert trie.size == 0

    # Insert new keys
    trie.insert(b"a", "1")
    assert trie.size == 1

    trie.insert(b"b", "2")
    assert trie.size == 2

    # Update existing key
    trie.insert(b"a", "updated-1")
    assert trie.size == 2  # Size should not change

    # Delete existing key
    trie.delete(b"a")
    assert trie.size == 1

    # Delete non-existent key
    trie.delete(b"non-existent")
    assert trie.size == 1  # Size should not change


def test_trie_iterator():
    trie = Trie.new()

    # Empty trie should not call function
    called = False

    def callback(k: bytes, v: str) -> bool:
        nonlocal called
        called = True
        return True

    trie.all(callback)
    assert not called

    # Insert some test data
    test_data = {
        b"a": "value-a",
        b"b": "value-b",
        b"c": "value-c",
        b"d": "value-d",
    }

    for k, v in test_data.items():
        trie.insert(k, v)

    # Collect all entries via iterator
    got_values = []

    def collect(k: bytes, v: str) -> bool:
        got_values.append((k, v))
        return True

    trie.all(collect)

    # Should visit all entries
    assert len(got_values) == len(test_data)

    # Values should match
    for k, v in got_values:
        assert test_data[k] == v

    # Early termination
    count = 0

    def early_stop(k: bytes, v: str) -> bool:
        nonlocal count
        count += 1
        return count < 2  # Stop after first entry

    trie.all(early_stop)
    assert count == 2


def test_large_scale_operations():
    trie = Trie.new()
    num_elements = 100_000
    keys = []
    values = []

    # Insert a large number of elements
    for i in range(num_elements):
        key = f"key-{i}".encode()
        value = f"value-{i}"
        keys.append(key)
        values.append(value)
        trie.insert(key, value)

    assert trie.size == num_elements

    # Verify that all elements can be retrieved
    for key, expected_value in zip(keys, values):
        val = trie.get(key)
        assert val is not None
        assert val == expected_value

    # Delete every other element
    for i in range(0, num_elements, 2):
        trie.delete(keys[i])

    # Verify that the correct elements have been deleted
    for i, (key, expected_value) in enumerate(zip(keys, values)):
        val = trie.get(key)
        if i % 2 == 0:
            assert val is None
        else:
            assert val is not None
            assert val == expected_value

    assert trie.size == num_elements // 2


def test_hash_order_independence():
    # Create a map of key-value pairs
    items = {str(i).encode(): f"value{i}" for i in range(1, 11)}

    num_tries = 1000
    first_hash = None

    # Insert items in different orders
    for i in range(num_tries):
        trie = Trie.new()
        # Convert items to list and shuffle for random order
        items_list = list(items.items())
        random.shuffle(items_list)

        for k, v in items_list:
            trie.insert(k, v)

        current_hash = trie.hash()
        if i == 0:
            first_hash = current_hash
        else:
            assert current_hash == first_hash


def test_key_type_support():
    trie = Trie.new()

    # Test integer keys
    trie.insert(42, "int-value")
    val = trie.get(42)
    assert val is not None
    assert val == "int-value"

    # Test string keys
    trie.insert("hello", "str-value")
    val = trie.get("hello")
    assert val is not None
    assert val == "str-value"

    # Verify size
    assert trie.size == 2

    # Test deletion
    trie.delete(42)
    val = trie.get(42)
    assert val is None
    assert trie.size == 1
