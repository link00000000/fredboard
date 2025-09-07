{ pkgs, ... }: pkgs.buildGoApplication {
  pname = "fredboard";
  version = "dev";

  src = ../..;
  modules = ../../gomod2nix.toml;

  subPackages = [
    "cmd/fredboard"
  ];

  meta = {
    description = "A music player bot for Discord";
    homepage = "https://github.com/link00000000/fredboard";
    license = pkgs.lib.licenses.mit;
    maintainers = with pkgs.lib.maintainers; [ link00000000 ];
  };
}
