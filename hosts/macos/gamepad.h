#import <Foundation/Foundation.h>
@class GCController;
typedef void (^PFGamepadSink)(int32_t kind, int32_t a, int32_t b);
@interface PFGamepad : NSObject
- (instancetype)initWithSink:(PFGamepadSink)sink;
- (instancetype)initWithSink:(PFGamepadSink)sink controllers:(NSArray<GCController *> *(^)(void))controllers;
- (void)close;
@end
