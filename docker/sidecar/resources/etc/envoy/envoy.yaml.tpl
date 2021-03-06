node:
  id: test
  cluster: test
admin:
  address:
    socket_address:
      protocol: TCP
      address: 0.0.0.0
      port_value: 9901
static_resources:
  listeners:
  - name: listener_outbound
    address:
      socket_address:
        protocol: TCP
        address: 0.0.0.0
        port_value: 10001
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager"
          access_log:
          - name: envoy.access_loggers.file
            typed_config:
              "@type": "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog"
              path: /dev/stdout
          stat_prefix: ingress_http
          strip_any_host_port: true # This allows having a svc port different than the container port
          route_config:
            name: external_route
            virtual_hosts:
            - name: local_k8s
              domains: ["*.cluster.local"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: local_k8s
            - name: external_k8s
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: passthrough
          http_filters:
          - name: envoy.filters.http.router
    - filter_chain_match:
        transport_protocol: tls
      filters:
      - name: envoy.filters.network.tcp_proxy
        typed_config:
          "@type": "type.googleapis.com/envoy.extensions.filters.network.tcp_proxy.v3.TcpProxy"
          stat_prefix: ingress_tcp
          cluster: passthrough
    listener_filters:
    - name: envoy.filters.listener.original_dst
    - name: envoy.filters.listener.tls_inspector
  - name: listener_inbound
    address:
      socket_address:
        protocol: TCP
        address: 0.0.0.0
        port_value: 10000
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager"
          access_log:
          - name: envoy.access_loggers.file
            typed_config:
              "@type": "type.googleapis.com/envoy.extensions.access_loggers.file.v3.FileAccessLog"
              path: /dev/stdout
          stat_prefix: ingress_http
          route_config:
            name: local_route
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match:
                  prefix: "/"
                route:
                  cluster: passthrough
          http_filters:
          - name: envoy.filters.http.router
      transport_socket:
        name: envoy.transport_sockets.tls
        typed_config:
          "@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.DownstreamTlsContext"
          common_tls_context:
            tls_certificate_sds_secret_configs:
              name: cert_sds
              sds_config:
                path: /etc/envoy/cert_sds.yaml
    listener_filters:
    - name: envoy.filters.listener.original_dst
  clusters:
  - name: passthrough
    connect_timeout: 30s
    type: original_dst
    dns_lookup_family: V4_ONLY
    lb_policy: CLUSTER_PROVIDED
  - name: local_k8s
    connect_timeout: 30s
    type: original_dst
    dns_lookup_family: V4_ONLY
    lb_policy: CLUSTER_PROVIDED
    transport_socket:
      name: envoy.transport_sockets.tls
      typed_config:
        "@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.UpstreamTlsContext"
        common_tls_context:
          validation_context:
            trusted_ca:
              filename: {{ env.Getenv "CERTIFICATES_PATH" }}/ca.crt
            match_subject_alt_names:
