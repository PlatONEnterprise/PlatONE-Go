/******************************************************************************
 * Copyright (c) 2012-2020, Juzhenyuan TECHNOLOGIES (SHENZHEN) LTD.
 * File        : nizkpail.hpp
 * Version     : 0.1
 * Description : -
 * Author      : Liao Yan
 * Date        : 2019-02-21
*******************************************************************************/

#pragma once

#include "fixedhash.hpp"

extern "C" {
	void pailEncrypt(const char* _number, int _numberSize, 
						const char* _pubkey, int _pubkeySize,
						char* _result, int _resultSize);
	
	void pailHomAdd(const char* _cipher1, int _cipher1Size,
						const char* _cipher2, int _cipher2Size,
						const char* _pubkey, int _pubkeySize,
						char* _result, int _resultSize);
	
	void pailHomSub(const char* _cipher1, int _cipher1Size,
						const char* _cipher2, int _cipher2Size,
						const char* _pubkey, int _pubkeySize,
						char* _result, int _resultSize);
	
	void nizkVerifyProof(const char* _pai, int _paiSize,
						const char* _fromBalCipher, int _fromBalCipherSize,
						const char* _fromAmountCipher, int _fromAmountCipherSize,
						const char* _toAmountCipher, int _toAmountCipherSize,
						const char* _fromPubkey, int _fromPubkeySize,
						const char* _toPubkey, int _toPubkeySize,
						char* _result, int _resultSize);
}

namespace bcwasm {
	std::string pailEncrypt(const std::string& _number, const std::string& _pubkey) {
		int resultSize = 1025;
		char* result = (char*)malloc(resultSize);
		memcpy(result, 0, resultSize);
		
		::pailEncrypt(_number.data(), _number.size(), _pubkey.data(), _pubkey.size(), result, resultSize);
		
		std::string ret = result;
		free(result);
		
		return ret;
	}

	std::string pailHomAdd(const std::string& _cipher1, const std::string& _cipher2, const std::string& _pubkey) {
		int resultSize = 1025;
		char* result = (char*)malloc(resultSize);
		memcpy(result, 0, resultSize);
		
		::pailHomAdd(_cipher1.data(), _cipher1.size(), _cipher2.data(), _cipher2.size(), _pubkey.data(), _pubkey.size(), result, resultSize);
		
		std::string ret = result;
		free(result);
		
		return ret;
	}

	std::string pailHomSub(const std::string& _cipher1, const std::string& _cipher2, const std::string& _pubkey) {
		int resultSize = 1025;
		char* result = (char*)malloc(resultSize);
		memcpy(result, 0, resultSize);
		
		::pailHomSub(_cipher1.data(), _cipher1.size(), _cipher2.data(), _cipher2.size(), _pubkey.data(), _pubkey.size(), result, resultSize);
		
		std::string ret = result;
		free(result);
		
		return ret;
	}

	
	std::string nizkVerifyProof(const std::string& _pai,
								const std::string& _fromBalCipher,
								const std::string& _fromAmountCipher,
								const std::string& _toAmountCipher,
								const std::string& _fromPubkey,
								const std::string& _toPubkey) {
		int resultSize = 1025;
		char* result = (char*)malloc(resultSize);
		memcpy(result, 0, resultSize + 1);
		
		::nizkVerifyProof(_pai.data(), _pai.size(),
							_fromBalCipher.data(), _fromBalCipher.size(),
							_fromAmountCipher.data(), _fromAmountCipher.size(),
							_toAmountCipher.data(), _toAmountCipher.size(),
							_fromPubkey.data(), _fromPubkey.size(),
							_toPubkey.data(), _toPubkey.size(),
							result, resultSize);
		
		std::string ret = result;
		free(result);
		
		return ret;
	}
}

