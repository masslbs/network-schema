# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import os
import cbor2
from pprint import pprint
from web3 import Account, Web3

from massmarket_hash_event import (
    get_signer_of_patchset,
)

# check that we can recompute the signatures from the test vectors
def test_verify_vector_file():
    files = [f for f in os.listdir("../vectors") if f.endswith("Okay.cbor")]
    assert len(files) > 0, "no test vectors found"
    for file in files:
        with open(f"../vectors/{file}", "rb") as f:
            print(file)
            vector = cbor2.load(f)
            check_vector(vector)


def check_vector(vector):
    assert len(vector["Snapshots"]) > 0
    assert "Signer" in vector
    signer = vector["Signer"]["Address"]
    # convert the signer address to a hex string
    signer = Web3.to_checksum_address(signer)

    # pprint(vector)
    patch_set = vector["PatchSet"]

    # 1. recompute root from patches and extract signer pubkey
    extracted_signer = get_signer_of_patchset(patch_set)

    # verify the signature
    their_addr = Web3.to_checksum_address(extracted_signer)
    assert their_addr == signer, f"invalid signer on event"

if __name__ == "__main__":
    test_verify_vector_file()
