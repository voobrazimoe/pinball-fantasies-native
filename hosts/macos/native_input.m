#import "native_input.h"
#import <IOKit/hidsystem/IOLLEvent.h>
void pf_macos_modifiers(PFInput *input,uint16_t changedKey,uint64_t flags) {
    /* AppKit flagsChanged identifies the physical side that changed. Device
       masks can clear lost breaks, but unchanged held siblings cannot invent
       new presses after focus regain. */
    const uint16_t keys[6]={56,60,59,62,58,61};
    const uint64_t masks[6]={NX_DEVICELSHIFTKEYMASK,NX_DEVICERSHIFTKEYMASK,
        NX_DEVICELCTLKEYMASK,NX_DEVICERCTLKEYMASK,NX_DEVICELALTKEYMASK,NX_DEVICERALTKEYMASK};
    bool sides[6];
    for (unsigned i=0;i<6;i++)
        sides[i]=(flags&masks[i]) && (changedKey==keys[i] || input->down[keys[i]]);
    pf_input_modifiers(input,sides);
}
