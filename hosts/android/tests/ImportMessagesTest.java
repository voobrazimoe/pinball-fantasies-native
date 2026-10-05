package io.github.voobrazimoe.pinballfantasies;

public final class ImportMessagesTest {
    static void check(boolean value) { if (!value) throw new AssertionError(); }
    public static void main(String[] args) {
        String failure = ImportMessages.failure("java.io.IOException: INTRO.PRG: unsupported data layout: [INTRO.PRG: unsupported layout dos-retail-linked-v1: missing FORM anchor]");
        check(failure.equals("This Pinball Fantasies DOS version is not supported yet (INTRO.PRG)."));
        check(ImportMessages.failure("java.io.IOException: TABLE2.MOD: truncated MOD patterns")
                .equals("The selected DOS data are incomplete or incompatible (TABLE2.MOD)."));
        check(ImportMessages.failure("Missing required file: INTRO.PRG")
                .equals("This folder is missing required Pinball Fantasies DOS files."));
        for (String detail : new String[]{null,"java.io.IOException: /secret/provider/path", "java.lang.IllegalStateException: [giant error]", "OTHER.PRG: unsupported layout"}) {
            String result = ImportMessages.failure(detail);
            check(!result.contains("java.") && !result.contains("secret") && !result.contains("[") && !result.contains("OTHER.PRG"));
        }
        System.out.println("PASS: concise import errors with allowlisted filename context");
    }
}
