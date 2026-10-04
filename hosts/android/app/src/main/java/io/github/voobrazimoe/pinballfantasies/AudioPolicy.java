package io.github.voobrazimoe.pinballfantasies;

// UI-thread reducer, independent of Android for asset-free policy tests.
final class AudioPolicy {
    interface Host {
        boolean request();
        void abandon();
        void output(boolean enabled);
    }
    private final Host host;
    private boolean eligible, requested, granted, closed;
    AudioPolicy(Host host) { this.host = host; }
    void eligible(boolean value) {
        if (closed || eligible == value) return;
        eligible = value;
        if (!value) { granted = false; host.output(false); release(); }
        else {
            requested = true;
            granted = host.request();
            if (!granted) release();
            host.output(granted);
        }
    }
    // Android constants: gain=1, permanent=-1, transient=-2, duck=-3.
    void focus(int event) {
        if (closed || !eligible || !requested) return;
        if (event == 1) { granted = true; host.output(true); }
        else if (event < 0) {
            granted = false; host.output(false);
            if (event == -1) release();
        }
    }
    private void release() {
        if (requested) { requested = false; host.abandon(); }
    }
    void close() {
        if (closed) return;
        eligible(false); closed = true;
    }
}
