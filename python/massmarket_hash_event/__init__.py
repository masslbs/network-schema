# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

__all__ = [
    "get_signer_of_patchset",
    "transport_pb2",
    "authentication_pb2",
    "shop_requests_pb2",
    "error_pb2",
    "storage_pb2",
]

import cbor2
from sha3 import keccak_256
from web3 import Web3
w3 = Web3()

from eth_account.messages import encode_defunct

from massmarket_hash_event.mmr.db import FlatDB
from massmarket_hash_event.mmr.algorithms import add_leaf_hash

def get_signer_of_patchset(patchSet):
    assert "Patches" in patchSet
    assert "Header" in patchSet
    header = patchSet["Header"]
    assert "RootHash" in header
    want_root = header["RootHash"]
    patches = patchSet["Patches"]
    # print(f"Patch Count: {len(patches)}")

    patches_bytes = [cbor2.dumps(patch) for patch in patches]
    hashed_patches = [keccak_256(patch).digest() for patch in patches_bytes]

    # Add zeros until we have a power of 2 length
    next_power_2 = 1
    while next_power_2 < len(hashed_patches):
        next_power_2 *= 2
    
    # Pad with zeros to reach power of 2 length
    zero_hash = keccak_256(b'').digest()
    while len(hashed_patches) < next_power_2:
        hashed_patches.append(zero_hash)    
    # print(f"Hashed Patches (filled): {len(hashed_patches)}")
    tree = FlatDB()
    positions = [add_leaf_hash(tree, patch) for patch in hashed_patches]
    # print(f"Positions: {positions}")
    calculated_root = tree.get(positions[-1]-1)
    assert calculated_root == want_root

    header_bytes = cbor2.dumps(header)
    encoded_header = encode_defunct(header_bytes)
    pub_key = w3.eth.account.recover_message(encoded_header, signature=patchSet["Signature"])
    return pub_key

    