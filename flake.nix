# SPDX-FileCopyrightText: 2024 Mass Labs
#
# SPDX-License-Identifier: MIT

{
  description = "Mass Market Contracts";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-24.05";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    nixpkgs,
    utils,
    ...
  }:
    utils.lib.eachDefaultSystem (system: let
      pkgs = import nixpkgs {
        inherit system;
        overlays = [ ];
      };

      # web3 needs parsimonious v0.9.0
      # https://github.com/ethereum/web3.py/issues/3110#issuecomment-1737826910
      packageOverrides = self: super: {
        parsimonious = super.parsimonious.overridePythonAttrs (old: rec {
          pname = "parsimonious";
          version = "0.9.0";
          src = pkgs.python3.pkgs.fetchPypi {
            inherit pname version;
            sha256 = "sha256-sq0a5jovZb149eCorFEKmPNgekPx2yqNRmNqXZ5KMME=";
          };
          doCheck = false;
        });
      };

      pinnedPython = pkgs.python3.override {
        inherit packageOverrides;
        self = pkgs.python3;
      };

      protobuf_to_dict = pkgs.python3Packages.buildPythonPackage rec {
        pname = "protobuf-to-dict";
        version = "0.3.1";
        src = fetchGit {
          url = "https://github.com/masslbs/protobuf-to-dict.git";
          rev = "39d7ec2a3a72b5938fe9bddbc593d210bccb64b8";
          ref = "patch-reqs";
        };
        propagatedBuildInputs = [
          pkgs.python3Packages.pip
          pkgs.python3Packages.six
          pkgs.python3Packages.nose
          pkgs.python3Packages.dateutil
        ];
        doCheck = false;
      };

      mass-python = pinnedPython.withPackages (ps:
        with ps; [
          protobuf
          protobuf_to_dict
          web3
          safe-pysha3
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
        buf
        black
        reuse
        protobuf
        protolint
        pyright
        mass-python
      ];
    in {
      devShell = pkgs.mkShell {
        inherit buildInputs;

        shellHook = ''
          export PYTHON=${mass-python}/bin/python
          export PYTHONPATH=$PYTHONPATH:$PWD/python
        '';
      };
    });
}
