import cbor2
from sha3 import keccak_256

# TODO: move to dedicated package instead of vendoring the code
from massmarket_hash_event.mmr.algorithms import included_root, verify_inclusion_path, add_leaf_hash
from massmarket_hash_event.mmr.db import FlatDB

def hex(x: bytes) -> str:
    return "0x" + x.hex()

def test_merkle_proofs():
    vector_file = "../vectors/MerkleProofs.cbor"
    """Verify merkle proofs from test vectors."""
    with open(vector_file, 'rb') as f:
        vectors = cbor2.load(f)

    for test_case in vectors:
        test_name = test_case['Name']
        underline = "="*len(test_name)
        print(f"\n{underline}\n{test_name}\n{underline}\n")
        
        # Convert patches to bytes for hashing
        patches_bytes = [cbor2.dumps(patch) for patch in test_case['Patches']]
        hashed_patches = [keccak_256(patch).digest() for patch in patches_bytes]

        # Create merkle tree
        tree = FlatDB()
        positions = [add_leaf_hash(tree, patch) for patch in hashed_patches]
        print(f"leaf positions: {positions}")

        wantRoot = test_case['RootHash']
        print(f"wanted root: {hex(wantRoot)}")
      
        # Verify each proof
        proofs = test_case['Proofs']
        assert proofs is not None

        print()
        for proof_index, proof in enumerate(proofs):
            (leaf_index, size, path) = proof

            # this is a special case for proofing a single item.
            # TODO: format vectors differently
            if path is None:
                path = []
            # print(f"proof {proof_index}:\n\tleaf_index: {leaf_index}\n\tsize: {size}\n\tpath: {[hex(p) for p in path]}")

            root = included_root(leaf_index, tree.get(leaf_index), path)

            assert root == wantRoot
            
            (ok, pathconsumed) = verify_inclusion_path(leaf_index, tree.get(leaf_index), path, wantRoot)
            assert ok
            # print(f"proof {proof_index} ok. consumed {pathconsumed} of {len(path)}")
            assert pathconsumed == len(path)


         
if __name__ == "__main__":
    test_merkle_proofs()
