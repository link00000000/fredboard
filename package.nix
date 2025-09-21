{ lib
, stdenv
, cmake
}:

stdenv.mkDerivation {
  name = "fretboard";

  src = lib.sourceByRegex ./. [
    "^src.*"
    "^test.*"
    "CMakeLists.txt"
  ];

  nativeBuildInputs = [ cmake ];
  buildInputs = [];
  checkInputs = [];

  cmakeFlags = [];
  doCheck = true;
}
