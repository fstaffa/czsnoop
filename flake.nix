{
  description = "A Nix-flake-based Go 1.22 development environment";

  inputs.nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/0.1.*.tar.gz";

  outputs = { self, nixpkgs }:
    let
      goVersion = 22; # Change this to update the whole stack
      overlays =
        [ (final: prev: { go = final."go_1_${toString goVersion}"; }) ];
      supportedSystems =
        [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forEachSupportedSystem = f:
        nixpkgs.lib.genAttrs supportedSystems
        (system: f { pkgs = import nixpkgs { inherit overlays system; }; });
    in {
      devShells = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            nixfmt
            # go (version is specified by overlay)
            go

            # goimports, godoc, etc.
            gotools

            # https://github.com/golangci/golangci-lint
            golangci-lint
          ];
        };
      });
      formatter = forEachSupportedSystem ({ pkgs }: pkgs.nixfmt);
      packages = forEachSupportedSystem ({ pkgs }: {
        default = pkgs.buildGoModule {
          pname = "czsnoop";
          version = "0.0.1";
          src = ./.;
          vendorHash = "sha256-dy0UtsakZoAJz1aLsJcAsa7SMyM/mVUdcVcu4969Ew8=";
        };
      });
    };
}
