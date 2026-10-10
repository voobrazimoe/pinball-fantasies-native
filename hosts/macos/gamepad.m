#import "gamepad.h"
#import <GameController/GameController.h>

@interface PFGamepad ()
- (void)selectExcluding:(GCController *)excluded;
- (void)sample;
@end
@implementation PFGamepad {
    GCController *_controller;
    PFGamepadSink _sink;
    id _connect, _disconnect;
    uint32_t _buttons;
    int32_t _left, _right;
    BOOL _closed;
    BOOL _previousBackgroundMonitoring;
    NSArray<GCController *> *(^_controllers)(void);
}
static NSArray *buttons(GCExtendedGamepad *p) {
    id none=NSNull.null;
    return @[p.buttonA,p.buttonB,p.buttonX,p.buttonY,p.buttonOptions ?: none,none,
             p.buttonMenu,p.leftThumbstickButton ?: none,p.rightThumbstickButton ?: none,
             p.leftShoulder,p.rightShoulder,p.dpad.up,p.dpad.down,p.dpad.left,p.dpad.right];
}
static uint32_t buttonMask(GCExtendedGamepad *p) {
    uint32_t mask=0; NSArray *list=buttons(p);
    for (unsigned i=0;i<list.count;i++)
        if (list[i]!=NSNull.null && [(GCControllerButtonInput *)list[i] isPressed]) mask|=1u<<i;
    return mask;
}
- (instancetype)initWithSink:(PFGamepadSink)sink {
    return [self initWithSink:sink controllers:^{return GCController.controllers;}];
}
- (instancetype)initWithSink:(PFGamepadSink)sink controllers:(NSArray<GCController *> *(^)(void))controllers {
    self=[super init]; if (!self) return nil;
    _previousBackgroundMonitoring=GCController.shouldMonitorBackgroundEvents;
    GCController.shouldMonitorBackgroundEvents=YES;
    _controllers=[controllers copy];
    _sink=[sink copy]; __weak PFGamepad *weakSelf=self;
    NSNotificationCenter *center=NSNotificationCenter.defaultCenter;
    _connect=[center addObserverForName:GCControllerDidConnectNotification object:nil
        queue:NSOperationQueue.mainQueue usingBlock:^(NSNotification *n) {
            (void)n; [weakSelf selectExcluding:nil];
        }];
    _disconnect=[center addObserverForName:GCControllerDidDisconnectNotification object:nil
        queue:NSOperationQueue.mainQueue usingBlock:^(NSNotification *n) {
            PFGamepad *owner=weakSelf;
            if (owner && n.object==owner->_controller) [owner selectExcluding:n.object];
        }];
    [self selectExcluding:nil];
    return self;
}
- (void)selectExcluding:(GCController *)excluded {
    if (_closed || (_controller && _controller!=excluded)) return;
    if (_controller) {
        _controller.extendedGamepad.valueChangedHandler=nil;
        _controller=nil; _sink(3,0,0);
    }
    for (GCController *candidate in _controllers()) {
        if (candidate==excluded || !candidate.extendedGamepad) continue;
        _controller=candidate; _controller.handlerQueue=dispatch_get_main_queue();
        GCExtendedGamepad *profile=candidate.extendedGamepad;
        _buttons=buttonMask(profile);
        _left=(int32_t)(profile.leftTrigger.value*32767);
        _right=(int32_t)(profile.rightTrigger.value*32767);
        uint32_t mask=_buttons;
        if (_left>12000) mask|=1u<<15;
        if (_right>12000) mask|=1u<<16;
        _sink(2,(int32_t)mask,0);
        int32_t family=0;
        if([profile isKindOfClass:GCDualShockGamepad.class] || [profile isKindOfClass:GCDualSenseGamepad.class]) family=2;
        else if([profile isKindOfClass:GCXboxGamepad.class]) family=1;
        _sink(4,family,0);
        NSArray *face=@[profile.buttonA,profile.buttonB,profile.buttonX,profile.buttonY];
        for(unsigned i=0;i<4;i++) {
            NSString *symbol=[(GCControllerButtonInput *)face[i] sfSymbolsName];
            int32_t glyph=0;
            if([symbol containsString:@"triangle"]) glyph=8;
            else if([symbol containsString:@"square"]) glyph=7;
            else if([symbol containsString:@"xmark"]) glyph=5;
            else if([symbol isEqualToString:@"circle"] || [symbol containsString:@"circle.circle"]) glyph=6;
            else if([symbol hasPrefix:@"a."]) glyph=1;
            else if([symbol hasPrefix:@"b."]) glyph=2;
            else if([symbol hasPrefix:@"x."]) glyph=3;
            else if([symbol hasPrefix:@"y."]) glyph=4;
            if(glyph) _sink(5,i,glyph);
        }
        __weak PFGamepad *weakSelf=self; __weak GCController *weakController=candidate;
        profile.valueChangedHandler=^(GCExtendedGamepad *p,GCControllerElement *element) {
            (void)p; (void)element; PFGamepad *owner=weakSelf;
            if (owner && owner->_controller==weakController && !owner->_closed) [owner sample];
        };
        break;
    }
}
- (void)sample {
    GCExtendedGamepad *p=_controller.extendedGamepad;
    uint32_t next=buttonMask(p), changed=next^_buttons; _buttons=next;
    for (unsigned i=0;i<15;i++) if (changed&(1u<<i)) _sink(0,i,(next&(1u<<i))!=0);
    int32_t left=(int32_t)(p.leftTrigger.value*32767), right=(int32_t)(p.rightTrigger.value*32767);
    if (left!=_left) { _left=left; _sink(1,4,left); }
    if (right!=_right) { _right=right; _sink(1,5,right); }
}
- (void)close {
    if (_closed) return; _closed=YES;
    _controller.extendedGamepad.valueChangedHandler=nil; _controller=nil;
    NSNotificationCenter *center=NSNotificationCenter.defaultCenter;
    if (_connect) [center removeObserver:_connect];
    if (_disconnect) [center removeObserver:_disconnect];
    _connect=nil; _disconnect=nil; _sink(3,0,0);
    GCController.shouldMonitorBackgroundEvents=_previousBackgroundMonitoring;
}
- (void)dealloc { [self close]; }
@end
