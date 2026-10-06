#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
scratch=$(mktemp -d)
trap 'rm -rf "$scratch"' EXIT HUP INT TERM
javac -d "$scratch" hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/DataImport.java hosts/android/tests/DataImportTest.java hosts/android/tests/DataImportPrivateTest.java hosts/android/app/src/main/java/io/github/voobrazimoe/pinballfantasies/ImportMessages.java hosts/android/tests/ImportMessagesTest.java
java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.DataImportTest
java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.ImportMessagesTest

if [ -n "${PF_RUNTIME_DATA:-}" ] && [ -n "${PF_POWERPACK_DATA:-}" ]; then
    ./tools/go.sh build -o "$scratch/runtime-validator" ./cmd/personalvalidate
    java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.DataImportPrivateTest "$scratch/runtime-validator"
else
    java -cp "$scratch" io.github.voobrazimoe.pinballfantasies.DataImportPrivateTest
fi
