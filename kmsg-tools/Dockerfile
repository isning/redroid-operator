# Stage 1: Build statically linked socat using Nix
FROM docker.io/nixos/nix:latest AS builder

# Enable flakes
RUN echo "experimental-features = nix-command flakes" >> /etc/nix/nix.conf

WORKDIR /build
COPY flake.nix flake.lock ./

# Build the musl-static socat binary using the pinned flake
RUN nix build path:.#default -o /tmp/result

# Stage 2: Minimal runtime image to deliver the binary
# We use alpine over scratch so we have /bin/sh to execute the copy command
FROM docker.io/library/alpine:3.19

# Copy the static binary from the Nix build stage
COPY --from=builder /tmp/result/bin/socat /bin/socat

# Verify it runs and is ready to be injected
RUN /bin/socat -d -d -h | head -n 1
