# SPDX-FileCopyrightText: 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import cbor2
from sha3 import keccak_256
import pytest

# TODO: move to dedicated package instead of vendoring the code
from massmarket_hash_event.mmr.algorithms import add_leaf_hash
from massmarket_hash_event.mmr.db import FlatDB

from massmarket_hash_event import verify_proof, RootMismatchError


def hex(x: bytes) -> str:
    return "0x" + x.hex()


def test_merkle_proofs():
    vector_file = "../vectors/MerkleProofs.cbor"
    """Verify merkle proofs from test vectors."""
    with open(vector_file, "rb") as f:
        vectors = cbor2.load(f)

    for test_case in vectors:
        test_name = test_case["Name"]
        underline = "=" * len(test_name)
        print(f"\n{underline}\n{test_name}\n{underline}\n")

        # Convert patches to bytes for hashing
        patches_bytes = [cbor2.dumps(patch) for patch in test_case["Patches"]]
        hashed_patches = [keccak_256(patch).digest() for patch in patches_bytes]

        # Create merkle tree
        tree = FlatDB()
        positions = [add_leaf_hash(tree, patch) for patch in hashed_patches]
        print(f"leaf positions: {positions}")

        wantRoot = test_case["RootHash"]
        print(f"wanted root: {hex(wantRoot)}")

        # Verify each proof
        proofs = test_case["Proofs"]
        assert proofs is not None

        print()
        for proof_index, proof in enumerate(proofs):
            (leaf_index, size, path) = proof

            # this is a special case for proofing a single item.
            # TODO: format vectors differently
            if path is None:
                path = []

            verify_proof(leaf_index, tree.get(leaf_index), path, wantRoot)


def test_verify_proof_errors():
    # Create a simple tree with one element
    tree = FlatDB()
    element = keccak_256(b"test").digest()
    leaf_index = add_leaf_hash(tree, element) - 1
    root = tree.get(leaf_index)
    path = []

    # Test with wrong root hash
    wrong_root = keccak_256(b"wrong").digest()
    with pytest.raises(RootMismatchError) as exc_info:
        verify_proof(leaf_index, element, path, wrong_root)
    assert exc_info.value.calculated_root == root
    assert exc_info.value.expected_root == wrong_root

    # Test with wrong element
    wrong_element = keccak_256(b"wrong element").digest()
    with pytest.raises(RootMismatchError):
        verify_proof(leaf_index, wrong_element, path, root)
    assert exc_info.value.calculated_root == root
    assert exc_info.value.expected_root == wrong_root

    # TODO: add path length error


if __name__ == "__main__":
    test_merkle_proofs()
