{ self, ... }:
{ ... }:
{
  imports = [ ./services/fredboard ];
  nixpkgs.overlays = [ self.overlays.fredboard ];
}
