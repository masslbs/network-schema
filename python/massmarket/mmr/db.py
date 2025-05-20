# SPDX-FileCopyrightText: 2025 IETF / draft-bryce-cose-merkle-mountain-range-proofs-02
#
# SPDX-License-Identifier: BSD-2-Clause

from hashlib import sha256

from massmarket.mmr.algorithms import add_leaf_hash
from massmarket.mmr.algorithms import leaf_count
from massmarket.mmr.algorithms import complete_mmr
from massmarket.mmr.algorithms import hash_pospair64
from massmarket.mmr.algorithms import trailing_zeros


def hash_num64(v: int) -> bytes:
    """
    Compute the SHA-256 hash of v

    Args:
        v (int): assumed to be an unsigned integer using at most 64 bits
    Returns:
        bytes: the SHA-256 hash of the big endian representation of v
    """
    return sha256(v.to_bytes(8, byteorder="big", signed=False)).digest()


class FlatDB:
    """An implementation that satisfies the required interafce of addleafhash"""

    def __init__(self):
        self.store = []

    def append(self, v):
        self.store.append(v)
        return len(self.store)  # index of the *NEXT* item that will be added

    def get(self, i):
        return self.store[i]

    def init_canonical39(self):
        """Re-creates the kat db using addleafhash"""

        for ileaf in range(leaf_count(38)):
            # we know its a leaf, and we know len(self.store) is a valid mmr size,
            # so there is a short cut here for leaf index -> mmr index.
            # the count of trailing zeros in the leaf index is also the number of nodes we will need to add
            i = len(self.store)
            # and even numbered leaves are always singleton peaks
            if ileaf % 2:
                i = i + trailing_zeros(ileaf)

            add_leaf_hash(self, hash_num64(i))

    def init_size(self, mmrsize: int):
        """Re-creates the kat db using addleafhash"""

        mmr = complete_mmr(mmrsize - 1)

        for ileaf in range(leaf_count(mmr)):
            # we know its a leaf, and we know len(self.store) is a valid mmr size,
            # so there is a short cut here for leaf index -> mmr index.
            # the count of trailing zeros in the leaf index is also the number of nodes we will need to add
            i = len(self.store)
            # and even numbered leaves are always singleton peaks
            if ileaf % 2:
                i = i + trailing_zeros(ileaf)

            add_leaf_hash(self, hash_num64(i))


class KatDB:
    """A fixed size database for providing "known answers" """

    def __init__(self):
        # A map is used so we can build the tree in layers, with explicit put()
        # calls, for illustrative purposes In a more typical implementation,
        # this would just be a list.
        self.store = {}

    def parent_hash(self, iparent: int, ileft: int, iright: int) -> bytes:
        vleft = self.store[ileft]
        vright = self.store[iright]
        return hash_pospair64(iparent + 1, vleft, vright)

    def put(self, i: int, v: bytes):
        self.store[i] = v

    def get(self, i) -> bytes:
        return self.store[i]

    def init_canonical39(self):

        self.store = {}

        # height 0 (the leaves)
        self.put(0, hash_num64(0))
        self.put(1, hash_num64(1))
        self.put(3, hash_num64(3))
        self.put(4, hash_num64(4))
        self.put(7, hash_num64(7))
        self.put(8, hash_num64(8))
        self.put(10, hash_num64(10))
        self.put(11, hash_num64(11))
        self.put(15, hash_num64(15))
        self.put(16, hash_num64(16))
        self.put(18, hash_num64(18))
        self.put(19, hash_num64(19))
        self.put(22, hash_num64(22))
        self.put(23, hash_num64(23))
        self.put(25, hash_num64(25))
        self.put(26, hash_num64(26))
        self.put(31, hash_num64(31))
        self.put(32, hash_num64(32))
        self.put(34, hash_num64(34))
        self.put(35, hash_num64(35))
        self.put(38, hash_num64(38))

        # height 1
        self.put(2, self.parent_hash(2, 0, 1))
        self.put(5, self.parent_hash(5, 3, 4))
        self.put(9, self.parent_hash(9, 7, 8))
        self.put(12, self.parent_hash(12, 10, 11))
        self.put(17, self.parent_hash(17, 15, 16))
        self.put(20, self.parent_hash(20, 18, 19))
        self.put(24, self.parent_hash(24, 22, 23))
        self.put(27, self.parent_hash(27, 25, 26))
        self.put(33, self.parent_hash(33, 31, 32))
        self.put(36, self.parent_hash(36, 34, 35))

        # height 2
        self.put(6, self.parent_hash(6, 2, 5))
        self.put(13, self.parent_hash(13, 9, 12))
        self.put(21, self.parent_hash(21, 17, 20))
        self.put(28, self.parent_hash(28, 24, 27))
        self.put(37, self.parent_hash(37, 33, 36))

        # height 3
        self.put(14, self.parent_hash(14, 6, 13))
        self.put(29, self.parent_hash(29, 21, 28))

        # height 4
        self.put(30, self.parent_hash(30, 14, 29))
