{
  description = "A media player for Discord";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs = { self, nixpkgs, ... }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; config = { allowUnfree = true; }; };
  in {
    packages.${system}.default = pkgs.callPackage ./package.nix {
      stdenv = pkgs.clangStdenv;
    };
    devShells.${system}.default = pkgs.callPackage ./devshell.nix {
      fretboard = self.packages.${system}.default;
    };
  };
}
