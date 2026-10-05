package io.github.voobrazimoe.pinballfantasies;

import java.util.Locale;
import java.util.regex.*;

/** Allowlisted filename context only; parser details stay in diagnostic logging. */
final class ImportMessages {
    private static final Pattern FILE = Pattern.compile("(?i)(?<![A-Z0-9_])(INTRO\\.(?:PRG|MOD)|MOD2\\.MOD|TABLE[1-4]\\.(?:PRG|MOD)|PINBALL\\.CFG)(?![A-Z0-9_])");
    static String failure(String detail) {
        if (detail == null) detail = "";
        String lower = detail.toLowerCase(Locale.ROOT);
        if (lower.contains("missing required file"))
            return "This folder is missing required Pinball Fantasies DOS files.";
        Matcher match = FILE.matcher(detail);
        String file = match.find() ? match.group().toUpperCase(Locale.ROOT) : null;
        if (file != null && file.endsWith(".PRG") &&
                (lower.contains("unsupported") || lower.contains("incompatible")))
            return "This Pinball Fantasies DOS version is not supported yet (" + file + ").";
        if (file != null)
            return "The selected DOS data are incomplete or incompatible (" + file + ").";
        return "Could not import this folder. Check the DOS files and try again.";
    }
}
