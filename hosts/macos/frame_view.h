#import <AppKit/AppKit.h>
#import "host_logic.h"
@class PFApp;
@interface PFFrameView : NSView
@property(nonatomic,weak) PFApp *host;
@property(nonatomic,readonly) BOOL pointerInside;
- (BOOL)acceptPixels:(const uint8_t *)pixels width:(int32_t)w height:(int32_t)h stride:(int32_t)s;
@end
