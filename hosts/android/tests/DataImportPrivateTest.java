package io.github.voobrazimoe.pinballfantasies;

import java.io.*;
import java.nio.file.*;
import java.util.*;

/** Optional private A/B/D integration: the actual filesystem transaction calls the
 * real shared runtime validator. Provider streams stand in for SAF; no fixtures.
 */
public final class DataImportPrivateTest {
    static void check(boolean ok) { if (!ok) throw new AssertionError(); }
    static final class Source implements DataImport.Source {
        final File directory;
        final Map<String, File> replacements = new HashMap<>();
        Source(String path) { directory = new File(path); }
        public Map<String, String> entries() {
            Map<String, String> out = new LinkedHashMap<>();
            for (String name : DataImport.REQUIRED)
                if (new File(directory, name).isFile()) out.put(name.toLowerCase(Locale.ROOT), name);
            return out;
        }
        public InputStream open(String id) throws IOException {
            return new FileInputStream(replacements.getOrDefault(id, new File(directory, id)));
        }
    }
    static final class Engine implements DataImport.Engine {
        final String validator;
        int validations, boots, stops;
        Engine(String path) { validator = path; }
        void load(File data) throws IOException {
            Process process = new ProcessBuilder(validator, "--data-dir", data.getAbsolutePath()).redirectErrorStream(true).start();
            byte[] output = process.getInputStream().readAllBytes();
            try {
                String message = new String(output, java.nio.charset.StandardCharsets.UTF_8);
                if (process.waitFor() != 0) throw new IOException("Shared runtime rejected input: " +
                        message.substring(0, Math.min(message.length(), 512)));
            } catch (InterruptedException e) { Thread.currentThread().interrupt(); throw new IOException(e); }
        }
        public void validate(File data, File state) throws IOException { validations++; load(data); }
        public void stop() { stops++; }
        public void bootstrap(File data, File state) throws IOException { boots++; load(data); }
    }
    static void identical(File live, Source expected) throws IOException {
        for (String name : DataImport.REQUIRED)
            check(Arrays.equals(Files.readAllBytes(new File(live, name).toPath()), Files.readAllBytes(new File(expected.directory, name).toPath())));
    }
    static void rejects(DataImport importer, DataImport.Source source) throws IOException {
        try { importer.install(source); } catch (IOException expected) {
            check(expected.getMessage().length() < 1024); return;
        }
        throw new AssertionError("Invalid installation accepted");
    }
    public static void main(String[] args) throws Exception {
        String a = System.getenv("PF_RUNTIME_DATA"), b = System.getenv("PF_POWERPACK_DATA");
        if (a == null || b == null || args.length != 1) {
            System.out.println("SKIP private Android A/B transaction: supply PF_RUNTIME_DATA and PF_POWERPACK_DATA"); return;
        }
        File root = Files.createTempDirectory("pf-private-import-").toFile();
        try {
            File storage = new File(root, "no_backup"), state = new File(root, "State");
            Engine engine = new Engine(args[0]);
            DataImport importer = new DataImport(storage, state, engine);
            importer.recover();
            Source canonical = new Source(a), powerpack = new Source(b);
            importer.install(canonical); identical(importer.data(), canonical);
            importer.install(powerpack); identical(importer.data(), powerpack);
            check(storage.list().length == 1 && importer.data().list().length == 11);
            importer.bootstrap((DataImport.Source)null); identical(importer.data(), powerpack);
            int stops = engine.stops;
            for (int mask = 1; mask < 7; mask++) {
                Source hybrid = new Source(a);
                String[] names = {"INTRO.PRG", "TABLE1.PRG", "TABLE2.PRG"};
                for (int i = 0; i < names.length; i++)
                    if ((mask & (1 << i)) != 0) hybrid.replacements.put(names[i], new File(b, names[i]));
                rejects(importer, hybrid); identical(importer.data(), powerpack);
            }
            File malformed = new File(root, "malformed"); Files.write(malformed.toPath(), new byte[]{0});
            Source bad = new Source(b); bad.replacements.put("INTRO.PRG", malformed);
            rejects(importer, bad); identical(importer.data(), powerpack);
            DataImport.Source incomplete = new DataImport.Source() {
                public Map<String, String> entries() { Map<String, String> out = powerpack.entries(); out.remove("table4.mod"); return out; }
                public InputStream open(String id) throws IOException { return powerpack.open(id); }
            };
            rejects(importer, incomplete); identical(importer.data(), powerpack); check(engine.stops == stops);
            DataImport failing = new DataImport(storage, state, engine, (from, to) -> {
                if (from.getName().startsWith("Data.staging-")) throw new IOException("Injected adoption failure");
                Files.move(from.toPath(), to.toPath(), StandardCopyOption.ATOMIC_MOVE);
            });
            rejects(failing, canonical); identical(importer.data(), powerpack);
            // Interrupted adoption before commit: restore the accepted B backup.
            Files.move(importer.data().toPath(), new File(storage, "Data.previous").toPath());
            Files.createDirectory(new File(storage, "Data.staging-interrupted").toPath());
            Files.createDirectory(new File(storage, "State.validation-interrupted").toPath());
            importer.recover(); importer.bootstrap((DataImport.Source)null);
            identical(importer.data(), powerpack); check(storage.list().length == 1);
            // Interrupted cleanup after commit: live B wins and stale backup is removed.
            Files.createDirectory(new File(storage, "Data.previous").toPath());
            importer.recover(); importer.bootstrap((DataImport.Source)null);
            identical(importer.data(), powerpack); check(storage.list().length == 1);
            String d = System.getenv("PF_DELUXE_CD_DATA");
            if (d != null) {
                Source deluxe = new Source(d);
                importer.install(deluxe); identical(importer.data(), deluxe);
                importer.bootstrap((DataImport.Source)null);
                int before = engine.stops;
                String[] prgs = {"INTRO.PRG", "TABLE1.PRG", "TABLE2.PRG", "TABLE3.PRG", "TABLE4.PRG"};
                for (String other : new String[]{a,b}) for (int mask=1; mask<31; mask++) {
                    Source hybrid = new Source(d);
                    for (int i=0; i<prgs.length; i++) if ((mask & (1<<i)) != 0)
                        hybrid.replacements.put(prgs[i], new File(other, prgs[i]));
                    rejects(importer, hybrid); identical(importer.data(), deluxe);
                }
                String c = System.getenv("PF_UNSUPPORTED_CD_DATA");
                if (c != null) {
                    rejects(importer, new Source(c)); identical(importer.data(), deluxe);
                    for (String name : prgs) {
                        Source mix = new Source(d); mix.replacements.put(name,new File(c,name));
                        rejects(importer,mix); identical(importer.data(),deluxe);
                    }
                }
                Source damaged = new Source(d); damaged.replacements.put("INTRO.PRG",malformed);
                rejects(importer,damaged); identical(importer.data(),deluxe);
                check(engine.stops==before);
                rejects(failing,canonical); identical(importer.data(),deluxe);
                Files.move(importer.data().toPath(),new File(storage,"Data.previous").toPath());
                Files.createDirectory(new File(storage,"Data.staging-interrupted").toPath());
                importer.recover(); importer.bootstrap((DataImport.Source)null);
                identical(importer.data(),deluxe);
                Files.createDirectory(new File(storage,"Data.previous").toPath());
                importer.recover(); importer.bootstrap((DataImport.Source)null);
                identical(importer.data(),deluxe); check(storage.list().length==1);
                System.out.println("PASS private Android D staging/adoption/bootstrap, 60 A/B/D hybrids, C rejection, rollback and recovery");
            } else System.out.println("SKIP private Android D: set PF_DELUXE_CD_DATA");
            System.out.println("PASS private Android A/B staging, shared validation, six hybrid rejections, malformed/incomplete rejection, rollback, recovery, cleanup and relaunch");
        } finally { DataImport.remove(root); }
    }
}
