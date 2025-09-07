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

  outputs = { self, nixpkgs, flake-utils, gomod2nix, ... }: flake-utils.lib.eachDefaultSystem (system: let
    pkgs = import nixpkgs {
      inherit system;
      overlays = [ gomod2nix.overlays.default ];
    };
    
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
        gomod2nix.packages.${system}.default
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

    packages = {
      default = self.packages.${system}.fredboard-server;

      fredboard-server = pkgs.buildGoApplication {
        pname = "fredboard";
        version = "dev";
        inherit buildInputs;
        src = ./.;
        modules = ./gomod2nix.toml;

        subPackages = [
          "cmd/fredboard_server"
        ];

        meta = {
          description = "A music player bot for Discord";
          homepage = "https://github.com/link00000000/fredboard";
          license = pkgs.lib.licenses.mit;
          maintainers = with pkgs.lib.maintainers; [ link00000000 ];
        };
      };

      monitoring-vm = pkgs.writeShellScriptBin "start-monitoring-vm" ''
        export QEMU_OPTS="-nographic -serial mon:stdio -echr 0x02"
        ${self.nixosConfigurations.fredboard-monitoring.config.system.build.vm}/bin/run-fredboard-monitoring-vm
      '';
    };

    apps = {
      fredboard-monitoring = {
        type = "app";
        program = "${self.packages.${system}.monitoring-vm}/bin/start-monitoring-vm";
      };
    };
  })
  // {
    nixosConfigurations = {
      fredboard-monitoring = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [ ./nix/nixos-configurations/fredboard-monitoring.nix ];
      };
    };
  };
}
