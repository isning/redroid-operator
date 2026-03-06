{
  description = "kmsg-tools flake for static socat";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }: {
    packages.x86_64-linux = {
      socat = nixpkgs.legacyPackages.x86_64-linux.pkgsStatic.socat;
      busybox = nixpkgs.legacyPackages.x86_64-linux.pkgsStatic.busybox;
    };
    packages.aarch64-linux = {
      socat = nixpkgs.legacyPackages.aarch64-linux.pkgsStatic.socat;
      busybox = nixpkgs.legacyPackages.aarch64-linux.pkgsStatic.busybox;
    };
  };
}
