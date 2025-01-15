import { StandardMerkleTree } from "@openzeppelin/merkle-tree"
import { decode, encode } from "npm:cbor-x";
// import { Keccak } from "npm:sha3";

// Read and decode the CBOR test vectors file
const vectorBytes = await Deno.readFile("../vectors/MerkleProofs.cbor");
const vectors = decode(vectorBytes);


for (const vector of vectors) {
  console.log("\n==========\n" + vector.Name);

  const wantRoot = Buffer.from(vector.RootHash).toString("hex");
  
  // convert patches back to cbor data
  // const patches = vector.Patches.map((patch: Buffer) => ["0x" + encode(patch).toString("hex")]);
  const patches = vector.Patches.map((patch: Buffer) => [encode(patch)]);
  // console.log(patches) 

  console.log(`patch[0]: ${patches[0]}`)

  const tree = StandardMerkleTree.of(patches, ["bytes"], {sortLeaves: true});
  const gotRoot = tree.root
  console.log(`Got:\t${gotRoot}`);


  if (gotRoot !== wantRoot) {
    console.log(`Want:\t0x${wantRoot}`);
    console.log(`\n!!!mismatch!!!`);

    console.log(tree.render())
    // Deno.exit(1)
  }




  const proof = tree.getProof(patches[0]);
  console.log(`proof:`)
  console.log(proof)
}
