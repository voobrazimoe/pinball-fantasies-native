package io.github.voobrazimoe.pinballfantasies;

import java.io.*;
import java.nio.file.*;
import java.util.*;

/** Filesystem transaction shared by SAF and asset-free host tests. No data parser. */
final class DataImport {
    static final List<String> REQUIRED = Collections.unmodifiableList(Arrays.asList(
            "INTRO.PRG", "INTRO.MOD", "MOD2.MOD", "TABLE1.PRG", "TABLE1.MOD",
            "TABLE2.PRG", "TABLE2.MOD", "TABLE3.PRG", "TABLE3.MOD", "TABLE4.PRG", "TABLE4.MOD"));
    // Official 10-minute Party Land demo runtime roles; the engine decides.
    static final List<String> DEMO = Collections.unmodifiableList(Arrays.asList(
            "INTRO.PRG", "INTRO.MOD", "MOD2.MOD", "TABLE1.PRG", "TABLE1.MOD"));
    /** A selection without any TABLE2-4.PRG is offered as the demo. */
    static List<String> requiredFor(Map<String, String> selected) {
        for (String name : Arrays.asList("TABLE2.PRG", "TABLE3.PRG", "TABLE4.PRG"))
            if (selected.containsKey(name)) return REQUIRED;
        return DEMO;
    }
    interface Source {
        Map<String, String> entries() throws IOException;
        InputStream open(String id) throws IOException;
    }
    interface Engine {
        void validate(File data, File state) throws IOException;
        default void adopted() {}
        void stop() throws IOException;
        void bootstrap(File data, File state) throws IOException;
    }
    interface Rename { void move(File from, File to) throws IOException; }
    private final File root, state;
    private final Engine engine;
    private final Rename rename;
    DataImport(File root, File state, Engine engine) {
        this(root, state, engine, (from, to) -> Files.move(from.toPath(), to.toPath(), StandardCopyOption.ATOMIC_MOVE));
    }
    DataImport(File root, File state, Engine engine, Rename rename) {
        this.root = root; this.state = state; this.engine = engine; this.rename = rename;
    }
    static String canonical(String name) {
        if (name == null) return null;
        String upper = name.toUpperCase(Locale.ROOT);
        return REQUIRED.contains(upper) || "PINBALL.CFG".equals(upper) || "PINBALL.EXE".equals(upper) ? upper : null;
    }
    static Map<String, String> select(Map<String, String> entries) throws IOException {
        Map<String, String> selected = new LinkedHashMap<>();
        for (Map.Entry<String, String> entry : entries.entrySet()) {
            String name = canonical(entry.getKey());
            if (name != null && selected.put(name, entry.getValue()) != null)
                throw new IOException("Ambiguous filename: " + name);
        }
        for (String name : requiredFor(selected))
            if (!selected.containsKey(name)) throw new IOException("Missing required file: " + name);
        // The demo launcher only supplies the closing text; full installations ignore it.
        if (requiredFor(selected) != DEMO) selected.remove("PINBALL.EXE");
        return selected;
    }
    File data() { return new File(root, "Data"); }
    // Called once before any worker import. Backup is retained across an interrupted rename.
    void recover() throws IOException {
        Files.createDirectories(root.toPath());
        File previous = new File(root, "Data.previous");
        if (previous.exists() && !data().exists()) rename.move(previous, data());
        if (previous.exists() && data().exists()) remove(previous);
        File[] children = root.listFiles();
        if (children == null) throw new IOException("Cannot list private storage");
        for (File child : children)
            if (child.getName().startsWith("Data.staging-") || child.getName().startsWith("State.validation-")) remove(child);
    }
    void bootstrap() throws IOException {
        if (data().exists()) {
            Files.createDirectories(state.toPath());
            engine.bootstrap(data(), state);
        }
    }
    // Validate existing Data with disposable State before touching live player state.
    // Embedded and SAF sources share install(), including validation and rollback.
    void bootstrap(Source embedded) throws IOException {
        IOException invalid = null;
        if (data().exists()) {
            File validation = Files.createTempDirectory(root.toPath(), "State.validation-").toFile();
            try { engine.validate(data(), validation); }
            catch (IOException failure) { invalid = failure; }
            finally { remove(validation); }
            if (invalid == null) { bootstrap(); return; }
        }
        if (embedded != null) {
            try { install(embedded); }
            catch (IOException failure) {
                if (invalid != null) failure.addSuppressed(invalid);
                throw failure;
            }
        } else if (invalid != null) throw invalid;
    }
    void install(Source source) throws IOException {
        // null represents picker cancellation and makes no filesystem changes.
        if (source == null) return;
        Map<String, String> selected = select(source.entries());
        File staging = Files.createTempDirectory(root.toPath(), "Data.staging-").toFile();
        File validation = null;
        try {
            for (Map.Entry<String, String> entry : selected.entrySet()) {
                try (InputStream input = source.open(entry.getValue());
                     FileOutputStream output = new FileOutputStream(new File(staging, entry.getKey()))) {
                    if (input == null) throw new IOException("Unreadable file: " + entry.getKey());
                    byte[] buffer = new byte[65536];
                    int count;
                    while ((count = input.read(buffer)) != -1) output.write(buffer, 0, count);
                    output.getFD().sync();
                }
            }
            validation = Files.createTempDirectory(root.toPath(), "State.validation-").toFile();
            engine.validate(staging, validation);
            remove(validation); validation = null;
            File previous = new File(root, "Data.previous");
            // Never overwrite an unresolved backup.
            if (previous.exists()) throw new IOException("Previous installation needs recovery");
            engine.stop();
            boolean backedUp = false;
            try {
                if (data().exists()) { rename.move(data(), previous); backedUp = true; }
                try { rename.move(staging, data()); }
                catch (IOException failure) {
                    if (backedUp) rename.move(previous, data());
                    throw failure;
                }
            } catch (IOException failure) {
                try { bootstrap(); } catch (IOException boot) { failure.addSuppressed(boot); }
                throw failure;
            }
            // Commit is the staging -> Data rename; backup cleanup is best-effort after commit.
            try { remove(previous); } catch (IOException ignored) { /* recover on relaunch */ }
            engine.adopted();
            bootstrap();
        } finally {
            remove(staging);
            if (validation != null) remove(validation);
        }
    }
    static void remove(File file) throws IOException {
        if (!Files.exists(file.toPath(), LinkOption.NOFOLLOW_LINKS)) return;
        if (Files.isDirectory(file.toPath(), LinkOption.NOFOLLOW_LINKS)) {
            File[] children = file.listFiles();
            if (children == null) throw new IOException("Cannot clean private temporary directory");
            for (File child : children) remove(child);
        }
        Files.delete(file.toPath());
    }
}
