import { assertEquals } from "jsr:@std/assert";
import { testDecode, testEncode } from "./main.ts";
import sha3 from "npm:js-sha3";

Deno.test(function reencodeVectorsTest() {
  const fname = "../go/vectors_patch_shop.cbor";
  const data = Deno.readFileSync(fname);
  const decoded = testDecode(data);

  console.log("first level keys: ", decoded.keys());



  decoded.get("Patches").forEach((patch: any, idx: number) => {
    const after = patch.get("After");
    const reencoded = testEncode(after);
    assertEquals(reencoded.length, patch.get("Encoded").length);
    const hashed = sha3.keccak256(reencoded);

    const want = patch.get("Hash");
    const wantHex = bytesToHex(want);
    assertEquals(hashed, wantHex);
  })
});

function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('')
}
