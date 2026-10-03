#import <AppKit/AppKit.h>
#import "host_logic.h"
@interface PFApp : NSObject <NSApplicationDelegate, NSWindowDelegate>
- (void)key:(NSEvent *)event down:(BOOL)down;
- (void)modifiers:(NSEvent *)event;
- (void)motion:(NSEvent *)event;
- (void)button:(BOOL)down;
- (void)updateCursor;
@end
