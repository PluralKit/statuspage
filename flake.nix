{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };
  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        backend = pkgs.buildGoModule {
          pname = "pluralkit-status-backend";
          version = self.shortRev or "dirty";
          src = ./backend;
          vendorHash = "sha256-Hls0A3Bq9BlAw+nknihmkrK+taQLhwzdMnpn9wwP7PQ=";
        };

        frontend = pkgs.buildNpmPackage {
          pname = "pluralkit-status-frontend";
          version = self.shortRev or "dirty";
          src = ./frontend;

          npmDepsHash = "sha256-5UfeBqKUgls4gT6eL384tsEEuByZT5gO3xpymDo+K/o=";
          npmBuildCommand = "run build";
          installPhase = ''
            runHook preInstall
            cp -r build $out
            runHook postInstall
          '';
        };
      in
      with pkgs;
      {
        packages = {
          pluralkit-status-backend = backend;
          pluralkit-status-frontend = frontend;
        };
        devShells.default = mkShell {
          buildInputs = [
            pkgs.gnumake
            pkgs.go
            pkgs.nodejs_24
            pkgs.golangci-lint
          ];
        };
      }
    );
}
