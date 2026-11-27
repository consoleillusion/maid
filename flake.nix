{
  description = "maid - Metric Assisted AI Driven Development";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    # Beads (bd) is maid's task store. Its flake pins its own nixpkgs,
    # because bd needs a newer Go than maid does.
    beads.url = "github:gastownhall/beads";
  };

  outputs =
    { self, nixpkgs, beads }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system nixpkgs.legacyPackages.${system});
    in
    {
      packages = forAllSystems (
        system: pkgs: {
          default = self.packages.${system}.maid;
          maid = pkgs.buildGoModule {
            pname = "maid";
            version = "0.1.0-dev";
            src = self;
            subPackages = [ "cmd/maid" ];
            # maid uses only the Go standard library (ADR-0007).
            vendorHash = null;
            env.CGO_ENABLED = "0";
            # The tests need git, go and bd at run time; run them in the dev shell.
            doCheck = false;
            meta = {
              description = "MCP server that guides agents from design to accepted code";
              mainProgram = "maid";
              license = pkgs.lib.licenses.mit;
            };
          };
          bd = beads.packages.${system}.default;
        }
      );

      apps = forAllSystems (
        system: _: {
          default = {
            type = "app";
            program = "${self.packages.${system}.maid}/bin/maid";
          };
        }
      );

      devShells = forAllSystems (
        system: pkgs: {
          default = pkgs.mkShell {
            packages = [
              pkgs.go
              pkgs.gopls
              pkgs.gotools
              pkgs.golangci-lint
              pkgs.git
              pkgs.python3 # used by test/e2e and spikes/beads
              beads.packages.${system}.default
            ];
            shellHook = ''
              # `go install ./cmd/maid` puts maid in .bin/, which is on PATH here.
              export GOBIN="$PWD/.bin"
              export PATH="$GOBIN:$PATH"
              export MAID_BD="$(command -v bd)"
              # Print to stderr so `nix develop -c maid mcp` keeps stdout clean for MCP.
              echo "maid dev shell: $(go version | cut -d' ' -f3), bd $(bd version 2>/dev/null | head -n1)" >&2
              echo "build maid with: go install ./cmd/maid" >&2
            '';
          };
        }
      );
    };
}
