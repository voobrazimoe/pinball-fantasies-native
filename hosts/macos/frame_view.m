#import "frame_view.h"
#import "app.h"
#import <mach/mach_time.h>
#import <QuartzCore/QuartzCore.h>
@implementation PFFrameView {
    CGImageRef _image;
    int32_t _width, _height;
    CALayer *_frameLayer;
    NSTrackingArea *_tracking;
}
- (BOOL)isFlipped { return YES; }
- (BOOL)isOpaque { return YES; }
- (BOOL)acceptsFirstResponder { return YES; }
- (BOOL)pointerInside {
    NSPoint p=[self convertPoint:self.window.mouseLocationOutsideOfEventStream fromView:nil];
    return self.window.isKeyWindow && NSPointInRect(p,self.bounds);
}
- (void)updateTrackingAreas {
    [super updateTrackingAreas];
    if (_tracking) [self removeTrackingArea:_tracking];
    _tracking=[[NSTrackingArea alloc] initWithRect:NSZeroRect
        options:NSTrackingMouseEnteredAndExited|NSTrackingMouseMoved|NSTrackingActiveInKeyWindow|NSTrackingInVisibleRect
        owner:self userInfo:nil];
    [self addTrackingArea:_tracking];
}
- (instancetype)initWithFrame:(NSRect)frame {
    if ((self=[super initWithFrame:frame])) {
        self.layerContentsRedrawPolicy=NSViewLayerContentsRedrawOnSetNeedsDisplay;
        self.wantsLayer=YES;
    }
    return self;
}
- (BOOL)wantsUpdateLayer { return YES; }
- (BOOL)acceptPixels:(const uint8_t *)pixels width:(int32_t)w height:(int32_t)h stride:(int32_t)s {
    if (!pixels || !pf_frame_valid(w,h,s)) return NO;
    /* Core Animation may retain an older image while compositing. Give every
       submitted image immutable storage, independent of the borrowed ABI view. */
    CFDataRef data=CFDataCreate(NULL,pixels,(CFIndex)s*h);
    if (!data) return NO;
    CGDataProviderRef provider=CGDataProviderCreateWithCFData(data);
    CGColorSpaceRef color=CGColorSpaceCreateDeviceRGB();
    CGImageRef image=CGImageCreate(w,h,8,32,s,color,
        kCGBitmapByteOrder32Big|kCGImageAlphaLast,provider,NULL,false,kCGRenderingIntentDefault);
    CGDataProviderRelease(provider); CGColorSpaceRelease(color); CFRelease(data);
    if (!image) return NO;
    if (_image) CGImageRelease(_image);
    _image=image; _width=w; _height=h;
    self.needsDisplay=YES; return YES;
}
- (void)placeFrame {
    if (!_image || !_frameLayer) return;
    double lw,lh; pf_logical_size(_width,_height,&lw,&lh);
    double scale=fmin(self.bounds.size.width/lw,self.bounds.size.height/lh);
    _frameLayer.frame=CGRectMake((self.bounds.size.width-lw*scale)/2,
        (self.bounds.size.height-lh*scale)/2,lw*scale,lh*scale);
}
- (void)layout {
    [super layout];
    [CATransaction begin]; [CATransaction setDisableActions:YES];
    [self placeFrame];
    [CATransaction commit];
}
- (void)updateLayer {
    uint64_t start=mach_absolute_time();
    [CATransaction begin]; [CATransaction setDisableActions:YES];
    self.layer.backgroundColor=CGColorGetConstantColor(kCGColorBlack);
    if (!_frameLayer) {
        _frameLayer=[CALayer layer];
        _frameLayer.magnificationFilter=kCAFilterNearest;
        _frameLayer.minificationFilter=kCAFilterNearest;
        _frameLayer.contentsGravity=kCAGravityResize;
        [self.layer addSublayer:_frameLayer];
    }
    _frameLayer.contents=(__bridge id)_image;
    [self placeFrame];
    [CATransaction commit];
    [self.host frameDrawnFrom:start];
}
- (void)keyDown:(NSEvent *)e { [self.host key:e down:YES]; }
- (void)keyUp:(NSEvent *)e { [self.host key:e down:NO]; }
- (void)flagsChanged:(NSEvent *)e { [self.host modifiers:e]; }
- (void)mouseMoved:(NSEvent *)e { [self.host motion:e]; }
- (void)mouseDragged:(NSEvent *)e { [self.host motion:e]; }
- (void)mouseDown:(NSEvent *)e { (void)e; [self.host button:YES]; }
- (void)mouseUp:(NSEvent *)e { (void)e; [self.host button:NO]; }
- (void)mouseEntered:(NSEvent *)e { (void)e; [self.host updateCursor]; }
- (void)mouseExited:(NSEvent *)e { (void)e; [self.host updateCursor]; }
- (void)dealloc { if (_image) CGImageRelease(_image); }
@end
