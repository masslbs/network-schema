# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

.phony: all lint reuse go-pb-code

all: vectors/hamt_test.cbor vectors/ShopOkay.cbor python-package go-pb-code go-tests reuse

python-package: vectors/ShopOkay.cbor vectors/hamt_test.cbor
	make -C python

go-pb-code: *.proto
	./generate_protoc_go.bash

vectors/hamt_test.cbor:
	cd python && $(PYTHON) ./generate_hamt_test_vectors.py

vectors/ShopOkay.cbor:
	cd go/patch && TEST_DATA_OUT=../../vectors go test

go-tests: vectors/hamt_test.cbor
	cd go && go test ./...

lint:
	$(PYTHON) ./check.py
	buf format -w
	git ls-files | grep -E '.(py|pyi)$$' | xargs black
	protolint lint *.proto
	reuse lint

LIC := MIT
CPY := "Mass Labs"

reuse:
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights Makefile CHANGELOG.md README.md *.proto flake.nix .gitignore .github/workflows/test.yml *.py python/*.py python/tests/*.py python/massmarket/*.pyi python/massmarket/*.py python/massmarket/cbor/*.py python/pyproject.toml python/Makefile go.mod go.sum go/cbor/*.go go/pb/*.go go/objects/*.go go/patch/*.go go/hamt/*.go go/mmr/*.go go/internal/testhelper/*.go
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights --force-dot-license VERSION *.txt  flake.lock
	reuse annotate --license BSD-2-Clause --copyright "IETF / draft-bryce-cose-merkle-mountain-range-proofs-02" python/massmarket/mmr/*

clean:
	rm -r vectors
	mkdir vectors
