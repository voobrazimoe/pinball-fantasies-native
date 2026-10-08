package io.github.voobrazimoe.pinballfantasies;

import java.io.*;
import java.nio.file.*;
import java.util.*;

public final class DataImportTest {
    static void check(boolean condition) { if (!condition) throw new AssertionError(); }
    interface Work { void run() throws IOException; }
    static void rejects(Work work) throws IOException {
        try { work.run(); } catch (IOException expected) { return; }
        throw new AssertionError("Expected rejection");
    }
    static final class Engine implements DataImport.Engine {
        boolean reject, rejectExisting; String rejection = "bounded shared-loader failure"; int validations, boots, stops;
        File liveState;
        public void validate(File data, File state) throws IOException {
            validations++;
            check(!state.equals(liveState));
            check(state.getName().startsWith("State.validation-"));
            check(data.getName().startsWith("Data.staging-") || data.getName().equals("Data"));
            Files.write(new File(state, "validation-only").toPath(), new byte[]{1});
            if (reject || (rejectExisting && data.getName().equals("Data"))) throw new IOException(rejection);
        }
        public void stop() { stops++; }
        public void bootstrap(File data, File state) { boots++; check(state.equals(liveState)); }
    }
    static class Source implements DataImport.Source {
        Map<String, String> names = new LinkedHashMap<>();
        boolean unreadable; int opened, closed;
        Source() {
            for (String name : DataImport.REQUIRED) names.put(name.toLowerCase(Locale.ROOT), name);
            names.put("pinball.cfg", "PINBALL.CFG");
            for (String name : Arrays.asList("PINBALL.EXE", "TABLE1.HI", "OTHER.MOD", "subfolder")) names.put(name, name);
        }
        public Map<String, String> entries() { return names; }
        public InputStream open(String id) throws IOException {
            check(DataImport.canonical(id) != null);
            if (unreadable) throw new IOException("Provider denied read");
            opened++;
            return new ByteArrayInputStream(new byte[]{42}) {
                public void close() throws IOException { closed++; super.close(); }
            };
        }
    }
    static void original(File data) throws IOException {
        Files.createDirectories(data.toPath()); Files.write(new File(data, "sentinel").toPath(), new byte[]{7});
    }
    static void preserved(File data, File state) throws IOException {
        check(Files.readAllBytes(new File(data, "sentinel").toPath())[0] == 7);
        check(Files.readAllBytes(new File(state, "settings").toPath())[0] == 9);
    }
    public static void main(String[] args) throws Exception {
        File root = Files.createTempDirectory("a2-noncommercial-").toFile();
        try {
            File storage = new File(root, "no_backup"), state = new File(root, "files/State");
            Files.createDirectories(state.toPath()); Files.write(new File(state, "settings").toPath(), new byte[]{9});
            Engine engine = new Engine(); engine.liveState = state;
            DataImport importer = new DataImport(storage, state, engine);
            importer.recover(); original(importer.data());
            importer.install(null); preserved(importer.data(), state); check(engine.validations == 0);
            Source missing = new Source(); missing.names.remove("table4.mod");
            rejects(() -> importer.install(missing)); preserved(importer.data(), state); check(missing.opened == 0);
            Source ambiguous = new Source(); ambiguous.names.put("INTRO.PRG", "other");
            rejects(() -> importer.install(ambiguous)); preserved(importer.data(), state);
            Source denied = new Source(); denied.unreadable = true;
            rejects(() -> importer.install(denied)); preserved(importer.data(), state);
            engine.reject = true; engine.rejection = "INTRO.PRG: unsupported data layout"; Source rejected = new Source();
            rejects(() -> importer.install(rejected)); preserved(importer.data(), state);
            check(rejected.opened == 12 && rejected.closed == 12); check(engine.stops == 0);
            engine.reject = false;
            DataImport failing = new DataImport(storage, state, engine, (from, to) -> {
                if (from.getName().startsWith("Data.staging-")) throw new IOException("injected adoption failure");
                Files.move(from.toPath(), to.toPath(), StandardCopyOption.ATOMIC_MOVE);
            });
            rejects(() -> failing.install(new Source())); preserved(importer.data(), state); check(engine.boots == 1);
            Source valid = new Source(); valid.names.remove("pinball.cfg");
            importer.install(valid); check(engine.boots == 2);
            check(importer.data().list().length == 11);
            for (String name : DataImport.REQUIRED) check(new File(importer.data(), name).isFile());
            check(!new File(state, "validation-only").exists());
            check(storage.list().length == 1);
            importer.install(new Source());
            check(importer.data().list().length == 12);
            check(new File(importer.data(), "PINBALL.CFG").isFile());
            // A selection without TABLE2-4.PRG is offered as the 10-minute demo.
            Source demo = new Source(); demo.names.remove("pinball.cfg");
            for (String name : Arrays.asList("table2.prg", "table2.mod", "table3.prg", "table3.mod", "table4.prg", "table4.mod"))
                demo.names.remove(name);
            check(DataImport.requiredFor(DataImport.select(demo.entries())).equals(DataImport.DEMO));
            importer.install(demo); check(importer.data().list().length == 6); // + PINBALL.EXE
            check(new File(importer.data(), "PINBALL.EXE").isFile());
            Source partial = new Source(); partial.names.remove("table2.prg");
            rejects(() -> importer.install(partial)); check(importer.data().list().length == 6);
            importer.install(new Source()); check(importer.data().list().length == 12);
            // Failed initial rename leaves Data at its original path.
            DataImport firstRenameFailure = new DataImport(storage, state, engine, (from, to) -> {
                throw new IOException("injected initial rename failure");
            });
            rejects(() -> firstRenameFailure.install(new Source()));
            check(importer.data().list().length == 12);
            // Process death between old -> previous and candidate -> Data recovers the old directory.
            DataImport.remove(importer.data()); original(new File(storage, "Data.previous"));
            Files.createDirectories(new File(storage, "Data.staging-stale").toPath());
            Files.createDirectories(new File(storage, "State.validation-stale").toPath());
            importer.recover(); preserved(importer.data(), state); check(storage.list().length == 1);
            // Even if rollback cannot rename, recovery retains the previous installation.
            DataImport failedRollback = new DataImport(storage, state, engine, (from, to) -> {
                if (from.getName().equals("Data")) Files.move(from.toPath(), to.toPath());
                else throw new IOException("injected adoption and rollback failures");
            });
            rejects(() -> failedRollback.install(new Source()));
            check(!importer.data().exists());
            check(new File(storage, "Data.previous/sentinel").exists());
            importer.recover(); preserved(importer.data(), state);
            // Cleanup must never follow a symlink out of private scratch.
            File outside = new File(root, "outside"); original(outside);
            Files.createSymbolicLink(new File(storage, "Data.staging-link").toPath(), outside.toPath());
            importer.recover(); check(new File(outside, "sentinel").exists());
            // Embedded bootstrap uses the same transaction; valid existing Data wins.
            Source embedded = new Source();
            int bootsBefore = engine.boots, stopsBefore = engine.stops;
            importer.bootstrap(embedded);
            preserved(importer.data(), state);
            check(embedded.opened == 0 && engine.boots == bootsBefore + 1);
            check(engine.stops == stopsBefore);
            engine.rejectExisting = true;
            Source failedEmbedded = new Source(); failedEmbedded.unreadable = true;
            rejects(() -> importer.bootstrap(failedEmbedded));
            preserved(importer.data(), state);
            check(engine.stops == stopsBefore);
            engine.reject = true;
            rejects(() -> importer.bootstrap(new Source()));
            preserved(importer.data(), state);
            engine.reject = false;
            importer.bootstrap(new Source());
            check(importer.data().list().length == 12);
            check(Files.readAllBytes(new File(state, "settings").toPath())[0] == 9);
            engine.rejectExisting = false;
            DataImport.remove(importer.data());
            importer.bootstrap((DataImport.Source)null); // ordinary asset-free shell
            check(!importer.data().exists());
            engine.reject = true;
            rejects(() -> importer.bootstrap(new Source()));
            check(!importer.data().exists() && storage.list().length == 0);
            engine.reject = false;
            Source initial = new Source(); initial.names.remove("pinball.cfg");
            importer.bootstrap(initial);
            check(importer.data().list().length == 11 && initial.closed == 11);
            check(Files.readAllBytes(new File(state, "settings").toPath())[0] == 9);
            check(storage.list().length == 1);
            System.out.println("PASS: filtering, optional CFG, cancellation, copy/validation/adoption failures, adoption, recovery, cleanup");
        } finally { DataImport.remove(root); }
    }
}
