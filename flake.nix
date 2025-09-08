{
  description = "A music player bot for Discord";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    gomod2nix = {
      url = "github:nix-community/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, ... }@inputs:
  let
    system = "x86_64-linux";
  in {
    packages.${system} = {
      default = self.packages.${system}.fredboard;
      fredboard = import ./nix/packages/fredboard.nix system inputs;
    };

    devShells.${system} = {
      default = self.devShells.${system}.full;
      full = import ./nix/dev-shells/full.nix system inputs;
      minimal = import ./nix/dev-shells/minimal.nix system inputs;
    };

    nixosModules = {
      default = self.nixosModules.fredboard;
      fredboard = import ./nix/modules/nixos/services/fredboard inputs;
    };
  };
}
