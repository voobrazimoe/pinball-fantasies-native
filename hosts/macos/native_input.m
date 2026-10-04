#import "native_input.h"
#import <AppKit/AppKit.h>

static const uint16_t keys[6]={56,60,59,62,58,61};
static const uint64_t classes[3]={NSEventModifierFlagShift,
    NSEventModifierFlagControl,NSEventModifierFlagOption};

static void reconcile(PFInput *input,uint64_t flags) {
    /* An absent class proves both sides up, including breaks missed while
       inactive. A present class cannot identify a side or create a make. */
    for (unsigned i=0;i<6;i++) if (!(flags&classes[i/2])) {
        input->physicalModifiers[i]=false;
        input->down[keys[i]]=false;
    }
}
void pf_macos_focus(PFInput *input,bool focused,uint64_t flags) {
    /* AppKit's public aggregate snapshot only clears known released classes.
       Retain known physical sides so their first break after regain is not
       mistaken for a fresh press. No global per-key polling is needed. */
    reconcile(input,flags);
    pf_input_focus(input,focused);
}
void pf_macos_modifiers(PFInput *input,uint16_t changedKey,uint64_t flags) {
    unsigned changed=6;
    for (unsigned i=0;i<6;i++) if (changedKey==keys[i]) changed=i;
    /* Command, Caps Lock and unrelated events cannot alter flippers. */
    if (changed==6) return;
    bool wasDown=input->physicalModifiers[changed];
    reconcile(input,flags);
    bool pressed=(flags&classes[changed/2]) && !wasDown;
    input->physicalModifiers[changed]=pressed;
    bool sides[6];
    for (unsigned i=0;i<6;i++)
        sides[i]=input->physicalModifiers[i] &&
            (input->down[keys[i]] || (i==changed && pressed));
    pf_input_modifiers(input,sides);
}
