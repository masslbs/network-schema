# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import hashlib

from dataclasses import dataclass
from typing import TypeVar, Generic, Optional, Any, Callable

from massmarket.cbor_encoder import cbor_encode

BITS_PER_STEP = 6
MAX_DEPTH = 256 // BITS_PER_STEP

V = TypeVar("V")


@dataclass
class HashState:
    original_key: bytes
    hash_buf: bytes  # Store the full 32-byte (256-bit) SHA-256 hash
    consumed: int  # How many bits we've consumed so far

    @classmethod
    def new(cls, key: bytes) -> "HashState":
        # Calculate the full SHA-256 hash once
        h = hashlib.sha256()
        h.update(key)
        hash_buf = h.digest()  # Get the full 32-byte hash

        return cls(original_key=key, hash_buf=hash_buf, consumed=0)

    def next(self) -> int:
        bit_offset = self.consumed
        byte_offset = bit_offset // 8
        bit_in_byte = bit_offset % 8

        next16 = 0
        if byte_offset < 32:
            next16 = self.hash_buf[byte_offset] << 8
        if byte_offset + 1 < 32:
            next16 |= self.hash_buf[byte_offset + 1]

        shift = 16 - BITS_PER_STEP - bit_in_byte
        chunk = (next16 >> shift) & ((1 << BITS_PER_STEP) - 1)
        self.consumed += BITS_PER_STEP
        return chunk


@dataclass
class Entry(Generic[V]):
    key: Optional[bytes]
    value: Optional[V]
    node: Optional["Node[V]"]

    def to_array(self) -> list:
        """
        Convert this entry to an array for compact CBOR serialization:
          [ key, value, node_array or None ]
        """
        if self.node is not None:
            return [self.key, self.value, self.node.to_array()]
        else:
            return [self.key, self.value, None]

    @classmethod
    def from_array(cls, arr: list) -> "Entry[V]":
        """
        Reconstruct an Entry from a compact CBOR array:
          [ key, value, node_array or None ]
        """
        key, value, node_arr = arr
        node = Node.from_array(node_arr) if node_arr is not None else None
        return cls(key=key, value=value, node=node)


@dataclass
class Node(Generic[V]):
    bitmap: int = 0
    entries: list[Entry[V]] = None
    _hash: Optional[bytes] = None  # not serialized

    def __post_init__(self):
        if self.entries is None:
            self.entries = []

    def insert(self, key: bytes, value: V, hs: HashState) -> bool:
        if hs.consumed >= MAX_DEPTH * BITS_PER_STEP:
            return self.insert_fallback(key, value)

        idx = hs.next()
        if idx >= 64:
            raise ValueError(f"idx out of range: {idx}")

        pos = count_ones(self.bitmap, idx)

        if self.bitmap & (1 << idx) == 0:
            self.bitmap |= 1 << idx
            self.entries.insert(pos, Entry(key=key, value=value, node=None))
            self._hash = None
            return True

        entry = self.entries[pos]
        if entry.node is None:
            if entry.key == key:
                entry.value = value
                self._hash = None
                return False

            branch = Node()
            old_hs = HashState.new(entry.key)
            # Skip to the current consumption point
            for _ in range(hs.consumed // BITS_PER_STEP):
                old_hs.next()

            branch.insert(entry.key, entry.value, old_hs)

            new_hs = HashState.new(key)
            # Skip to the current consumption point
            for _ in range(hs.consumed // BITS_PER_STEP):
                new_hs.next()

            branch.insert(key, value, new_hs)

            self.entries[pos] = Entry(key=None, value=None, node=branch)
            self._hash = None
            return True

        inserted = entry.node.insert(key, value, hs)
        if inserted:
            self._hash = None
        return inserted

    def find(self, key: bytes) -> V | None:
        hs = HashState.new(key)
        current_node = self

        while True:
            if current_node is None:
                return None

            if hs.consumed >= MAX_DEPTH * BITS_PER_STEP:
                return current_node.find_fallback(key)

            idx = hs.next()
            if current_node.bitmap & (1 << idx) == 0:
                return None

            pos = count_ones(current_node.bitmap, idx)
            if pos >= len(current_node.entries):
                return None

            entry = current_node.entries[pos]
            if entry.node is None:
                if entry.key == key:
                    return entry.value
                return None

            current_node = entry.node

    def delete(self, key: bytes, hs: HashState) -> bool:
        if hs.consumed >= MAX_DEPTH * BITS_PER_STEP:
            return self.delete_fallback(key)

        idx = hs.next()
        if self.bitmap & (1 << idx) == 0:
            return False

        pos = count_ones(self.bitmap, idx)
        if pos >= len(self.entries):
            raise ValueError(f"pos for idx {idx} out of range: {pos}")

        entry = self.entries[pos]
        if entry.node is None:
            if entry.key != key:
                return False

            self._hash = None
            self.bitmap &= ~(1 << idx)
            self.entries.pop(pos)
            return True

        deleted = entry.node.delete(key, hs)
        if not deleted:
            return False

        if len(entry.node.entries) == 0:
            self.bitmap &= ~(1 << idx)
            self.entries.pop(pos)
        elif len(entry.node.entries) == 1 and entry.node.entries[0].node is None:
            child_entry = entry.node.entries[0]
            self.entries[pos] = Entry(
                key=child_entry.key, value=child_entry.value, node=None
            )

        self._hash = None
        return True

    def insert_fallback(self, key: bytes, value: V) -> bool:
        for i, e in enumerate(self.entries):
            if e.key == key:
                self.entries[i].value = value
                return False
        self.entries.append(Entry(key=key, value=value, node=None))
        self._hash = None
        return True

    def find_fallback(self, key: bytes) -> tuple[Optional[V], bool]:
        for e in self.entries:
            if e.node is None and e.key == key:
                return e.value, True
            elif e.node is not None:
                value, found = e.node.find_fallback(key)
                if found:
                    return value, True
        return None, False

    def delete_fallback(self, key: bytes) -> bool:
        for i, e in enumerate(self.entries):
            if e.node is None and e.key == key:
                self.entries.pop(i)
                self._hash = None
                return True
        return False

    def all(self, fn: Callable[[bytes, V], bool]) -> bool:
        if self is None:
            return True
        for e in self.entries:
            if e.node is None:
                if not fn(e.key, e.value):
                    return False
            else:
                if not e.node.all(fn):
                    return False
        return True

    def hash(self) -> bytes:
        if self._hash is not None:
            return self._hash

        h = hashlib.sha256()
        for e in self.entries:
            # TODO: hash the bitmap instead of the key
            # binary.Write(h, binary.BigEndian, n.Bitmap)
            if e.node is None:
                # TODO: take this out
                h.update(e.key)
                h.update(cbor_encode(e.value))
            else:
                h.update(e.node.hash())

        self._hash = h.digest()
        return self._hash

    def to_array(self) -> list:
        """
        Convert this node into an array for compact CBOR serialization:
          [ bitmap, list_of_entry_arrays ]
        The node's _hash is not serialized.
        """
        return [
            self.bitmap,
            [entry.to_array() for entry in self.entries],
        ]

    @classmethod
    def from_array(cls, arr: list) -> "Node[V]":
        """
        Reconstruct a Node from its array representation:
          [ bitmap, list_of_entry_arrays ]
        """
        if arr is None:
            return None
        bitmap, entries_arr = arr
        if bitmap == 0 and entries_arr is None:
            return None
        node = cls(bitmap=bitmap)
        for e_arr in entries_arr:
            node.entries.append(Entry.from_array(e_arr))
        return node


def count_ones(n: int, below: int) -> int:
    mask = (1 << below) - 1
    return bin(n & mask).count("1")


def deep_copy_node(node: "Node[V]") -> "Node[V]":
    if node is None:
        return None

    new_node = Node(bitmap=node.bitmap)
    new_node.entries = []

    for entry in node.entries:
        new_entry = Entry(
            key=entry.key,
            value=entry.value,
            node=deep_copy_node(entry.node) if entry.node else None,
        )
        new_node.entries.append(new_entry)

    return new_node


KeyType = bytes | int | str


def encode_key(k: KeyType) -> bytes:
    if isinstance(k, int):
        return k.to_bytes(8, "big")
    elif isinstance(k, str):
        return k.encode("utf-8")
    elif isinstance(k, bytes):
        return k
    else:
        raise ValueError(f"Invalid key type: {type(k)}")


@dataclass
class Trie(Generic[V]):
    root: Node[V]
    size: int = 0

    @classmethod
    def new(cls) -> "Trie[V]":
        return cls(root=Node())

    def insert(self, key: KeyType, value: V) -> None:
        if self.root is None:
            self.root = Node()
        key = encode_key(key)
        inserted = self.root.insert(key, value, HashState.new(key))
        if inserted:
            self.size += 1

    def get(self, key: KeyType) -> V | None:
        if self.root is None:
            return None
        key = encode_key(key)
        return self.root.find(key)

    def has(self, key: KeyType) -> bool:
        key = encode_key(key)
        return self.get(key) is not None

    def delete(self, key: KeyType) -> None:
        if self.root is None:
            return
        key = encode_key(key)
        deleted = self.root.delete(key, HashState.new(key))
        if deleted:
            self.size -= 1

    def all(self, fn: Callable[[bytes, V], bool]) -> None:
        if self.root is None:
            return
        self.root.all(fn)

    def hash(self) -> bytes:
        if self.root is None:
            return hashlib.sha256().digest()
        return self.root.hash()

    def copy(self) -> "Trie[V]":
        """Create a deep copy of the trie"""
        new_trie = Trie.new()
        new_trie.root = deep_copy_node(self.root)
        new_trie.size = self.size
        return new_trie

    def to_cbor_array(self) -> list:
        """
        Marshal the Trie into a CBOR-encoded byte buffer. Follows the Go approach:
        only the root node is serialized, and size is recomputed when unmarshaling.
        """
        # Convert the root node into its array representation and dump to CBOR
        return self.root.to_array()

    @classmethod
    def from_cbor_array(cls, data: list) -> "Trie[V]":
        """
        Unmarshal a CBOR-encoded byte buffer into a new Trie. Recomputes size
        by counting entries recursively.
        """
        root = Node.from_array(data)
        t = cls(root=root)
        t._recalculate_size()
        return t

    def _recalculate_size(self) -> None:
        """
        Recompute the size by counting leaf entries under the root node.
        """

        def count_entries(n: Node[V]) -> int:
            if not n:
                return 0
            c = 0
            for e in n.entries:
                if e.node is None:
                    c += 1
                else:
                    c += count_entries(e.node)
            return c

        self.size = count_entries(self.root)
