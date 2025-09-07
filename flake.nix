{
  description = "A music player bot for Discord";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs = { self, nixpkgs, flake-utils, gomod2nix, ... }@inputs: flake-utils.lib.eachDefaultSystem (system: {
    devShells = {
      default = self.devShells.${system}.full;
      full = import ./nix/dev-shells/full.nix system inputs;
      minimal = import ./nix/dev-shells/minimal.nix system inputs;
    };
    packages = {
      default = self.packages.${system}.fredboard;
      fredboard = import ./nix/packages/fredboard.nix system inputs;
    };
    apps = {
      fredboard-monitoring = {
        type = "app";
        program = "${self.packages.${system}.monitoring-vm}/bin/start-monitoring-vm";
      };
    };
  });
}
