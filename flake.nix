# SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
#
# SPDX-License-Identifier: MIT

{
  description = "Mass Market Network Schema";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      # url = "github:tweag/gomod2nix";
      # fix builds with go 1.23
      # https://github.com/nix-community/gomod2nix/pull/168
      url = "github:obreitwi/gomod2nix/fix/go_mod_vendor";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    bryce-mmrs-py = {
      url = "github:robinbryce/draft-bryce-cose-merkle-mountain-range-proofs";
      flake = false;
    };
  };

  outputs = {
    nixpkgs,
    utils,
    gomod2nix,
    bryce-mmrs-py,
    ...
  }:
    utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ gomod2nix.overlays.default ];
      };

      ourPython = pkgs.python311;

      pinnedPython = ourPython.override {
        self = ourPython;
      };

      mass-python = pinnedPython.withPackages (ps:
        with ps; [
          protobuf
          # web3 previous vectors
          safe-pysha3

          matplotlib # go/hamt_bench_plot.py
          cbor2
          xxhash

          # packaging massmarket_hash_event
          pytest
          setuptools
          setuptools-scm
          wheel
          build
          twine

        ]
      );

      buildInputs = with pkgs; [
        go_1_23
        go-outline
        gopls
        gopkgs
        go-tools
        delve
        revive
        errcheck
        unconvert
        godef
        clang
        cbor-diag
        deno
        buf
        black
        reuse
        protobuf
        protolint
        pyright
        mass-python
        gomod2nix.packages.${system}.default
      ];
    in {
      devShell = pkgs.mkShell {
        inherit buildInputs;
        shellHook = ''
          test -d go || {
            echo "go/ directory not found, skipping gomod2nix update"
            exit 0
          }
          gomod2nix generate
          # TODO: this should be an actual python package
          # but i dont want to package it up right now
          test -d python/mmr || {
            cp -r ${bryce-mmrs-py} python/mmr
          }
          export PYTHON=${mass-python}/bin/python
          export PYTHONPATH=$PYTHONPATH:$PWD/python
          export TEST_DATA_OUT=$PWD/vectors
        '';
      };
      packages.default = pkgs.buildGoApplication {
          name = "MassMarket Test Vectors";
          modules = ./gomod2nix.toml;
          # go = pkgs.go_1_23;
          buildInputs = [ pkgs.go_1_23 ];
          src = ./.;
          buildPhase = ''
            test -d go || {
              echo "go/ directory not found, skipping vector generation"
              exit 0
            }
            cd ./go
            mkdir -p $out
            export TEST_DATA_OUT=$out
            go test
            cd ../python
            ${mass-python}/bin/python generate_hamt_test_vectors.py
            '';
        };
    });
}
