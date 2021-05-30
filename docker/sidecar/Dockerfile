FROM envoyproxy/envoy-alpine:v1.18-latest

COPY resources/ /

RUN /setup.sh && rm /setup.sh

CMD /run.sh
