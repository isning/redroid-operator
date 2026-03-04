{
  description = "redroid-operator and kubectl-redroid plugin";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, pre-commit-hooks }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        # Common Go module build inputs.  The project uses a vendor/ directory so
        # vendorHash = null avoids the extra FOD fetch in the Nix sandbox.
        commonAttrs = {
          src = ./.;
          vendorHash = null;   # vendor/ is committed
          env.CGO_ENABLED = "0";
        };

        version = self.shortRev or self.dirtyShortRev or "dev";

        preCommitCheck = pre-commit-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            # Format all Go source files with gofmt
            gofmt = {
              enable = true;
              # -w rewrites files in place
            };
            # Organise imports (goimports is a superset of gofmt)
            goimports = {
              enable  = true;
              name    = "goimports";
              entry   = "goimports -w";
              types   = [ "go" ];
              # goimports ships with go tools; use the one from nixpkgs
              package = pkgs.gotools;
            };
            # Full lint suite matching the CI configuration
            golangci-lint = {
              enable  = true;
              package = pkgs.golangci-lint;
            };
          };
        };

        kubectl-redroid = pkgs.buildGoModule (commonAttrs // {
          pname = "kubectl-redroid";
          inherit version;
          subPackages = [ "cmd/kubectl-redroid" ];
          ldflags = [
            "-s" "-w"
            "-X github.com/isning/redroid-operator/cmd/kubectl-redroid/cmd.Version=${version}"
          ];
          meta = with pkgs.lib; {
            description = "kubectl plugin for managing RedroidInstance and RedroidTask resources";
            homepage    = "https://github.com/isning/redroid-operator";
            license     = licenses.asl20;
            mainProgram = "kubectl-redroid";
          };
        });

        redroid-operator = pkgs.buildGoModule (commonAttrs // {
          pname = "redroid-operator";
          inherit version;
          subPackages = [ "cmd" ];
          ldflags = [ "-s" "-w" ];
          meta = with pkgs.lib; {
            description = "Kubernetes operator for managing Redroid Android container instances";
            homepage    = "https://github.com/isning/redroid-operator";
            license     = licenses.asl20;
            mainProgram = "manager";
          };
        });
      in
      {
        # ── packages ───────────────────────────────────────────────────────────
        packages = {
          inherit kubectl-redroid redroid-operator;
          default = kubectl-redroid;
        };

        # ── apps ───────────────────────────────────────────────────────────────
        apps =
          let
            mkApp' = drv: name: flake-utils.lib.mkApp { inherit drv name; }
              // { meta.description = drv.meta.description; };
            appKubectl  = mkApp' kubectl-redroid  "kubectl-redroid";
            appOperator = mkApp' redroid-operator "manager";
          in {
            kubectl-redroid  = appKubectl;
            redroid-operator = appOperator;
            default          = appKubectl;
          };

        # ── devShell ───────────────────────────────────────────────────────────
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Go toolchain (1.22+ required; use latest stable from nixpkgs)
            go

            # Linting
            golangci-lint

            # Kubernetes helpers
            kubectl
            kustomize
            helm
            helm-docs

            # Docs tooling
            crane

            # Misc
            gnumake
            git
          ] ++ preCommitCheck.enabledPackages;

          shellHook = preCommitCheck.shellHook + ''
            export GOPATH="$HOME/go"
            export PATH="$GOPATH/bin:$PATH"
            echo "redroid-operator dev shell — Go $(go version | awk '{print $3}')"
          '';
        };

        # ── checks ─────────────────────────────────────────────────────────────
        checks = {
          pre-commit = preCommitCheck;
        };
      });
}

