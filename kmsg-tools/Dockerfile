# Stage 1: Build statically linked socat using Nix
FROM docker.io/nixos/nix:latest AS builder

# Enable flakes
RUN echo "experimental-features = nix-command flakes" >> /etc/nix/nix.conf

WORKDIR /build
COPY flake.nix flake.lock ./

# Build the musl-static socat and busybox binaries using the pinned flake
RUN git init && git add flake.nix flake.lock && \
    nix build path:.#socat -o /tmp/socat && \
    nix build path:.#busybox -o /tmp/busybox

# Stage 2: Minimal runtime image to deliver the binary
# We use alpine over scratch so we have /bin/sh to execute the copy command
FROM docker.io/library/alpine:3.19

# Copy the static binaries from the Nix build stage
COPY --from=builder /tmp/socat/bin/socat /bin/socat
COPY --from=builder /tmp/busybox/bin/busybox /bin/busybox

# Verify they run and are ready to be injected
RUN /bin/socat -d -d -h | head -n 1 && /bin/busybox | head -n 1
