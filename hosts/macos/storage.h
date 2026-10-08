#import <Foundation/Foundation.h>
typedef BOOL (*PFValidateAssets)(NSString *data, NSError **error);
NSArray<NSString *> *pf_required_assets(void);
NSArray<NSString *> *pf_demo_assets(void);
NSArray<NSString *> *pf_installation_assets(NSString *source);
NSString *pf_support_path(NSString *base);
BOOL pf_validate_assets(NSString *data,NSError **error);
BOOL pf_import_assets(NSString *source,NSString *destination,PFValidateAssets validate,NSError **error);
BOOL pf_storage_prepare(NSString **data,NSString **state,NSError **error);
BOOL pf_storage_prepare_at(NSString *base,NSString *resources,NSString **data,NSString **state,NSError **error);
