#import "storage.h"
#import <AppKit/AppKit.h>
#include "../../cmd/pfengine/abi.h"
static NSError *failure(NSString *text) {
    return [NSError errorWithDomain:@"PinballFantasies.Import" code:1
        userInfo:@{NSLocalizedDescriptionKey:text}];
}
NSArray<NSString *> *pf_required_assets(void) {
    return @[@"INTRO.PRG",@"INTRO.MOD",@"MOD2.MOD",@"TABLE1.PRG",@"TABLE1.MOD",
             @"TABLE2.PRG",@"TABLE2.MOD",@"TABLE3.PRG",@"TABLE3.MOD",@"TABLE4.PRG",@"TABLE4.MOD"];
}
NSString *pf_support_path(NSString *base) { return [base stringByAppendingPathComponent:@"PinballFantasies"]; }
BOOL pf_validate_assets(NSString *data,NSError **error) {
    NSFileManager *fm=NSFileManager.defaultManager;
    NSString *state=[NSTemporaryDirectory() stringByAppendingPathComponent:NSUUID.UUID.UUIDString];
    if (![fm createDirectoryAtPath:state withIntermediateDirectories:NO attributes:@{NSFilePosixPermissions:@0700} error:error]) return NO;
    char message[1024]={0};
    uint64_t engine=pf_engine_create((char *)data.fileSystemRepresentation,(char *)state.fileSystemRepresentation,0,message,sizeof(message));
    BOOL valid=engine!=0;
    if (engine && pf_engine_destroy(engine)!=PF_OK) { valid=NO; strcpy(message,"Shared settings validation failed"); }
    [fm removeItemAtPath:state error:NULL];
    if (!valid && error) *error=failure([NSString stringWithUTF8String:message]);
    return valid;
}
static BOOL regular(NSString *path,NSError **error) {
    NSURL *url=[NSURL fileURLWithPath:path]; NSNumber *file=nil,*link=nil;
    if (![url getResourceValue:&file forKey:NSURLIsRegularFileKey error:error] ||
        ![url getResourceValue:&link forKey:NSURLIsSymbolicLinkKey error:error]) return NO;
    if (!file.boolValue || link.boolValue) {
        if (error) *error=failure([NSString stringWithFormat:@"Select a regular original file: %@",path.lastPathComponent]);
        return NO;
    }
    return YES;
}
BOOL pf_import_assets(NSString *source,NSString *destination,PFValidateAssets validate,NSError **error) {
    NSFileManager *fm=NSFileManager.defaultManager;
    NSString *parent=destination.stringByDeletingLastPathComponent;
    NSString *stage=[parent stringByAppendingPathComponent:[@".import-" stringByAppendingString:NSUUID.UUID.UUIDString]];
    if (![fm createDirectoryAtPath:stage withIntermediateDirectories:NO attributes:@{NSFilePosixPermissions:@0700} error:error]) return NO;
    BOOL ok=NO;
    NSMutableArray *names=[pf_required_assets() mutableCopy];
    if ([fm fileExistsAtPath:[source stringByAppendingPathComponent:@"PINBALL.CFG"]]) [names addObject:@"PINBALL.CFG"];
    for (NSString *name in names) {
        NSString *input=[source stringByAppendingPathComponent:name];
        if (!regular(input,error) || ![fm copyItemAtPath:input toPath:[stage stringByAppendingPathComponent:name] error:error]) goto cleanup;
    }
    /* Shared Go hashes, decoders and settings semantics are authoritative. */
    if (!validate(stage,error)) goto cleanup;
    if ([fm fileExistsAtPath:destination]) {
        NSString *backup=[@".previous-" stringByAppendingString:NSUUID.UUID.UUIDString];
        BOOL result=[fm replaceItemAtURL:[NSURL fileURLWithPath:destination]
            withItemAtURL:[NSURL fileURLWithPath:stage] backupItemName:backup
            options:0 resultingItemURL:NULL error:error];
        if (!result) goto cleanup;
        [fm removeItemAtPath:[parent stringByAppendingPathComponent:backup] error:NULL];
        ok=YES;
    } else ok=[fm moveItemAtPath:stage toPath:destination error:error];
cleanup:
    if (!ok) [fm removeItemAtPath:stage error:NULL];
    return ok;
}
BOOL pf_storage_prepare(NSString **data,NSString **state,NSError **error) {
    NSFileManager *fm=NSFileManager.defaultManager;
    NSURL *base=[fm URLForDirectory:NSApplicationSupportDirectory inDomain:NSUserDomainMask
                    appropriateForURL:nil create:YES error:error];
    if (!base) return NO;
    NSString *root=pf_support_path(base.path);
    if (![fm createDirectoryAtPath:root withIntermediateDirectories:YES attributes:@{NSFilePosixPermissions:@0700} error:error]) return NO;
    *state=[root stringByAppendingPathComponent:@"State"];
    *data=[root stringByAppendingPathComponent:@"Data"];
    if (![fm createDirectoryAtPath:*state withIntermediateDirectories:YES attributes:@{NSFilePosixPermissions:@0700} error:error]) return NO;
    if ([fm fileExistsAtPath:*data] && pf_validate_assets(*data,NULL)) return YES;
    while (YES) {
        NSAlert *explanation=[NSAlert new]; explanation.messageText=@"Original game files are required";
        explanation.informativeText=@"Select the folder containing your original DOS Pinball Fantasies INTRO.PRG, INTRO.MOD, MOD2.MOD and TABLE1–4 PRG/MOD files. Validated files will be copied to Application Support. Game data is not included.";
        [explanation addButtonWithTitle:@"Choose Files…"]; [explanation addButtonWithTitle:@"Quit"];
        if ([explanation runModal]!=NSAlertFirstButtonReturn) return NO;
        NSOpenPanel *panel=[NSOpenPanel openPanel]; panel.canChooseDirectories=YES;
        panel.canChooseFiles=YES; panel.allowsMultipleSelection=NO;
        panel.message=@"Choose the original game folder or any required file inside it";
        if ([panel runModal]!=NSModalResponseOK) return NO;
        NSNumber *directory=nil;
        [panel.URL getResourceValue:&directory forKey:NSURLIsDirectoryKey error:NULL];
        NSString *source=directory.boolValue?panel.URL.path:panel.URL.path.stringByDeletingLastPathComponent;
        NSError *importError=nil;
        if (pf_import_assets(source,*data,pf_validate_assets,&importError)) return YES;
        NSAlert *bad=[NSAlert new]; bad.messageText=@"These files could not be imported";
        bad.informativeText=importError.localizedDescription ?: @"Required original files were not validated";
        [bad runModal];
    }
}
