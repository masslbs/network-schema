# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import os
import cbor2
from pprint import pprint
from web3 import Web3

from massmarket import (
    get_signer_of_patchset,
)
from massmarket.cbor.patch import SignedPatchSet


# check that we can recompute the signatures from the test vectors
def test_verify_vector_file():
    files = [f for f in os.listdir("../vectors") if f.endswith("Okay.cbor")]
    assert len(files) > 0, "no test vectors found"
    for file in files:
        if file in ["InventoryOkay.cbor", "ShopOkay.cbor"]:
            continue  # TODO: canonical fix needed
        with open(f"../vectors/{file}", "rb") as f:
            print(file)
            vector = cbor2.load(f)
            check_vector(vector)


def check_vector(vector):
    assert len(vector["Snapshots"]) > 0
    assert "Signer" in vector
    signer = vector["Signer"]["Address"]["Address"]
    # convert the signer address to a hex string
    signer = Web3.to_checksum_address(signer)

    # pprint(vector)
    patch_set = vector["PatchSet"]

    # 0. reencode the patch set and verify that it round trips
    cbor_bytes = cbor2.dumps(patch_set)
    reencoded_patch_set = SignedPatchSet.from_cbor(cbor_bytes)
    assert reencoded_patch_set.header.root_hash == patch_set["Header"]["RootHash"]

    # 1. recompute root from patches and extract signer pubkey
    extracted_signer = get_signer_of_patchset(reencoded_patch_set)

    # 2. verify the signature
    their_addr = Web3.to_checksum_address(extracted_signer)
    assert their_addr == signer, f"invalid signer on event"


if __name__ == "__main__":
    test_verify_vector_file()
