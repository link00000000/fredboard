{ config, ... }:

let
  grafanaCfg = config.services.grafana;
  lokiCfg = config.services.loki;
  tempoCfg = config.services.tempo;
  prometheusCfg = config.services.prometheus;
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
      (forward config.services.grafana.settings.server.http_port)
      (forward lokiCfg.configuration.server.http_listen_port)
      (forward tempoCfg.settings.server.http_listen_port)
      (forward prometheusCfg.port)
    ];
  };

  networking.firewall.enable = false;
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
          url = "http://localhost:${builtins.toString lokiCfg.configuration.server.http_listen_port}";
        }
        {
          name = "Tempo";
          type = "tempo";
          access = "proxy";
          url = "http://localhost:${builtins.toString tempoCfg.settings.server.http_listen_port}";
        }
        {
          name = "Prometheus";
          type = "prometheus";
          access = "proxy";
          url = "http://localhost:${builtins.toString prometheusCfg.port}";
        }
      ];
    };
  };

  services.prometheus = {
    enable = true;
    listenAddress = "0.0.0.0";
    port = 9009;
    scrapeConfigs = [
      {
        job_name = "loki";
        static_configs = [{ targets = [ "localhost:${builtins.toString lokiCfg.configuration.server.http_listen_port}" ]; }];
      }
      {
        job_name = "tempo";
        static_configs = [{ targets = [ "localhost:${builtins.toString tempoCfg.settings.server.http_listen_port}" ]; }];
      }
      {
        job_name = "grafana";
        static_configs = [{ targets = [ "localhost:${builtins.toString grafanaCfg.settings.server.http_port}" ]; }];
      }
    ];
  };

  services.loki = {
    enable = true;
    configuration = {
      auth_enabled = false;

      server = {
        http_listen_port = 3100;
        grpc_listen_port = 9095;
      };

      ingester = {
        wal.dir = "/tmp/loki/wal";
        lifecycler = {
          address = "127.0.0.1";
          ring = {
            kvstore.store = "inmemory";
            replication_factor = 1;
          };
        };
        chunk_idle_period = "5m";
        chunk_retain_period = "30s";
        chunk_target_size = 1536000;
        chunk_block_size = 262144;
      };

      schema_config.configs = [
        {
          from = "2023-01-01";
          store = "tsdb";
          object_store = "filesystem";
          schema = "v13";
          index = {
            prefix = "index_";
            period = "24h";
          };
        }
      ];

      storage_config.tsdb_shipper = {
          active_index_directory = "/var/lib/loki/index";
          cache_location = "/var/lib/loki/index_cache";
          cache_ttl = "24h";
      };

      chunk_store_config = {};

      compactor = {
        working_directory = "/var/lib/loki/compactor";
      };

      limits_config = {
        reject_old_samples = true;
        reject_old_samples_max_age = "168h";
      };

      table_manager = {
        retention_deletes_enabled = true;
        retention_period = "168h";
      };

      ruler = {
        enable_api = true;
        storage = {
          type = "local";
          local.directory = "/var/lib/loki/rules";
        };
        rule_path = "/var/lib/loki/rules";
        ring.kvstore.store = "inmemory";
      };
    };
  };

  services.tempo = {
    enable = true;
    settings = {
      server = {
        http_listen_port = 3200;
        grpc_listen_port = 9096;
      };

      distributor = {
        receivers.otlp.protocols = {
          grpc = {};
          http = {};
        };
      };

      ingester = {
        trace_idle_period = "10s";
        max_block_duration = "5m";
      };

      compactor = {
        compaction.block_retention = "1h";
      };

      storage.trace = {
        backend = "local";
        local.path = "/var/lib/tempo/data/tempo/blocks";
        wal.path = "/var/lib/tempo/data/tempo/wal";
      };
    };
  };
}

