{ self, ... }:
final: prev: {
  fredboard = self.packages.${final.system}.fredboard;
}
