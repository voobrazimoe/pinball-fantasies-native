#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
javac -d "$scratch" hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/DataImport.java hosts/android/tests/DataImportTest.java
java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.DataImportTest
