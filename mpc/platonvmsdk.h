#ifndef _PlatONEVMSDK_H
#define _PlatONEVMSDK_H

#ifdef _WIN32
#  if defined(PlatONEMPCVMSDK_STATIC)
#    define PlatONEMPCVMSDK_DLL_API
#  else
#    if defined(PlatONEMPCVMSDK_EXPORTS)
#      define PlatONEMPCVMSDK_DLL_API __declspec(dllexport)
#    else
#      define PlatONEMPCVMSDK_DLL_API __declspec(dllimport)
#    endif
#  endif
#else
#  define PlatONEMPCVMSDK_DLL_API
#endif

/*
return 0 is success, other is failed
ERR_NO_ERROR	0
ERR_INIT_ENGINE	1
ERR_COMMIT_TASK	2
ERR_NOT_INIT	3
*/
#if __cplusplus
extern "C" {
#endif

	// new
	int PlatONEMPCVMSDK_DLL_API notify_security_init(const char* icecfg, const char* url);
	int PlatONEMPCVMSDK_DLL_API notify_security_commit(const char* taskid, const char* pubkey, const char* address,
		const char* ir_address, const char* method, const char* extra);

	// old
	int PlatONEMPCVMSDK_DLL_API notify_security_calculation(const char* taskid, const char* pubkey, const char* address,
		const char* ir_address, const char* method, const char* extra);

#if __cplusplus
}
#endif

#endif //!_PlatONEVMSDK_H


