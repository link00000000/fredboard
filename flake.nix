{
  description = "A Nix-flake-based Go 1.23 development environment";
inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      goVersion = 23; # Change this to update the whole stack

      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        pkgs = import nixpkgs {
          inherit system;
          overlays = [ self.overlays.default ];
        };
      });
    in
    {
      overlays.default = final: prev: {
        go = final."go_1_${toString goVersion}";
      };

      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            # Dev dependencies
            go
            golangci-lint
            gnumake
            air

            # Compile-time dependencies
            xorg.libX11
            libGL
            
            # Runtime dependencies
            yt-dlp
            ffmpeg

            # Tools
            delve
            dotenv-cli
            gotools
            hexyl
            vlc
            graphviz
            gopls
          ];
          hardeningDisable = [ "fortify" ]; # Required to prevent error when running `dlv test`
        };
      });
    };
}
