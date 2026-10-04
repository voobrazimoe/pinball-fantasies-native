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
    private boolean resumed, focused;
    private Controls controls;
    private boolean diagnostics;
    static native void nativeDiagnostics(boolean enabled);
    private void diagnostic(String message) {
        if (diagnostics) Log.i("PinballFantasies", message);
    }
    static native void nativeActive(long session, boolean resumed, boolean focused);
    static native void nativeInput(long session, int kind, int a, int b);
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
        if (event.endsWith("REJECTED")) Log.e("PinballFantasies", event + " " + message);
        else diagnostic(event + " " + message);
        runOnUiThread(() -> {
            if (!closed) {
                importButton.setEnabled(true);
                if (event.equals("A2_DATA_READY") || event.equals("A2_IMPORT_FINISHED"))
                    importButton.setVisibility(android.view.View.GONE);
                Toast.makeText(this, message, Toast.LENGTH_LONG).show();
            }
        });
    }
    private void requestImport() {
        if (!importButton.isEnabled()) return;
        importButton.setEnabled(false);
        diagnostic("A2_IMPORT_REQUESTED");
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
        diagnostics = getIntent().getBooleanExtra("PF_DIAGNOSTICS", false)
                || "1".equals(System.getenv("PF_DIAGNOSTICS"));
        nativeDiagnostics(diagnostics);
        super.onCreate(savedInstanceState);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            WindowManager.LayoutParams attributes = getWindow().getAttributes();
            attributes.layoutInDisplayCutoutMode =
                    WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES;
            getWindow().setAttributes(attributes);
        }
        applyImmersiveMode();
        session = nativeOpen();
        controls = new Controls((kind,a,b) -> nativeInput(session,kind,a,b));
        getOnBackPressedDispatcher().addCallback(this,new androidx.activity.OnBackPressedCallback(true) {
            @Override public void handleOnBackPressed() { controls.tap(1); }
        });
        addContentView(new ControlOverlay(this,controls), new FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,FrameLayout.LayoutParams.MATCH_PARENT));
        android.widget.LinearLayout menu = new android.widget.LinearLayout(this);
        menu.setOrientation(android.widget.LinearLayout.VERTICAL);
        android.widget.LinearLayout row = new android.widget.LinearLayout(this);
        menu.addView(row);
        String[] labels={"Enter","Esc","P","M","Y","N"};
        int[] codes={28,1,25,50,21,49};
        for (int i=0;i<labels.length;i++) addKey(row,labels[i],codes[i]);
        Button more = new Button(this); more.setText("Keys"); more.setTextSize(11);
        more.setAlpha(.65f); more.setFocusable(false);
        android.widget.HorizontalScrollView alphabet = new android.widget.HorizontalScrollView(this);
        android.widget.LinearLayout letters = new android.widget.LinearLayout(this);
        alphabet.addView(letters);
        for (int k=29;k<=54;k++) {
            final int code=Controls.make(k);
            Button letter=new Button(this); letter.setText(Character.toString((char)('A'+k-29)));
            letter.setAlpha(.65f); letter.setFocusable(false);
            letter.setOnClickListener(v -> controls.tap(code)); letters.addView(letter);
        }
        alphabet.setVisibility(android.view.View.GONE);
        more.setOnClickListener(v -> alphabet.setVisibility(
                alphabet.getVisibility()==android.view.View.GONE ? android.view.View.VISIBLE : android.view.View.GONE));
        row.addView(more,new android.widget.LinearLayout.LayoutParams(0,
                (int)(40*getResources().getDisplayMetrics().density),1));
        row = new android.widget.LinearLayout(this); menu.addView(row);
        for (int i=0;i<8;i++) addKey(row,"F"+(i+1),59+i);
        Button dataButton=new Button(this); dataButton.setText("Data"); dataButton.setTextSize(11);
        dataButton.setAlpha(.65f); dataButton.setFocusable(false);
        dataButton.setOnClickListener(v -> requestImport());
        row.addView(dataButton,new android.widget.LinearLayout.LayoutParams(0,
                (int)(40*getResources().getDisplayMetrics().density),1));
        menu.addView(alphabet);
        addContentView(menu,new FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.WRAP_CONTENT,Gravity.TOP));
        importer = new DataImport(getNoBackupFilesDir(), new File(getFilesDir(), "State"),
                new DataImport.Engine() {
                    public void validate(File data, File state) throws IOException { engineCall(0, data, state); }
                    public void adopted() { diagnostic("A2_DATA_ADOPTED"); }
                    public void stop() throws IOException { engineCall(2, null, null); }
                    public void bootstrap(File data, File state) throws IOException { engineCall(1, data, state); }
                });
        importButton = new Button(this);
        importButton.setText("Import DOS folder");
        importButton.setEnabled(false);
        importButton.setOnClickListener(view -> requestImport());
        FrameLayout.LayoutParams layout = new FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.WRAP_CONTENT, FrameLayout.LayoutParams.WRAP_CONTENT,
                Gravity.CENTER);
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
        resumed = true; syncInputs();
    }

    @Override
    public void onWindowFocusChanged(boolean hasFocus) {
        super.onWindowFocusChanged(hasFocus);
        focused = hasFocus; syncInputs();
        if (hasFocus) {
            applyImmersiveMode();
        }
    }

    private void addKey(android.widget.LinearLayout row,String text,int code) {
        Button button=new Button(this);
        button.setText(text); button.setTextSize(11); button.setMinWidth(0);
        button.setMinimumWidth(0); button.setPadding(0,0,0,0); button.setAlpha(.65f);
        button.setFocusable(false);
        button.setOnClickListener(v -> controls.tap(code));
        row.addView(button,new android.widget.LinearLayout.LayoutParams(0,
                (int)(40*getResources().getDisplayMetrics().density),1));
    }
    private void syncInputs() {
        if (controls==null) return;
        // Suspend under the native mutex before clearing Java ownership. Pending
        // spring/mouse edges are discarded by the engine, not fired on focus loss.
        nativeActive(session,resumed,focused);
        controls.enabled=resumed && focused;
        if (!controls.enabled) controls.clear();
    }
    @Override protected void onPause() {
        resumed=false; syncInputs(); super.onPause();
    }
    @Override public boolean dispatchKeyEvent(android.view.KeyEvent event) {
        if (controls!=null && (event.getAction()==android.view.KeyEvent.ACTION_DOWN ||
                event.getAction()==android.view.KeyEvent.ACTION_UP) &&
                controls.key(event.getKeyCode(),event.getAction()==android.view.KeyEvent.ACTION_DOWN,
                             event.getRepeatCount())) return true;
        return super.dispatchKeyEvent(event);
    }
    @Override public void onBackPressed() {
        if (controls!=null) controls.tap(1);
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
