# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import xxhash
from dataclasses import dataclass
from typing import TypeVar, Generic, Optional, Any, Callable
import cbor2

BITS_PER_STEP = 5
MAX_DEPTH = (64 + BITS_PER_STEP - 1) // BITS_PER_STEP

V = TypeVar("V")


@dataclass
class HashState:
    original_key: bytes
    hash: int
    consumed: int
    seed: int

    @classmethod
    def new(cls, key: bytes) -> "HashState":
        return cls(
            original_key=key, hash=hash_key_with_seed(key, 0), consumed=0, seed=0
        )

    def next(self) -> int:
        if self.consumed + BITS_PER_STEP > MAX_DEPTH * BITS_PER_STEP:
            self.seed += 1
            self.hash = hash_key_with_seed(self.original_key, self.seed)
            self.consumed = 0

        shift = self.consumed
        mask = (1 << BITS_PER_STEP) - 1
        chunk = (self.hash >> shift) & mask
        self.consumed += BITS_PER_STEP
        return chunk


@dataclass
class Entry(Generic[V]):
    key: Optional[bytes]
    value: Optional[V]
    node: Optional["Node[V]"]


@dataclass
class Node(Generic[V]):
    bitmap: int = 0
    entries: list[Entry[V]] = None
    _hash: Optional[bytes] = None

    def __post_init__(self):
        if self.entries is None:
            self.entries = []

    def insert(self, key: bytes, value: V, hs: HashState) -> bool:
        if hs.consumed >= MAX_DEPTH * BITS_PER_STEP:
            return self.insert_fallback(key, value)

        idx = hs.next()
        if idx >= 32:
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
            old_hs = HashState(
                original_key=entry.key,
                hash=hash_key(entry.key),
                consumed=hs.consumed,
                seed=hs.seed,
            )
            branch.insert(entry.key, entry.value, old_hs)

            new_hs = HashState(
                original_key=key, hash=hash_key(key), consumed=hs.consumed, seed=hs.seed
            )
            branch.insert(key, value, new_hs)

            self.entries[pos] = Entry(key=None, value=None, node=branch)
            self._hash = None
            return True

        inserted = entry.node.insert(key, value, hs)
        if inserted:
            self._hash = None
        return inserted

    def find(self, key: bytes) -> tuple[Optional[V], bool]:
        hs = HashState.new(key)
        current_node = self

        while True:
            if current_node is None:
                return None, False

            if hs.consumed >= MAX_DEPTH * BITS_PER_STEP:
                return current_node.find_fallback(key)

            idx = hs.next()
            if current_node.bitmap & (1 << idx) == 0:
                return None, False

            pos = count_ones(current_node.bitmap, idx)
            if pos >= len(current_node.entries):
                return None, False

            entry = current_node.entries[pos]
            if entry.node is None:
                if entry.key == key:
                    return entry.value, True
                return None, False

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

        h = xxhash.xxh64()
        for e in self.entries:
            if e.node is None:
                h.update(e.key)
                h.update(cbor2.dumps(e.value))
            else:
                h.update(e.node.hash())

        self._hash = h.digest()
        return self._hash


def hash_key(key: bytes) -> int:
    return xxhash.xxh64(key).intdigest()


def hash_key_with_seed(key: bytes, seed: int) -> int:
    return xxhash.xxh64(key, seed=seed).intdigest()


def count_ones(n: int, below: int) -> int:
    mask = (1 << below) - 1
    return bin(n & mask).count("1")


def deep_copy_node(node: Node[V]) -> Node[V]:
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


@dataclass
class Trie(Generic[V]):
    root: Node[V]
    size: int = 0

    @classmethod
    def new(cls) -> "Trie[V]":
        return cls(root=Node())

    def insert(self, key: bytes, value: V) -> None:
        if self.root is None:
            self.root = Node()

        inserted = self.root.insert(key, value, HashState.new(key))
        if inserted:
            self.size += 1

    def get(self, key: bytes) -> tuple[Optional[V], bool]:
        if self.root is None:
            return None, False
        return self.root.find(key)

    def delete(self, key: bytes) -> None:
        if self.root is None:
            return
        deleted = self.root.delete(key, HashState.new(key))
        if deleted:
            self.size -= 1

    def all(self, fn: Callable[[bytes, V], bool]) -> None:
        if self.root is None:
            return
        self.root.all(fn)

    def hash(self) -> bytes:
        if self.root is None:
            return xxhash.xxh64().digest()
        return self.root.hash()

    def copy(self) -> "Trie[V]":
        """Create a deep copy of the trie"""
        new_trie = Trie.new()
        new_trie.root = deep_copy_node(self.root)
        new_trie.size = self.size
        return new_trie
