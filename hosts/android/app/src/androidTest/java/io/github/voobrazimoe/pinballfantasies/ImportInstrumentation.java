package io.github.voobrazimoe.pinballfantasies;

import android.app.Instrumentation;
import android.app.Activity;
import android.os.Bundle;
import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.*;

/** Runs against the packaged Go library. All candidate bytes are invented and invalid. */
public final class ImportInstrumentation extends Instrumentation {
    @Override public void onCreate(Bundle arguments) { super.onCreate(arguments); start(); }
    static void check(boolean value) { if (!value) throw new AssertionError(); }
    @Override public void onStart() {
        Bundle result = new Bundle();
        File root = null;
        long session = 0;
        try {
            root = Files.createTempDirectory(getTargetContext().getNoBackupFilesDir().toPath(), "a2-test-").toFile();
            File data = new File(root, "Data"), state = new File(root, "State.validation-test");
            Files.createDirectories(data.toPath()); Files.createDirectories(state.toPath());
            session = PinballActivity.nativeOpen();
            String missing = PinballActivity.nativeEngine(session, 0,
                    data.getAbsolutePath().getBytes(StandardCharsets.UTF_8),
                    state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
            check(missing != null && !missing.isEmpty() && missing.length() < 1024);
            for (String name : DataImport.REQUIRED) Files.write(new File(data, name).toPath(), new byte[]{1,2,3});
            String malformed = PinballActivity.nativeEngine(session, 0,
                    data.getAbsolutePath().getBytes(StandardCharsets.UTF_8),
                    state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
            check(malformed != null && !malformed.isEmpty() && malformed.length() < 1024);
            String uri = PinballActivity.nativeEngine(session, 0, "content://test".getBytes(StandardCharsets.UTF_8),
                    state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
            check(uri != null && uri.contains("filesystem"));
            result.putString("stream", "PASS: packaged ABI 1 JNI missing/malformed rejection, bounded error, URI guard\n");
            finish(Activity.RESULT_OK, result);
        } catch (Throwable failure) {
            result.putString("stream", "FAIL: " + failure + "\n");
            finish(Activity.RESULT_CANCELED, result);
        } finally {
            PinballActivity.nativeClose(session);
            if (root != null) try { DataImport.remove(root); } catch (IOException ignored) { }
        }
    }
}
