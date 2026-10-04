#import "native_input.h"
#import <IOKit/hidsystem/IOLLEvent.h>
void pf_macos_modifiers(PFInput *input,uint64_t flags) {
    /* AppKit flagsChanged carries the IOKit device-side flags in addition to
       aggregate Shift/Control/Option. Never toggle by aggregate state alone. */
    const uint64_t masks[6]={NX_DEVICELSHIFTKEYMASK,NX_DEVICERSHIFTKEYMASK,
        NX_DEVICELCTLKEYMASK,NX_DEVICERCTLKEYMASK,NX_DEVICELALTKEYMASK,NX_DEVICERALTKEYMASK};
    bool sides[6];
    for (unsigned i=0;i<6;i++) sides[i]=(flags&masks[i])!=0;
    pf_input_modifiers(input,sides);
}
