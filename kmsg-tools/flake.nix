{
  description = "kmsg-tools flake for static socat";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }: {
    packages.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.pkgsStatic.socat;
    packages.aarch64-linux.default = nixpkgs.legacyPackages.aarch64-linux.pkgsStatic.socat;
  };
}
