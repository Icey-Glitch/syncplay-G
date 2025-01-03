{
  pkgs ? (
    let
      inherit (builtins) fetchTree fromJSON readFile;
      inherit ((fromJSON (readFile ./flake.lock)).nodes) nixpkgs gomod2nix;
    in
      import (fetchTree nixpkgs.locked) {
        overlays = [
          (import "${fetchTree gomod2nix.locked}/overlay.nix")
        ];
      }
  ),
  mkGoEnv ? pkgs.mkGoEnv,
  gomod2nix ? pkgs.gomod2nix,
  vscode-with-extensions ? pkgs.vscode-with-extensions,
  vscodium ? pkgs.vscodium,
  extensions,
}: let
  goEnv = mkGoEnv {pwd = ./src/.;};

  vscodiumNew = vscode-with-extensions.override {
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
  pkgs.mkShell {
    nativebuildInputs = with pkgs; [
      cjson
      gcc
      clang
    ];
    packages = with pkgs; [
      goEnv
      gomod2nix

      vscodiumNew

      pprof
      valgrind
      #perf_data_converter
      graphviz

      jetbrains.goland

      nil
      tmux
      tcpdump
      wireshark

      golangci-lint
      go
      go-tools
      gotools
      go-task
      delve
      golint
      gopls
      linuxKernel.packages.linux_6_11.perf
      delve

      clang
      gcc
      cjson
    ];

    # name the shell
    name = "golang-shell";
  }
