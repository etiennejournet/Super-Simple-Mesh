FROM envoyproxy/envoy-alpine:v1.18-latest
LABEL tag=1.0.0

COPY resources/ /

RUN /setup.sh && rm /setup.sh

CMD /run.sh
