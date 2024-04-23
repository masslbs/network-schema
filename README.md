<!--
SPDX-FileCopyrightText: 2024 Mass Labs

SPDX-License-Identifier: MIT
-->

# masslbs networks schema

This repository specifies the structure and content of the communication between _Relays_ and their _Clients_. It assumes familiarity with the general architecture of the mass market system and the [smart contracts](https://github.com/masslbs/contracts) it uses.

On an abstract level, the _Relays_ build an [Append-Only Log](https://en.wikipedia.org/wiki/Append-only) per registered _Store_. These logs are accessible via a Request/Response scheme. The _Clients_ cryptographically sign _Events_ and write them to the _Relay_. A _Relay_ keeps track of which _Events_ were send and received from which _Client_ and thus is able to push _Events_ to _Clients_ that haven't written them such that all _Clients_ can build the same state of a _Store_ eventually.

For a detailed description of each message see `schema.proto` and the `CHANGELOG.md`.

For a detailed description of how Events are signed as well as the HTTP Reqeusts acompanying the WebSocket connection, see our [documentation page](https://docs.mass.market).

This repo also contains a `python` folder with the code for the [massmarket-hash-event](https://pypi.org/project/massmarket-hash-event/#description) pip package, used in our test suite.

## tooling

Protobuf

* `protoc`
* [protolint](https://github.com/yoheimuta/protolint)

Python Package

* pyproject
* web3.py for eth_typedData v4
* pytest

License Maintenance

* [reuse](https://github.com/fsfe/reuse-tool#install)

### Updating the python package

```bash
nix develop
# to update schema_pb2.(py|pyi)
make
cd python
# make sure the tests pass first
pytest
# -n to switch off venv. nix already gives us that
$PYTHON -m build -n
# see bitwarden for login info
$PYTHON -m twine upload dist/*
```

## LICENSE

MIT
