// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

module github.com/masslbs/network-schema/v5

go 1.24

toolchain go1.24.4

require (
	github.com/datatrails/go-datatrails-merklelog/mmr v0.4.1
	github.com/ethereum/go-ethereum v1.16.1
	github.com/fxamacker/cbor/v2 v2.8.0
	github.com/go-playground/validator/v10 v10.27.0
	github.com/google/go-cmp v0.5.9
	github.com/huandu/go-clone/generic v1.7.2
	github.com/jackc/pgx/v5 v5.7.5
	github.com/peterldowns/testy v0.0.3
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/bits-and-blooms/bitset v1.20.0 // indirect
	github.com/consensys/gnark-crypto v0.18.0 // indirect
	github.com/crate-crypto/go-eth-kzg v1.3.0 // indirect
	github.com/crate-crypto/go-ipa v0.0.0-20240724233137-53bbb0ceb27a // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/ethereum/c-kzg-4844/v2 v2.1.0 // indirect
	github.com/ethereum/go-verkle v0.2.2 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/huandu/go-clone v1.7.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/supranational/blst v0.3.14 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/exp v0.0.0-20231110203233-9a3e6036ecaa // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
)

// replace github.com/masslbs/go-pgmmr => ../tmp/go-pgmmr
