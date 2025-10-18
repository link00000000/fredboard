{ pkgs,
  fretboard
}:

pkgs.mkShell {
  inputsFrom = [ fretboard ];

  buildInputs = with pkgs; [
    valgrind
    ldb
    clang-tools
    pkgs.vscode-extensions.ms-vscode.cpptools
  ];

  shellHook = ''
    export CPPDBG_PATH="${pkgs.vscode-extensions.ms-vscode.cpptools}/share/vscode/extensions/ms-vscode.cpptools/debugAdapters/bin/OpenDebugAD7"
  '';
}
