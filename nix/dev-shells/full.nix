system: { nixpkgs, gomod2nix, ... }:
let
  pkgs = import nixpkgs {
    inherit system;
    overlays = [ gomod2nix.overlays.default ];
  };
  goEnv = pkgs.mkGoEnv {
    pwd = ../..;
  };
in pkgs.mkShell {
    packages = with pkgs; [
      goEnv

      # Required runtime packages
      ffmpeg
      yt-dlp

      # Other useful tools
      delve
      dotenv-cli
      go-tools
      gomod2nix.packages.${system}.default
      gopls
      gotools
      graphviz
      hexyl
      vlc
    ];

    # Required to prevent error when running `dlv test`
    hardeningDisable = [ "fortify" ];
}
