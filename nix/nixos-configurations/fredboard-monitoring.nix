{ config, ... }:

{
  system.stateVersion = "25.05";

  networking.hostName = "fredboard-monitoring";
  time.timeZone = "UTC";

  virtualisation.vmVariant = {
    virtualisation.cores = 4;
    virtualisation.memorySize = 2 * 1024;
    virtualisation.diskSize = 32 * 1024;
    #virtualisation.diskImage = "./.qemu/${config.system.name}.qcow2";
    virtualisation.forwardPorts = [
      { host.port = 3000; guest.port = 3000; }
      { host.port = 3100; guest.port = 3100; }
      { host.port = 3200; guest.port = 3200; }
      { host.port = 9009; guest.port = 9009; }
    ];
  };

  users.users.root.password = "";
  services.getty.autologinUser = "root";

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
          url = "http://localhost:3100";
        }
        {
          name = "Tempo";
          type = "tempo";
          access = "proxy";
          url = "http://localhost:3200";
        }
        {
          name = "Prometheus";
          type = "prometheus";
          access = "proxy";
          url = "http://localhost:9009";
        }
      ];
    };
  };

/*
  services.loki = {
    enable = true;
    configuration = {
      auth_enabled = false;
      server.http_listen_port = 3100;
      ingester = {
        lifecycler = {
          ring = {
            kvstore.store = "inmemory";
          };
          final_sleep = "0s";
        };
        chunk_idle_period = "5m";
        max_chunk_age = "1h";
      };
      schema_config = {
        configs = [{
          from = "2022-01-01";
          store = "boltdb-shipper";
          object_store = "filesystem";
          schema = "v11";
          index = {
            prefix = "index_";
            period = "24h";
          };
        }];
      };
      storage_config = {
        boltdb_shipper = {
          active_index_directory = "/tmp/loki/index";
          cache_location = "/tmp/loki/cache";
        };
        filesystem.directory = "/tmp/loki/chunks";
      };
    };
  };
  */

  services.promtail = {
    enable = true;
    configuration = {
      server.http_listen_port = 9080;
      positions = {
        filename = "/tmp/positions.yaml";
      };
      clients = [
        { url = "http://localhost:3100/loki/api/v1/push"; }
      ];
      scrape_configs = [
        {
          job_name = "system";
          static_configs = [{
            targets = ["localhost"];
            labels = {
              job = "syslog";
              __path__ = "/var/log/*.log";
            };
          }];
        }
      ];
    };
  };

  services.tempo = {
    enable = true;
    settings = {
      server.http_listen_port = 3200;
      distributor = { receivers = { otlp = { protocols.grpc.endpoint = "localhost:4317"; }; }; };
      ingester = { };
      compactor = { };
      storage = {
        trace = {
          backend = "local";
          local = { path = "/tmp/tempo/traces"; };
        };
      };
    };
  };

  services.prometheus = {
    enable = true;
    listenAddress = "0.0.0.0";
    port = 9009;
    scrapeConfigs = [
      {
        job_name = "loki";
        static_configs = [{ targets = [ "localhost:3100" ]; }];
      }
      {
        job_name = "tempo";
        static_configs = [{ targets = [ "localhost:3200" ]; }];
      }
      {
        job_name = "grafana";
        static_configs = [{ targets = [ "localhost:3000" ]; }];
      }
    ];
  };
}

