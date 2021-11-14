ARG UBUNTU_BUILDER=ubuntu:18.04
ARG UBUNTU_RUNNER=quay.io/team-helium/blockchain-node:blockchain-node-ubuntu18-1.1.47

FROM $UBUNTU_BUILDER as rosetta-builder


WORKDIR /src

RUN apt update \
      && apt install -y --no-install-recommends \
         curl ca-certificates git \
      && curl -L https://golang.org/dl/go1.17.1.linux-amd64.tar.gz | tar xzf -

ENV PATH="/src/go/bin:$PATH" \
    CGO_ENABLED=0

COPY . rosetta-helium

RUN cd rosetta-helium && go build -o rosetta-helium

FROM $UBUNTU_RUNNER

EXPOSE 8080
EXPOSE 44158

RUN apt update \
    && apt install -y --no-install-recommends \
         ca-certificates git npm

WORKDIR /app

COPY --from=rosetta-builder /src/rosetta-helium/rosetta-helium rosetta-helium
COPY --from=rosetta-builder /src/rosetta-helium/docker/start.sh start.sh
COPY --from=rosetta-builder /src/rosetta-helium/helium-constructor helium-constructor

RUN cd helium-constructor \
      && npm install \
      && npm run build \
      && chmod +x /app/start.sh \
      && cat /opt/blockchain_node/config/sys.config | grep -oP '(?<=\{blessed_snapshot_block_height\, ).*?(?=\})' > /app/lbs.txt

ENTRYPOINT ["/app/start.sh"]
