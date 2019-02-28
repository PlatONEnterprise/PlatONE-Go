/******************************************************************************
 * Copyright (c) 2012-2020, Juzhenyuan TECHNOLOGIES (SHENZHEN) LTD.
 * File        : nizkpail.h
 * Version     : 0.1
 * Description : -
 * Author      : Liao Yan
 * Date        : 2019-02-26
*******************************************************************************/

#ifndef NIZKPAIL_H_
#define NIZKPAIL_H_

#include <stdlib.h>

#ifdef __cplusplus
extern "C" {
#endif

char* pailEncrypt(const char* _number, const char* _pubkey);

char* pailHomAdd(const char* _cipher1, const char* _cipher2, const char* _pubkey);

char* pailHomSub(const char* _cipher1, const char* _cipher2, const char* _pubkey);

char* nizkVerifyProof(const char* _pai,
                    const char* _fromBalCipher,
                    const char* _fromAmountCipher,
                    const char* _toAmountCipher,
                    const char* _fromPubkey,
                    const char* _toPubkey);

#ifdef __cplusplus
}
#endif

#endif // NIZKPAIL_H_

