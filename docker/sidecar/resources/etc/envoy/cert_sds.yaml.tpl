resources:
- "@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret"
  name: cert_sds
  tls_certificate:
    certificate_chain:
      filename: {{ env.Getenv "CERTIFICATES_PATH" }}/tls.crt
    private_key:
      filename: {{ env.Getenv "CERTIFICATES_PATH" }}/tls.key
