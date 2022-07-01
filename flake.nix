{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/master";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachSystem [
      "x86_64-linux"
      "aarch64-linux"
    ]
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [
              (final: prev: {
                go = prev.go_1_18;
                buildGoModule = prev.buildGo118Module;
              })
            ];
          };
        in
        rec {
          packages = flake-utils.lib.flattenTree rec {
            phosgraphe = let lib = pkgs.lib; in
              pkgs.buildGoModule {
                name = "phosgraphe";
                version = "0.0.1";

                nativeBuildInputs = with pkgs; [
                  pkg-config
                  ghostscript
                ];

                buildInputs = with pkgs; [
                  imagemagick
                ];

                vendorSha256 = "sha256-JtMYGE4j5R1NFpYlPC+fT6G16rN09phKMhBii8D9zrQ=";
                src = ./.;
                meta = {
                  license = lib.licenses.mit;
                  maintainers = [ "DAlperin" ];
                  platforms = lib.platforms.linux;
                };
              };
             default = phosgraphe;
          };

#          packages.default = packages.phosgraphe;
          defaultApp = packages.phosgraphe;

          checks = {
            format =
              pkgs.runCommand "check-format"
                {
                  buildInputs = with pkgs; [ go_1_18 ];
                } ''
                if [ "$(${pkgs.go_1_18}/bin/gofmt -s -l ${./.} | wc -l)" -gt 0 ]; then exit 1; fi
                touch $out # it worked!
              '';
          };

          devShells.default = pkgs.mkShell {
            nativeBuildInputs = [ pkgs.bashInteractive ];
            buildInputs = with pkgs; [
              go_1_18
              imagemagick
              pkg-config
              libpng
              ghostscript
            ];
          };
        });
}
