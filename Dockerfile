# Multi-stage build for dabazo.
# Target image is Debian-based so the `apt` package-manager driver works
# end-to-end inside the container (install/start/stop postgres, etc.).

# ---------- build stage ----------
FROM golang:1.26-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/dabazo ./cmd/dabazo

# ---------- runtime stage ----------
FROM debian:bookworm-slim

# Runtime deps dabazo shells out to:
#   - sudo: dabazo install/start/stop prefix apt-get and service commands with sudo
#   - ca-certificates: for any outbound https from apt
#   - gnupg + lsb-release: apt repo key/dist probing
#   - postgresql-client: so `dabazo migrate` and `dabazo snapshot` have psql + pg_dump available
#     (the server itself is installed by `dabazo install` at runtime)
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
        sudo \
        ca-certificates \
        gnupg \
        lsb-release \
        postgresql-client \
 && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/dabazo /usr/local/bin/dabazo

# Non-root user with passwordless sudo so `dabazo install` can invoke apt-get.
# This is a *test* image — do not use this sudo policy in production.
RUN useradd --create-home --shell /bin/bash dabazo \
 && echo 'dabazo ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/dabazo \
 && chmod 0440 /etc/sudoers.d/dabazo

USER dabazo
WORKDIR /home/dabazo

# Default to an interactive shell so you can run `dabazo ...` by hand.
# Build:  docker build -t dabazo .
# Run:    docker run -it --rm dabazo
# Inside: dabazo help
#         dabazo install --db postgres:16 --port 5432 --name dev
#         dabazo start
#         ...
CMD ["bash"]
