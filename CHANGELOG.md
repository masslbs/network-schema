<!--
SPDX-FileCopyrightText: 2024 Mass Labs

SPDX-License-Identifier: MIT
-->

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