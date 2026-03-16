{
  description = "ransomware - A simple demonstration tool to simulate a ransomware attack";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
        version = self.shortRev or self.dirtyShortRev or "dev";
      in {
        packages = {
          ransomware = pkgs.buildGoModule {
            pname = "ransomware";
            inherit version;
            src = ./.;

            vendorHash = "sha256-ZL4qOSRh/MteZwvQyTJGAttHvIN2jr2qquELtg2l+QA=";

            ldflags = [
              "-s"
              "-w"
              "-X main.version=${version}"
              "-X main.commit=${self.rev or "dirty"}"
              "-X main.date=1970-01-01T00:00:00Z"
            ];

            meta = with pkgs.lib; {
              description = "A simple demonstration tool to simulate a ransomware attack";
              homepage = "https://github.com/marmos91/ransomware";
              license = licenses.mit;
              maintainers = [];
              mainProgram = "ransomware";
            };
          };

          default = self.packages.${system}.ransomware;
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            goreleaser
          ];
        };
      }
    );
}
