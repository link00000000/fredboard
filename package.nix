{ lib
, stdenv
, cmake
, pkg-config
, dpp
}:

stdenv.mkDerivation {
  name = "fretboard";

  src = lib.sourceByRegex ./. [
    "^src.*"
    "^test.*"
    "CMakeLists.txt"
  ];

  nativeBuildInputs = [ cmake pkg-config ];
  buildInputs = [ dpp ];
  checkInputs = [];

  cmakeFlags = [];
  doCheck = true;
}
