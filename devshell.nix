{ pkgs,
  fretboard
}:

pkgs.mkShell {
  inputsFrom = [ fretboard ];

  buildInputs = with pkgs; [
    valgrind
    ldb
    clang-tools
  ];
}
