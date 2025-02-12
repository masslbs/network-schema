# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

.phony: all lint reuse go-pb-code

all: vectors/hamt_test.json vectors/ShopOkay.cbor python-package go-pb-code reuse

python-package: vectors/ShopOkay.cbor vectors/hamt_test.json
	make -C python

go-pb-code: *.proto
	./generate_protoc_go.bash

vectors/hamt_test.json:
	cd python && $(PYTHON) ./generate_hamt_test_vectors.py

vectors/ShopOkay.cbor:
	cd go/cbor && TEST_DATA_OUT=../../vectors go test

lint:
	$(PYTHON) ./check.py
	buf format -w
	git ls-files | grep -E '.(py|pyi)$$' | xargs black
	protolint lint *.proto
	reuse lint

LIC := MIT
CPY := "Mass Labs"

reuse:
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights Makefile CHANGELOG.md README.md *.proto flake.nix .gitignore .github/workflows/test.yml *.py python/*.py python/tests/*.py python/massmarket_hash_event/*.pyi python/massmarket_hash_event/*.py python/pyproject.toml python/Makefile go.mod go.sum go/cbor/*.go go/pb/*.go
	reuse annotate --license  $(LIC) --copyright $(CPY) --merge-copyrights --force-dot-license VERSION *.txt  flake.lock

clean:
	rm -r vectors
	mkdir vectors