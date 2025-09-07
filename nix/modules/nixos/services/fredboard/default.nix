{ config, lib, pkgs, ... }:
let
  cfg = config.services.fredboard;
in
{
  options.services.fredboard = {
    enable = lib.mkEnableOption "Enable fredboard service";
    package = lib.mkOption {
      type = lib.types.package;
      default = pkgs.fredboard;
      description = "Package for fredboard";
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.fredboard = {
      description = "Fredboard Discord bot";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];
      serviceConfig = {
        ExecStart = "${cfg.package}/bin/fredboard";
        Restart = "on-failure";
      };
    };
  };
}
