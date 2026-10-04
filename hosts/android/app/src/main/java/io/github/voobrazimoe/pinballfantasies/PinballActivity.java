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
    private ControlMenu menu;
    private ControlOverlay overlay;
    private final int[] viewport=new int[6];
    static native void nativeViewport(int[] bounds);
    private final android.os.Handler uiHandler=new android.os.Handler(android.os.Looper.getMainLooper());
    private final Runnable refreshUi=new Runnable() {
        public void run() {
            if (closed) return;
            nativeViewport(viewport);
            overlay.viewport(viewport);
            long state=nativeState(session);
            menu.snapshot(state<0 ? -1 : (int)(state&255),
                    state<0 ? 0 : (int)((state>>8)&255), state<0 ? 0 : (int)((state>>16)&255));
            uiHandler.postDelayed(this,100);
        }
    };
    static native long nativeState(long session);
    static native void nativeDetach(long session);
    private boolean diagnostics;
    private AndroidAudio audio;
    static native boolean nativeHasEngine(long session);
    static native void nativeAudio(long session, boolean granted, boolean route);
    private void syncAudio() {
        if (audio != null) audio.eligible(!closed && resumed && focused && nativeHasEngine(session));
    }
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
        if (operation == 2) runOnUiThread(() -> { if (!closed) audio.eligible(false); });
    }
    private void status(String event, String message) {
        if (event.endsWith("REJECTED")) Log.e("PinballFantasies", event + " " + message);
        else diagnostic(event + " " + message);
        runOnUiThread(() -> {
            if (!closed) {
                syncAudio();
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
        if (audio != null) audio.close();
        closed = true;
        worker.shutdownNow();
        uiHandler.removeCallbacks(refreshUi);
        if (isChangingConfigurations()) nativeDetach(session); else nativeClose(session);
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
        audio = new AndroidAudio(this, (enabled, route) -> nativeAudio(session, enabled, route), this::diagnostic);
        android.os.Handler inputHandler = new android.os.Handler(getMainLooper());
        controls = new Controls((kind,a,b) -> nativeInput(session,kind,a,b),
                (milliseconds,release) -> inputHandler.postDelayed(release,milliseconds));
        getOnBackPressedDispatcher().addCallback(this,new androidx.activity.OnBackPressedCallback(true) {
            @Override public void handleOnBackPressed() { if (!menu.dismiss()) controls.tap(1); }
        });
        FrameLayout interaction = new FrameLayout(this);
        overlay = new ControlOverlay(this,controls);
        interaction.addView(overlay,new FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.MATCH_PARENT,FrameLayout.LayoutParams.MATCH_PARENT));
        menu = new ControlMenu(this,controls);
        interaction.addView(menu,new FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.MATCH_PARENT));
        overlay.state=menu.state;
        controls.gameplay=false;
        uiHandler.post(refreshUi);
        // Keep framebuffer edge-to-edge. Only interactive UI uses these insets.
        final androidx.core.graphics.Insets[] safe = {androidx.core.graphics.Insets.NONE};
        Runnable geometry = () -> overlay.safeArea(safe[0].left,safe[0].top,
                safe[0].right,safe[0].bottom);
        interaction.addOnLayoutChangeListener((v,l,t,r,b,ol,ot,or,ob)->geometry.run());
        androidx.core.view.ViewCompat.setOnApplyWindowInsetsListener(interaction,(v,insets)->{
            safe[0]=ControlMenu.interactiveInsets(insets);
            menu.safeInsets(safe[0].left,safe[0].top,safe[0].right,safe[0].bottom);
            if (importButton!=null) {
                FrameLayout.LayoutParams lp=(FrameLayout.LayoutParams)importButton.getLayoutParams();
                lp.setMargins(safe[0].left,safe[0].top,safe[0].right,safe[0].bottom);
                importButton.setLayoutParams(lp);
            }
            geometry.run(); return insets;
        });
        addContentView(interaction,new FrameLayout.LayoutParams(FrameLayout.LayoutParams.MATCH_PARENT,
                FrameLayout.LayoutParams.MATCH_PARENT));
        androidx.core.view.ViewCompat.requestApplyInsets(interaction);
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
        interaction.addView(importButton, layout);
        androidx.core.view.ViewCompat.requestApplyInsets(interaction);
        worker.execute(() -> {
            try {
                importer.recover();
                String[] bundled = getAssets().list("personal-data");
                DataImport.Source embedded = bundled == null || bundled.length == 0 ? null : new DataImport.Source() {
                    public Map<String, String> entries() {
                        Map<String, String> entries = new LinkedHashMap<>();
                        for (String name : bundled) entries.put(name, "personal-data/" + name);
                        return entries;
                    }
                    public InputStream open(String id) throws IOException { return getAssets().open(id); }
                };
                importer.bootstrap(embedded);
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

    private void syncInputs() {
        if (controls==null) return;
        // Suspend under the native mutex before clearing Java ownership. Pending
        // spring/mouse edges are discarded by the engine, not fired on focus loss.
        nativeActive(session,resumed,focused);
        syncAudio();
        controls.enabled=resumed && focused;
        if (!controls.enabled) controls.clear();
    }
    @Override protected void onPause() {
        resumed=false; syncInputs(); super.onPause();
    }
    @Override public boolean dispatchKeyEvent(android.view.KeyEvent event) {
        // Consume only the make that dismisses our transient UI. Its matching
        // up has no Controls owner, so it cannot leak an Escape into the engine.
        if (menu!=null && event.getKeyCode()==android.view.KeyEvent.KEYCODE_BACK &&
                event.getAction()==android.view.KeyEvent.ACTION_DOWN && menu.dismiss()) return true;
        if (controls!=null && (event.getAction()==android.view.KeyEvent.ACTION_DOWN ||
                event.getAction()==android.view.KeyEvent.ACTION_UP) &&
                controls.key(event.getKeyCode(),event.getAction()==android.view.KeyEvent.ACTION_DOWN,
                             event.getRepeatCount())) return true;
        return super.dispatchKeyEvent(event);
    }
    @Override public void onBackPressed() {
        if (controls!=null && !menu.dismiss()) controls.tap(1);
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
