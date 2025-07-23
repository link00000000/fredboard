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

  nixpkgs.config.allowUnfree = true;
  nix.settings = {
    trusted-users = [ "@wheel" ];
    experimental-features = [ "nix-command" "flakes" ];
  };

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
      (forward 4317)
      (forward 4318)
    ];
  };

  networking.firewall.enable = false;
  users.users.root.password = "";
  users.motd = builtins.concatStringsSep "\n" [
    "Grafana: http://localhost:${builtins.toString grafanaCfg.settings.server.http_port}"
    "QEMU Help: Ctrl-b h"
  ];
  services.getty.autologinUser = "root";

  /************************************************** Service overview ***************************************************
   *
   *                                    ┌────────────────┐  All Metrics    ┌──────────┐                 ┌────────────────┐
   *                                    │                ├────────────────►│          │                 │                │
   *                                    │                │                 │          │  All Metrics    │                │
   *                                    │                │  Prometheus     │Prometheus├────────────────►│                │
   *                                    │                │  Telemetry      │          │                 │                │
   *                                    │                │◄────────────────┤          │                 │                │
   *                                    │                │                 └──────────┘                 │                │
   *                                    │                │                                              │                │
   * ┌─────────────────┐                │                │  All Spans      ┌──────────┐                 │                │
   * │                 │  Fredboard     │                ├────────────────►│          │                 │                │
   * │                 │  Telemetry     │                │                 │          │  All Spans      │                │
   * │    Fredboard    ├───────────────►│ OTel Collector │  Tempo          │  Tempo   ├────────────────►│    Grafana     │
   * │                 │                │                │  Telemetry      │          │                 │                │
   * │                 │                │                │◄────────────────┤          │                 │                │
   * └─────────────────┘                │                │                 └──────────┘                 │                │
   *                                    │                │                                              │                │
   *                                    │                │  All Logs       ┌──────────┐                 │                │
   *                                    │                ├────────────────►│          │                 │                │
   *                                    │                │                 │          │  All Logs       │                │
   *                                    │                │  Loki           │   Loki   ├────────────────►│                │
   *                                    │                │  Telemetry      │          │                 │                │
   *                                    │                │◄────────────────┤          │                 │                │
   *                                    └───────────────┬┘                 └──────────┘                 └────────────────┘
   *                                     ▲              │                                                                 
   *                                     │              │                                                                 
   *                                     └──────────────┘                                                                 
   *                                      OTel Collector                                                                  
   *                                      Telemetry                                                                       
   */

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
    extraFlags = [ 
      "--web.enable-remote-write-receiver" # Enables the remote-write receiver feature required for the OTel collector to push metrics to prometheus
    ];
  };

  services.tempo = {
    enable = true;

    # docs: https://grafana.com/docs/tempo/latest/configuration/#configure-tempo
    settings = {

      # Tempo uses the server from `dskit/server`.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#server
      server = {
        http_listen_port = 3200;
        grpc_listen_port = 9096;
      };

      # Distributors receive spans and forward them to the appropriate ingesters.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#distributor
      distributor = {
        # Receiver entry is equivalent in format to the receiver node in the otel collector:
        # https://github.com/open-telemetry/opentelemetry-collector/tree/main/receiver
        receivers = {
          otlp.protocols = {
            grpc.endpoint = "0.0.0.0:4417";
            http.endpoint = "0.0.0.0:4418";
          };
        };
      };

      # The ingester is responsible for batching up traces and pushing them to TempoDB.
      # 
      # A live, or active, trace is a trace that has received a new batch of spans in more than a configured amount of time
      # (default 10 seconds, set by ingester.trace_idle_period). After 10 seconds (or the configured amount of time),
      # the trace is flushed to disk and appended to the WAL. When Tempo receives a new batch, a new live trace is created in memory.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#ingester
      ingester = {
        # Amount of time a trace must be idle before flushing it to the wal.
        # (default: 10s)
        trace_idle_period = "10s";

        # Maximum length of time before cutting a block
        # (default: 30m)
        max_block_duration = "5m";

        # Flush all traces to backend when ingester is stopped
        # (default: false)
        flush_all_on_shutdown = true;
      };

      # The metrics-generator processes spans and write metrics using the Prometheus remote write protocol.
      #
      # Metrics-generator processors are disabled by default. To enable it for a specific tenant, set `metrics_generator.processors`
      # in the overrides section.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#metrics-generator
      metrics_generator = {};

      # The Query Frontend is responsible for sharding incoming requests for faster processing in parallel (by the queriers).
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#query-frontend
      query_frontend = {};

      # The Querier is responsible for querying the backends/cache for the traceID.
      #
      # It also queries compacted blocks that fall within the (2 * BlocklistPoll) range where the value of Blocklist poll
      # duration is defined in the storage section.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#querier
      querier = {};

      # Compactors stream blocks from the storage backend, combine them and write them back.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#compactor
      compactor = {};

      # Tempo supports Amazon S3, GCS, Azure, and local file system for storage. In addition, you can use Memcached or
      # Redis for increased query performance.
      #
      # While you can use local storage, object storage is recommended for production workloads. A local backend won’t correctly retrieve traces
      # with a distributed deployment unless all components have access to the same disk. Tempo is designed for object storage more than local storage.
      # 
      # You can estimate how much storage space you need by considering the ingested bytes and retention. For example, ingested bytes per day * retention days = stored bytes.
      #
      # You can not use both local and object storage in the same Tempo deployment.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#storage
      storage = {
        trace = {
          # The storage backend to use
          # Should be one of "gcs", "s3", "azure" or "local" (only supported in the monolithic mode)
          # CLI flag -storage.trace.backend
          backend = "local";

          local.path = "/var/lib/tempo/data/tempo/blocks";
          wal.path = "/var/lib/tempo/data/tempo/wal";
        };
      };

      # Memberlist is the default mechanism for all of the Tempo pieces to coordinate with each other.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#memberlist
      memberlist = {};

      # Tempo provides an overrides module for users to set global or per-tenant override settings.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#overrides
      overrides = {
        defaults = {
          # Enables metric generation from Tempo.
          metrics_generator.processors = [ "service-graphs" "span-metrics" "local-blocks" ];
        };
      };

      # By default, Tempo reports anonymous usage data about the shape of a deployment to Grafana Labs. This data is used to determine how common the deployment of
      # certain features are, if a feature flag has been enabled, and which replication factor or compression levels are used.
      #
      # By providing information on how people use Tempo, usage reporting helps the Tempo team decide where to focus their development and documentation
      # efforts. No private information is collected, and all reports are completely anonymous.
      #
      # The following configuration values are used:
      #
      # * Receivers enabled
      # * Frontend concurrency and version
      # * Storage cache, backend, WAL and block encodings
      # * Ring replication factor, and kvstore
      # * Features toggles enabled
      #
      # No performance data is collected.
      #
      # You can view the report by visiting this address on your Tempo instance: http://localhost:3200/status/usage-stats
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#usage-report
      usage_report = {
        reporting_enabled = false;
      };

      # Use this block to configure caches available throughout the application. Multiple caches can be created and assigned roles which
      # determine how they are used by Tempo.
      #
      # docs: https://grafana.com/docs/tempo/latest/configuration/#cache
      cache = {};
    };
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

  services.opentelemetry-collector = {
    enable = true;

    # The structure of any Collector configuration file consists of four classes of pipeline components that access telemetry data:
    #
    # * Receivers
    # * Processors
    # * Exporters
    # * Connectors 
    #
    # After each pipeline component is configured you must enable it using the pipelines within the service section of the configuration file.
    #
    # Besides pipeline components you can also configure extensions, which provide capabilities that can be added to the Collector,
    # such as diagnostic tools. Extensions don’t require direct access to telemetry data and are enabled through the service section.
    #
    # docs: https://opentelemetry.io/docs/collector/configuration/#basics
    settings = {
      # Receivers collect telemetry from one or more sources. They can be pull or push based, and may support one or more data sources.
      # Receivers are configured in the receivers section. Many receivers come with default settings, so that specifying the name of the
      # receiver is enough to configure it. If you need to configure a receiver or want to change the default configuration, you can do
      # so in this section. Any setting you specify overrides the default values, if present.
      #
      # Configuring a receiver does not enable it. Receivers are enabled by adding them to the appropriate pipelines within the service section.
      #
      # The Collector requires one or more receivers.
      #
      # docs: https://opentelemetry.io/docs/collector/configuration/#receivers
      receivers = {
        otlp = {
          protocols = {
            grpc.endpoint = "0.0.0.0:4317";
            http.endpoint = "0.0.0.0:4318";
          };
        };
      };

      # Processors take the data collected by receivers and modify or transform it before sending it to the exporters. Data processing
      # happens according to rules or settings defined for each processor, which might include filtering, dropping, renaming, or
      # recalculating telemetry, among other operations. The order of the processors in a pipeline determines the order of the processing
      # operations that the Collector applies to the signal.
      #
      # Processors are optional, although some are recommended.
      #
      # You can configure processors using the processors section of the Collector configuration file. Any setting you specify overrides
      # the default values, if present.
      #
      # Configuring a processor does not enable it. Processors are enabled by adding them to the appropriate pipelines within the
      # service section.
      #
      # docs: https://opentelemetry.io/docs/collector/configuration/#processors
      processors = {
        memory_limiter = {
          check_interval = "5s";
          limit_mib = 4000;
          spike_limit_mib = 500;
        };
      };

      # Exporters send data to one or more backends or destinations. Exporters can be pull or push based, and may support one or
      # more data sources.
      #
      # Each key within the exporters section defines an exporter instance, The key follows the type/name format, where type specifies
      # the exporter type (e.g., otlp, kafka, prometheus), and name (optional) can be appended to provide a unique name for multiple
      # instance of the same type.
      #
      # Most exporters require configuration to specify at least the destination, as well as security settings, like authentication
      # tokens or TLS certificates. Any setting you specify overrides the default values, if present.
      #
      # Configuring an exporter does not enable it. Exporters are enabled by adding them to the appropriate pipelines within the service section.
      #
      # The Collector requires one or more exporters.
      #
      # docs: https://opentelemetry.io/docs/collector/configuration/#exporters
      exporters = {
        prometheusremotewrite = {
          endpoint = "http://localhost:${builtins.toString prometheusCfg.port}/api/v1/write";
          tls.insecure = true;
        };
        "otlp/tempo" = {
          endpoint = "http://localhost:4417";
          tls.insecure = true;
        };
        "otlphttp/loki" = {
          endpoint = "http://localhost:${builtins.toString lokiCfg.configuration.server.http_listen_port}/otlp";
          tls.insecure = true;
        };
      };

      # Connectors join two pipelines, acting as both exporter and receiver. A connector consumes data as an exporter at the end of one
      # pipeline and emits data as a receiver at the beginning of another pipeline. The data consumed and emitted may be of the same
      # type or of different data types. You can use connectors to summarize consumed data, replicate it, or route it.
      #
      # You can configure one or more connectors using the connectors section of the Collector configuration file. By default,
      # no connectors are configured. Each type of connector is designed to work with one or more pairs of data types and may only
      # be used to connect pipelines accordingly.
      #
      # Configuring a connector doesn’t enable it. Connectors are enabled through pipelines within the service section.
      #
      # docs: https://opentelemetry.io/docs/collector/configuration/#connectors
      connectors = {};

      # Extensions are optional components that expand the capabilities of the Collector to accomplish tasks not directly involved
      # with processing telemetry data. For example, you can add extensions for Collector health monitoring, service discovery, or
      # data forwarding, among others.
      #
      # You can configure extensions through the extensions section of the Collector configuration file. Most extensions come with
      # default settings, so you can configure them just by specifying the name of the extension. Any setting you specify
      # overrides the default values, if present.
      #
      # Configuring an extension doesn’t enable it. Extensions are enabled within the service section.
      #
      # By default, no extensions are configured.
      #
      # docs: https://opentelemetry.io/docs/collector/configuration/#extensions
      extensions = {};

      # The service section is used to configure what components are enabled in the Collector based on the configuration found in
      # the receivers, processors, exporters, and extensions sections. If a component is configured, but not defined within the
      # service section, then it’s not enabled.
      #
      # docs: https://opentelemetry.io/docs/collector/configuration/#service
      service = {
        # The extensions subsection consists of a list of desired extensions to be enabled.
        #
        # docs: https://opentelemetry.io/docs/collector/configuration/#service-extensions
        extensions = [];

        # The pipelines subsection is where the pipelines are configured, which can be of the following types:
        #
        # * `traces` collect and processes trace data.
        # * `metrics` collect and processes metric data.
        # * `logs` collect and processes log data.
        #
        # A pipeline consists of a set of receivers, processors and exporters. Before including a receiver, processor, or exporter
        # in a pipeline, make sure to define its configuration in the appropriate section.
        #
        # You can use the same receiver, processor, or exporter in more than one pipeline. When a processor is referenced in
        # multiple pipelines, each pipeline gets a separate instance of the processor.
        #
        # Note that the order of processors dictates the order in which data is processed.
        #
        # As with components, use the type[/name] syntax to create additional pipelines for a given type. Here is an example
        # extending the previous configuration.
        #
        # docs: https://opentelemetry.io/docs/collector/configuration/#pipelines
        pipelines = {
          metrics = {
            receivers = [ "otlp" ];
            processors = [ "memory_limiter" ];
            exporters = [ "prometheusremotewrite" ];
          };
          traces = {
            receivers = [ "otlp" ];
            processors = [ "memory_limiter" ];
            exporters = [ "otlp/tempo" ];
          };
          logs = {
            receivers = [ "otlp" ];
            processors = [ "memory_limiter" ];
            exporters = [ "otlphttp/loki" ];
          };
        };

        # The telemetry config section is where you can set up observability for the Collector itself.
        # It consists of two subsections: logs and metrics.
        #
        # docs: https://opentelemetry.io/docs/collector/configuration/#telemetry
        telemetry = {};
      };
    };
  };
}

