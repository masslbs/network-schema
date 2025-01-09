# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

import os
import cbor2


from web3 import Account, Web3

from massmarket_hash_event import (
    hash_patchset,
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

    patch_set = vector["PatchSet"]

    # re-encode the patchset
    encoded_data = hash_patchset(patch_set)

    # verify the signature
    pub_key = Account.recover_message(encoded_data, signature=vector["Signature"])
    their_addr = Web3.to_checksum_address(pub_key)
    assert their_addr == signer, f"invalid signer on event"


if __name__ == "__main__":
    test_verify_vector_file()
