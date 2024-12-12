// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

import { Decoder, Encoder } from "cbor-x";
// import sha3 from "npm:js-sha3";

export function testDecode(data: Uint8Array): any {
  return new Decoder({
    mapsAsObjects: false,
  }).decode(data);
}

export function testEncode(data: any): Uint8Array {
  return new Encoder({
    sortMaps: true,
    useRecords: false,
    variableMapSize: true,
    useTag259ForMaps: false,
 }).encode(data);
}

// Learn more at https://deno.land/manual/examples/module_metadata#concepts
if (import.meta.main) {
 
}
