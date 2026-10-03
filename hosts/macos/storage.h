#import <Foundation/Foundation.h>
typedef BOOL (*PFValidateAssets)(NSString *data, NSError **error);
NSArray<NSString *> *pf_required_assets(void);
NSString *pf_support_path(NSString *base);
BOOL pf_validate_assets(NSString *data,NSError **error);
BOOL pf_import_assets(NSString *source,NSString *destination,PFValidateAssets validate,NSError **error);
BOOL pf_storage_prepare(NSString **data,NSString **state,NSError **error);
