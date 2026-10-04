#import "frame_view.h"
#import "app.h"
#import <mach/mach_time.h>
#include <stdlib.h>
#include <string.h>
@implementation PFFrameView {
    uint8_t *_bytes;
    size_t _capacity;
    int32_t _width, _height, _stride;
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
- (BOOL)acceptPixels:(const uint8_t *)pixels width:(int32_t)w height:(int32_t)h stride:(int32_t)s {
    if (!pixels || !pf_frame_valid(w,h,s)) return NO;
    size_t required=(size_t)s*h;
    if (required>_capacity) {
        void *next=realloc(_bytes,required);
        if (!next) return NO;
        _bytes=next; _capacity=required;
    }
    /* ABI view can be invalidated by the next retrieval: retain only our copy. */
    memcpy(_bytes,pixels,required); _width=w; _height=h; _stride=s;
    self.needsDisplay=YES; return YES;
}
- (void)drawRect:(NSRect)dirty {
    uint64_t start=mach_absolute_time();
    (void)dirty;
    [[NSColor blackColor] setFill]; NSRectFill(self.bounds);
    if (!_bytes) return;
    CGContextRef ctx=NSGraphicsContext.currentContext.CGContext;
    CGColorSpaceRef color=CGColorSpaceCreateDeviceRGB();
    CGDataProviderRef provider=CGDataProviderCreateWithData(NULL,_bytes,(size_t)_stride*_height,NULL);
    CGImageRef image=CGImageCreate(_width,_height,8,32,_stride,color,
        kCGBitmapByteOrder32Big|kCGImageAlphaLast,provider,NULL,false,kCGRenderingIntentDefault);
    if (image) {
        double lw,lh; pf_logical_size(_width,_height,&lw,&lh);
        double scale=fmin(self.bounds.size.width/lw,self.bounds.size.height/lh);
        CGRect dest=CGRectMake((self.bounds.size.width-lw*scale)/2,
                              (self.bounds.size.height-lh*scale)/2,lw*scale,lh*scale);
        CGContextSaveGState(ctx); CGContextSetInterpolationQuality(ctx,kCGInterpolationNone);
        /* CoreGraphics image rows are top-down; compensate the flipped AppKit CTM. */
        CGContextTranslateCTM(ctx,dest.origin.x,dest.origin.y+dest.size.height);
        CGContextScaleCTM(ctx,1,-1);
        CGContextDrawImage(ctx,CGRectMake(0,0,dest.size.width,dest.size.height),image);
        CGContextRestoreGState(ctx); CGImageRelease(image);
    }
    CGDataProviderRelease(provider); CGColorSpaceRelease(color);
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
- (void)dealloc { free(_bytes); }
@end
