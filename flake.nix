{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    nix-vscode-extensions.url = "github:nix-community/nix-vscode-extensions";
  };
  outputs = {
    self,
    nixpkgs,
    flake-utils,
    nix-vscode-extensions,
  }:
    flake-utils.lib.eachDefaultSystem
    (
      system: let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
          config = {
            allowUnfree = true;
          };
        };
        extensions = nix-vscode-extensions.extensions.${system};
        inherit (pkgs) vscode-with-extensions vscodium;

        packages.default = vscode-with-extensions.override {
          vscode = vscodium;
          vscodeExtensions = with extensions; [
            vscode-marketplace.golang.go
            vscode-marketplace-release.github.copilot
            vscode-marketplace-release.github.copilot-chat
            open-vsx.catppuccin.catppuccin-vsc
            open-vsx.jnoortheen.nix-ide
          ];
        };
      in
        with pkgs; {
          devShells.default = mkShell {
            # 👇 we can just use `rustToolchain` here:
            buildInports = [
              go

              gopls
              delve

              # goimports, godoc, etc.
              gotools

              # https://github.com/golangci/golangci-lint
              golangci-lint
            ];
            packages = [
              # Development Tools
              go
              packages.default
              nil
              wireshark
              tcpdump
              tmux

              # Development Tools
              hotspot
              pprof
              gperftools
              graphviz
              perf_data_converter

              linuxKernel.packages.linux_6_11.perf

              jetbrains.goland
              # goimports, godoc, etc.
              gotools

              # https://github.com/golangci/golangci-lint
              golangci-lint
              gopls
              delve
              # Development time dependencies
              gtest
            ];
          };
        }
    );
}
