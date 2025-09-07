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
    ];
}
