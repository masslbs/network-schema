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
    # if we ever need to refresh the mmr python code
    # bryce-mmrs-py = {
    #   url = "github:robinbryce/draft-bryce-cose-merkle-mountain-range-proofs";
    #   flake = false;
    # };
  };

  outputs = {
    nixpkgs,
    utils,
    gomod2nix,
    # bryce-mmrs-py,
    ...
  }:
    utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [gomod2nix.overlays.default];
      };

      ourPython = pkgs.python312;

      # https://github.com/ethereum/eth-typing/issues/63#issuecomment-2291678106
      pyunormalize = ourPython.pkgs.buildPythonPackage rec {
        pname = "pyunormalize";
        version = "15.1.0";
        src = ourPython.pkgs.fetchPypi {
          inherit pname version;
          hash = "sha256-z0qHRRoPHLdpEaqX9DL0V54fVkovDITOSIxzpzkBtsE=";
        };
        buildInputs = [pinnedPython.pkgs.setuptools];
        # le sigh...
        postUnpack = ''
          sed -i 's/version=get_version()/version="${version}"/' ${pname}-${version}/setup.py
          sed -i 's/del _version//' ${pname}-${version}/${pname}/__init__.py
        '';
      };

      packageOverrides = self: super: {
        parsimonious = super.parsimonious.overridePythonAttrs (old: rec {
          pname = "parsimonious";
          version = "0.10.0";
          src = self.fetchPypi {
            inherit pname version;
            sha256 = "sha256-goFgDaGA7IrjVCekq097gr/sHj0eUvgMtg6oK5USUBw=";
          };
        });

        websockets = super.websockets.overridePythonAttrs (old: rec {
          pname = "websockets";
          version = "13.1";
          src = self.fetchPypi {
            inherit pname version;
            sha256 = "sha256-o7M2YIfBvAonlREe3K3duLO1lQnV211+o/3Wn5VKiHg=";
          };
          doCheck = false;
        });

        web3 = super.web3.overridePythonAttrs (old: rec {
          pname = "web3";
          version = "7.8.0";
          src = self.fetchPypi {
            inherit pname version;
            sha256 = "sha256-cSvJ/Wse9uRn7iTCW1geGVHKssuhf59UjxJYdzTyyFc=";
          };
          propagatedBuildInputs =
            super.web3.propagatedBuildInputs
            ++ [
              pyunormalize
            ];
          doCheck = false;
        });

        # transient test failures
        asn1tools = super.asn1tools.overridePythonAttrs (old: {
          doCheck = false; # TODO: just on darwin..?
        });

        eth-account = super.eth-account.overridePythonAttrs (old: {
          doCheck = false;
          doInstallCheck = false;
          propagatedBuildInputs =
            super.eth-account.propagatedBuildInputs
            ++ [
              pinnedPython.pkgs.pydantic
            ];
        });

        django = super.django.overrideAttrs (old: rec {
          doInstallCheck = false;
        });
      };

      pinnedPython = ourPython.override {
        inherit packageOverrides;
        self = ourPython;
      };

      mass-python = pinnedPython.withPackages (
        ps:
          with ps; [
            protobuf
            web3

            matplotlib # go/hamt_bench_plot.py
            cbor2

            # packaging massmarket
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
        cddl # can be used to generate json from cddl files
        cbor-diag
        protoc-gen-go
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
          # vendored mmr code
          # TODO: this should be an actual python package
          # but i dont want to package it up right now
          # test -d python/massmarket/mmr || {
          #   cp -r $\{bryce-mmrs-py\} python/massmarket/mmr
          # }
          export PYTHON=${mass-python}/bin/python
          export PYTHONPATH=$PYTHONPATH:$PWD/python
          export TEST_DATA_OUT=$PWD/vectors
        '';
      };
      packages.default = pkgs.buildGoApplication {
        name = "MassMarket Test Vectors";
        modules = ./gomod2nix.toml;
        buildInputs = [pkgs.go_1_23];
        src = ./.;
        buildPhase = ''
          mkdir -p $out
          cp constants.txt $out
          cp VERSION $out
          cp -r cddl $out/cddl
          test -d go || {
            echo "go/ directory not found, skipping vector generation"
            exit 0
          }
          mkdir -p $out/vectors
          export TEST_DATA_OUT=$out/vectors
          pushd go/patch
          go test
          popd
          cd python
          ${mass-python}/bin/python generate_hamt_test_vectors.py
        '';
      };
    });
}
