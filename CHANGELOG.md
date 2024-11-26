<!--
SPDX-FileCopyrightText: 2024 - 2025 Mass Labs

SPDX-License-Identifier: MIT
-->

# V4 (2024-12-12) DRAFT

- Implement shop event types in CBOR
  - This enables canonical encoding of events
  - Which also paves the way for merkle roots of a shop
  - See `cddl/shop_logical.cddl`
- Required checks via go-playground/validator
- patch vectors for testing updates of a shop
- Introduces Merkle Mountain Range (MMR) for proofs of inclusion of events
  - See https://datatracker.ietf.org/doc/html/draft-bryce-cose-merkle-mountain-range-proofs
- Introduces a Hashed Array Mapped Trie (HAMT) for shop object storage
- Introduces a new patch format for shop updates
   - See `cddl/patch.cddl`

# V3 (2024-11-12) DevCon '24 relase, shipping regions and currency management

- Deprecates V2
- Remove need for manual encoding.txt / introduces envelope.proto
- Introduces event subscriptions, no implicit push of events
- Authentication is only necessary for clients that want to write events.
- Introduces base_types.proto to differentiate various `bytes`
- Rename _item_ to _listing_
- Tags are only used for categories, introduce seperate view states for listings
- Add variations to listings and orders
- Added order modifiers for shipping and taxes
- Remove shop manifest metadata in favour of NFT metadata
- Replace all price handling with Uint256 values

# V2 (2024-07-17) EthCC '24 relase, guest testing checkout

- Deprecates V1
- Seperate out event types from the transport.
- Introduce a semantic layering of transport > auth > shop
- Rename _store_ to _shop_
- Rename _cart_ to _order_
- Changes all Tag and Order related messages to follow the Update pattern we are using in Manifest and Item.
- Replace EIP-712 signatures with EIP-191, since we don't verify events on-chain
- Define error codes
- Add multi-chain currency managment (base and accepted currencies, payee)

# V1.1 (2024-06-04) Payments V2 integration

- Our new payments smart contracts allows the client to choose if they want to use counter-factual addresses to pay or not
- To test this out clients need to re-construct a `PaymentRequest`
  - Therefore we are adding `payment_ttl` and to the `CartFinalized` event
  - The `shopSignature` is not checked yet. Use 64 bytes zeros
  - `payment_id` can be compared with the relay by calling `getPaymentId` on the contract
- We use the owner of the store NFT for the `payeeAddress` by default
  - The Clerk can chose a different escrow contract during `CommitCartRequest` if necessary
  - How this is done will change in protocol version 2

# V1 (2024-04-24) EthDuba '24 relase, point-of-sales

- _Users_ that are registered with the _Store_ (see `StoreReg` smart contract) can POST to `https://my.relay/v1/enroll_keycard`.
  - The _Relay_ creates a `NewKeyCard` _Event_ to inform other _Users_ about them.
- _Users_ that have enrolled a KeyCard connect to `wss://my.relay/v1/sessions`.
  - They can use `AuthenticateRequest` to begin authentication.
  - The server responds with a `challenge` in the response that needs to be signed over / solved.
  - The `ChallengeSolvedRequest` is used to send that signature back.
- The server (_Relay_) sends `PingRequests` in fixed intervals which need to be responded to to keep the connection alive.
- `SyncStatusRequest` is sent by a _Relay_ to inform a client about how many _Events_ are left to sync.
- _Event_ types to facilitate listing and inventory managment are:
  - `CreateItem`, `UpdateItem`, `Create/AddTo/RemoveFrom/DeleteTag` and `ChangeStock`.
- _Event_ types for shopping are `CreateCart`, `ChangeCart`, `CartFinalized` and `CartAbandoned`.
  - To finish a purchase, the `CartFinalized` event has the needed information for the transaction.
  - The _Relay_ starts watching for deposists to the `purchase_address`.
  - Once enough money has been transfered, the _Relay_ creates a `ChangeStock` _Event_ which includes a `cart_id` reference.
  - This signals the _User_ that the purchase is complete.
- Ability to write and receive _Events_ to/from a _Relay_ (`EventWriteReq/Resp`, `EventPushReq/Resp`).
  - `EventPushRequest` needs to be responded to without an error to receive more events.
  - This builds the back preassure mechanism of this protocol.
- `GetBlobUploadURL` returns an URL that files can be uploaded to.
  - Once the upload is complete, the HTTP response will contain a reference for the uploaded file.
- See our documentation for further details of the architecture and how it fits together with the smart contracts.
