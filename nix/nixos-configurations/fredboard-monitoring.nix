{ config, ... }:

let
  grafanaPort = config.services.grafana.settings.server.http_port;
  #lokiPort = config.services.loki.configuration.server.http_listen_port;
  lokiPort = 3100;
  #tempoPort = config.services.tempo.settings.server.http_listen_port;
  tempoPort = 3200;
  prometheusPort = config.services.prometheus.port;
in

{
  system.stateVersion = "25.05";

  networking.hostName = "fredboard-monitoring";
  time.timeZone = "UTC";

  virtualisation.vmVariant = {
    virtualisation.cores = 4;
    virtualisation.memorySize = 2 * 1024;
    virtualisation.diskSize = 32 * 1024;
    virtualisation.diskImage = "./.qemu/${config.system.name}.qcow2";
    virtualisation.forwardPorts = let
      forward = port: { host.port = port; guest.port = port; };
    in [
      (forward grafanaPort)
      (forward lokiPort)
      (forward tempoPort)
      (forward prometheusPort)
    ] ++ (builtins.map forward config.services.openssh.ports);
  };

  networking.firewall.enable = false;
  users.users.root.password = "root";
  services.getty.autologinUser = "root";
  services.openssh = {
    enable = true;
    ports = [ 2222 ];
  };

  services.grafana = {
    enable = true;
    settings.server = {
      http_port = 3000;
      domain = "localhost";
      http_addr = "0.0.0.0";
    };
    provision = {
      enable = true;
      datasources.settings.datasources = [
        {
          name = "Loki";
          type = "loki";
          access = "proxy";
          url = "http://localhost:${builtins.toString lokiPort}";
        }
        {
          name = "Tempo";
          type = "tempo";
          access = "proxy";
          url = "http://localhost:${builtins.toString tempoPort}";
        }
        {
          name = "Prometheus";
          type = "prometheus";
          access = "proxy";
          url = "http://localhost:${builtins.toString prometheusPort}";
        }
      ];
    };
  };

  services.loki = {
    enable = true;
    configFile = ./loki/config.yaml;
  };

  services.prometheus = {
    enable = true;
    listenAddress = "0.0.0.0";
    port = 9009;
    scrapeConfigs = [
      {
        job_name = "loki";
        static_configs = [{ targets = [ "localhost:${builtins.toString lokiPort}" ]; }];
      }
      {
        job_name = "tempo";
        static_configs = [{ targets = [ "localhost:${builtins.toString tempoPort}" ]; }];
      }
      {
        job_name = "grafana";
        static_configs = [{ targets = [ "localhost:${builtins.toString grafanaPort}" ]; }];
      }
    ];
  };

  services.tempo = {
    enable = true;
    configFile = ./tempo/config.yaml;
  };
}

