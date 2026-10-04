package io.github.voobrazimoe.pinballfantasies;

import android.app.Activity;
import android.content.Intent;
import android.database.Cursor;
import android.net.Uri;
import android.provider.DocumentsContract;
import android.util.Log;
import android.view.Gravity;
import android.widget.Button;
import android.widget.FrameLayout;
import android.widget.Toast;
import java.io.*;
import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import android.os.Build;
import android.os.Bundle;
import android.view.WindowManager;

import androidx.core.view.WindowCompat;
import androidx.core.view.WindowInsetsCompat;
import androidx.core.view.WindowInsetsControllerCompat;

import com.google.androidgamesdk.GameActivity;

public final class PinballActivity extends GameActivity {
    static {
        System.loadLibrary("c++_shared");
        System.loadLibrary("pinball_android");
    }

    private static final int IMPORT_TREE = 2;
    private final ExecutorService worker = Executors.newSingleThreadExecutor();
    private DataImport importer;
    private long session;
    private Button importButton;
    private volatile boolean closed;
    static native long nativeOpen();
    static native void nativeClose(long session);
    static native String nativeEngine(long session, int operation, byte[] data, byte[] state);

    private void engineCall(int operation, File data, File state) throws IOException {
        String failure = nativeEngine(session, operation,
                data == null ? null : data.getAbsolutePath().getBytes(StandardCharsets.UTF_8),
                state == null ? null : state.getAbsolutePath().getBytes(StandardCharsets.UTF_8));
        if (failure != null) throw new IOException(failure);
    }
    private void status(String event, String message) {
        Log.i("PinballFantasies", event + " " + message);
        runOnUiThread(() -> {
            if (!closed) {
                importButton.setEnabled(true);
                Toast.makeText(this, message, Toast.LENGTH_LONG).show();
            }
        });
    }
    private void requestImport() {
        importButton.setEnabled(false);
        Log.i("PinballFantasies", "A2_IMPORT_REQUESTED");
        Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT_TREE);
        intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
        try { startActivityForResult(intent, IMPORT_TREE); }
        catch (RuntimeException failure) { status("A2_IMPORT_REJECTED", failure.toString()); }
    }
    @Override
    protected void onActivityResult(int request, int result, Intent intent) {
        super.onActivityResult(request, result, intent);
        if (request != IMPORT_TREE) return;
        if (result != Activity.RESULT_OK || intent == null || intent.getData() == null) {
            status("A2_IMPORT_CANCELLED", "Import cancelled");
            return;
        }
        Uri tree = intent.getData();
        worker.execute(() -> {
            try {
                importer.install(new DataImport.Source() {
                    public Map<String, String> entries() throws IOException {
                        Uri children = DocumentsContract.buildChildDocumentsUriUsingTree(tree,
                                DocumentsContract.getTreeDocumentId(tree));
                        Map<String, String> entries = new LinkedHashMap<>();
                        String[] projection = {DocumentsContract.Document.COLUMN_DOCUMENT_ID,
                                DocumentsContract.Document.COLUMN_DISPLAY_NAME,
                                DocumentsContract.Document.COLUMN_MIME_TYPE};
                        try (Cursor cursor = getContentResolver().query(children, projection, null, null, null)) {
                            if (cursor == null) throw new IOException("Provider cannot list selected folder");
                            while (cursor.moveToNext()) {
                                String name = cursor.getString(1);
                                if (DataImport.canonical(name) == null ||
                                        DocumentsContract.Document.MIME_TYPE_DIR.equals(cursor.getString(2))) continue;
                                if (entries.put(name, cursor.getString(0)) != null)
                                    throw new IOException("Ambiguous filename: " + name);
                            }
                        }
                        return entries;
                    }
                    public InputStream open(String id) throws IOException {
                        return getContentResolver().openInputStream(
                                DocumentsContract.buildDocumentUriUsingTree(tree, id));
                    }
                });
                status("A2_IMPORT_FINISHED", "Validated Data adopted; real engine ready");
            } catch (IOException | RuntimeException failure) {
                status("A2_IMPORT_REJECTED", failure.toString());
            }
        });
    }
    @Override
    protected void onDestroy() {
        closed = true;
        worker.shutdownNow();
        nativeClose(session);
        super.onDestroy();
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            WindowManager.LayoutParams attributes = getWindow().getAttributes();
            attributes.layoutInDisplayCutoutMode =
                    WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES;
            getWindow().setAttributes(attributes);
        }
        applyImmersiveMode();
        session = nativeOpen();
        importer = new DataImport(getNoBackupFilesDir(), new File(getFilesDir(), "State"),
                new DataImport.Engine() {
                    public void validate(File data, File state) throws IOException { engineCall(0, data, state); }
                    public void adopted() { Log.i("PinballFantasies", "A2_DATA_ADOPTED"); }
                    public void stop() throws IOException { engineCall(2, null, null); }
                    public void bootstrap(File data, File state) throws IOException { engineCall(1, data, state); }
                });
        importButton = new Button(this);
        importButton.setText("Import DOS folder");
        importButton.setEnabled(false);
        importButton.setOnClickListener(view -> requestImport());
        FrameLayout.LayoutParams layout = new FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.WRAP_CONTENT, FrameLayout.LayoutParams.WRAP_CONTENT,
                Gravity.TOP | Gravity.CENTER_HORIZONTAL);
        addContentView(importButton, layout);
        worker.execute(() -> {
            try {
                importer.recover();
                importer.bootstrap();
                status(importer.data().exists() ? "A2_DATA_READY" : "A2_SHELL_NO_DATA",
                        importer.data().exists() ? "Real engine ready" : "Select your original DOS folder to import");
            } catch (IOException | RuntimeException failure) {
                status("A2_BOOTSTRAP_REJECTED", failure.toString());
            }
        });
    }

    @Override
    protected void onResume() {
        super.onResume();
        applyImmersiveMode();
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        if (hasFocus) {
            applyImmersiveMode();
        }
    }

    private void applyImmersiveMode() {
        WindowCompat.setDecorFitsSystemWindows(getWindow(), false);
        WindowInsetsControllerCompat controller =
                WindowCompat.getInsetsController(getWindow(), getWindow().getDecorView());
        controller.setSystemBarsBehavior(
                WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
        controller.hide(WindowInsetsCompat.Type.systemBars());
    }
}
