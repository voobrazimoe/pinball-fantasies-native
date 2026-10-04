#include <dlfcn.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

enum { PF_ABI_VERSION_EXPECTED = 1 };

typedef int32_t (*abi_version_fn)(void);
typedef uint64_t (*create_fn)(char *, char *, int64_t, char *, uint32_t);

static void *require_symbol(void *library, const char *name) {
    void *symbol = dlsym(library, name);
    if (symbol == NULL) {
        fprintf(stderr, "missing symbol %s: %s\n", name, dlerror());
    }
    return symbol;
}

int main(int argc, char **argv) {
    if (argc != 2) {
        fprintf(stderr, "usage: %s /path/to/libpfengine.so\n", argv[0]);
        return 2;
    }

    void *library = dlopen(argv[1], RTLD_NOW | RTLD_LOCAL);
    if (library == NULL) {
        fprintf(stderr, "dlopen failed: %s\n", dlerror());
        return 1;
    }

    abi_version_fn abi_version =
        (abi_version_fn)require_symbol(library, "pf_engine_abi_version");
    create_fn create = (create_fn)require_symbol(library, "pf_engine_create");
    if (abi_version == NULL || create == NULL) {
        dlclose(library);
        return 1;
    }

    if (abi_version() != PF_ABI_VERSION_EXPECTED) {
        fprintf(stderr, "unexpected ABI version\n");
        dlclose(library);
        return 1;
    }

    char data[] = "/data/local/tmp/pfengine-a0-missing-data";
    char state[] = "/data/local/tmp/pfengine-a0-missing-state";
    char error[512];
    memset(error, 0, sizeof(error));
    if (create(data, state, 0, error, (uint32_t)sizeof(error)) != 0) {
        fprintf(stderr, "engine unexpectedly accepted missing data\n");
        dlclose(library);
        return 1;
    }
    if (error[0] == '\0') {
        fprintf(stderr, "missing-data failure did not return an error message\n");
        dlclose(library);
        return 1;
    }

    printf("OK: ABI %d loaded; missing-data create failed cleanly: %s\n",
           PF_ABI_VERSION_EXPECTED, error);
    dlclose(library);
    return 0;
}
