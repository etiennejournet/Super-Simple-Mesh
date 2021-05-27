resources:
- "@type": "type.googleapis.com/envoy.extensions.transport_sockets.tls.v3.Secret"
  name: cert_sds
  tls_certificate:
    certificate_chain:
      filename: {{ env.Getenv "CERTIFICATE_FILE" "/var/run/autocert.step.sm/site.crt" }}
    private_key:
      filename: {{ env.Getenv "PRIVATE_KEY_FILE" "/var/run/autocert.step.sm/site.key" }}
