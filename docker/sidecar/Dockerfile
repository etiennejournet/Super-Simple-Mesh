FROM envoyproxy/envoy-alpine:v1.18-latest

COPY resources/ /

USER root
RUN /setup.sh && rm /setup.sh

USER envoy
CMD /run.sh
