# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

__all__ = [
    "verify_proof",
    "get_root_hash_of_patches",
    "get_signer_of_patchset",
    # protobuf
    "transport_pb2",
    "authentication_pb2",
    "shop_requests_pb2",
    "error_pb2",
    "storage_pb2",
]

from typing import List

from hashlib import sha256
from web3 import Web3

w3 = Web3()

from eth_account.messages import encode_defunct

from massmarket.mmr.db import FlatDB
from massmarket.mmr.algorithms import add_leaf_hash

from massmarket.cbor_encoder import cbor_encode
from massmarket.cbor.patch import Patch
from massmarket.cbor.patch import SignedPatchSet


def get_root_hash_of_patches(patches: List[Patch]):
    patches_bytes = []
    for i, patch in enumerate(patches):
        patch_data = patch
        if hasattr(patch, "to_cbor_dict"):
            patch_data = patch.to_cbor_dict()
        encoded_patch = cbor_encode(patch_data)
        # print(f"DEBUG patch {i}: {encoded_patch.hex()}")
        patches_bytes.append(encoded_patch)
    hashed_patches = [sha256(patch).digest() for patch in patches_bytes]

    # Add zeros until we have a power of 2 length
    next_power_2 = 1
    while next_power_2 < len(hashed_patches):
        next_power_2 *= 2

    # Pad with zeros to reach power of 2 length
    zero_hash = sha256(b"").digest()
    while len(hashed_patches) < next_power_2:
        hashed_patches.append(zero_hash)
    # print(f"Hashed Patches (filled): {len(hashed_patches)}")
    tree = FlatDB()
    positions = [add_leaf_hash(tree, patch) for patch in hashed_patches]
    # print(f"Positions: {positions}")
    calculated_root = tree.get(positions[-1] - 1)
    return calculated_root


# from pprint import pprint


def get_signer_of_patchset(ps: SignedPatchSet):
    # print(f"Patch Count: {len(patches)}")

    calculated_root = get_root_hash_of_patches(ps.patches)
    # pprint(calculated_root.hex())
    # pprint(ps.header.root_hash.hex())
    assert calculated_root == ps.header.root_hash

    header_bytes = cbor_encode(ps.header.to_cbor_dict())
    encoded_header = encode_defunct(header_bytes)
    pub_key = w3.eth.account.recover_message(encoded_header, signature=ps.signature)
    return pub_key


from massmarket.mmr.algorithms import included_root, verify_inclusion_path


class RootMismatchError(Exception):
    def __init__(self, calculated_root: bytes, expected_root: bytes):
        self.calculated_root = calculated_root
        self.expected_root = expected_root
        super().__init__(
            f"Root mismatch: calculated {calculated_root.hex()} but expected {expected_root.hex()}"
        )


class VerificationError(Exception):
    def __init__(self):
        super().__init__("Failed to verify inclusion path")


class PathLengthError(Exception):
    def __init__(self, consumed: int, total: int):
        self.consumed = consumed
        self.total = total
        super().__init__(
            f"Path length mismatch: consumed {consumed} elements but path has {total}"
        )


def verify_proof(
    leaf_index: int, element: bytes, path: list[bytes], wanted_root: bytes
) -> None:
    root = included_root(leaf_index, element, path)
    if root != wanted_root:
        raise RootMismatchError(root, wanted_root)

    (ok, pathconsumed) = verify_inclusion_path(leaf_index, element, path, wanted_root)
    if not ok:
        raise VerificationError()
    if pathconsumed != len(path):
        raise PathLengthError(pathconsumed, len(path))
