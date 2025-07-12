{
  description = "A music player bot for Discord";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { nixpkgs, flake-utils, ... }: flake-utils.lib.eachDefaultSystem (system: let
    pkgs = import nixpkgs { inherit system; };
    
    buildInputs = with pkgs; [
      gnumake
      go_1_23
      golangci-lint
      libGL
      xorg.libX11
    ];

    runtimeDependencies = with pkgs; [
      ffmpeg
      yt-dlp
    ];
  in {
    devShells.default = pkgs.mkShell {
      packages = with pkgs; [
        delve
        dotenv-cli
        go-tools
        gopls
        gotools
        graphviz
        hexyl
        vlc
      ]
      ++ buildInputs
      ++ runtimeDependencies;

      # Required to prevent error when running `dlv test`
      hardeningDisable = [ "fortify" ];
    };

    packages.default = pkgs.buildGo123Module {
      pname = "fredboard";
      version = builtins.readFile ./version;
      src = ./.;

      subPackages = [
        "cmd/fredboard_server"
      ];

      vendorHash = "sha256-urCKBLWE4Pjiy1q0hSxEzu8XK1OR2VMkNyHSnai3EkM=";

      # gopus package has C files that are not properly vendored by 'go mod vendor',
      # so we need to manually add them to the vendored directory.
      preBuild = let
        gopusRepo = let
          goMod = builtins.readFile ./go.mod;
          matches = builtins.match ".*layeh.com/gopus v[0-9.]+-[0-9]+-([a-f0-9]+).*" goMod;
          rev = if matches != null && matches != [] then (builtins.head matches) else throw "gopus rev not found";
        in pkgs.fetchFromGitHub {
          inherit rev;
          owner = "layeh";
          repo = "gopus";
          sha256 = "sha256-i1N5ETtqTfmLZ3yb4yhZXJsGBsykxFyifdSo94Th8RU=";
        };
      in /* bash */ ''
        # vendor/ is not writable, so we need to make a copy of everything to a writable version of vendor/
        mv vendor vendor.old
        mkdir -p vendor/layeh.com/gopus/opus-1.1.2
        cp -r ${gopusRepo}/opus-1.1.2/* vendor/layeh.com/gopus/opus-1.1.2/
        cp -r vendor.old/* vendor/
      '';

      meta = {
        description = "A music player bot for Discord";
        homepage = "https://github.com/link00000000/fredboard";
        license = pkgs.lib.licenses.mit;
        maintainers = with pkgs.lib.maintainers; [ link00000000 ];
      };
    };
  });
}
