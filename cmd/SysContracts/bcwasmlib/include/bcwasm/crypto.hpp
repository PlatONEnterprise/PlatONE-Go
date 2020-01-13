#pragma once

#include "fixedhash.hpp"

extern "C" {
	void smSigVerify(const char* _msg, int _msgSize, const char* _userid, int _useridSize,
						const char* _pubkey, int _pubkeySize, const char * _sig, int _sigSize,
						char* _result, int _resultSize);
    void sm2secSigVerify(const char* _msg, int _msgSize, const char* _pubkey, int _pubkeySize, const char * _sig, int _sigSize,
						char* _result, int _resultSize);
    void secp256r1SigVerify(const char* _msg, int _msgSize, const char* _pubkey, int _pubkeySize, const char * _sig, int _sigSize,
						char* _result, int _resultSize);
    void secp256k1SigVerify(const char* _msg, int _msgSize, const char* _pubkey, int _pubkeySize, const char * _sig, int _sigSize,
						char* _result, int _resultSize);
}

namespace bcwasm {
	std::string sigVerify(const std::string& _msg, const std::string& _userid, const std::string& _pubkey, const std::string& _sig) {
		int resultSize = 1;
		char* result = (char*)malloc(resultSize);
		memcpy(result, 0, resultSize);
		::smSigVerify(_msg.data(), _msg.size(), _userid.data(), _userid.size(), _pubkey.data(), _pubkey.size(), _sig.data(), _sig.size(), result, resultSize);
		std::string ret = result;
		free(result);
		return ret;
	}
    std::string sm2secSigVerify(const std::string& _msg, const std::string& _pubkey, const std::string& _sig) {
        int resultSize = 1;
		char* result = (char*)malloc(resultSize);
		memcpy(result, 0, resultSize);
		::sm2secSigVerify(_msg.data(), _msg.size(), _pubkey.data(), _pubkey.size(), _sig.data(), _sig.size(), result, resultSize);
		std::string ret = result;
		free(result);
		return ret;
    }
    std::string secp256k1SigVerify(const std::string& _msg, const std::string& _pubkey, const std::string& _sig) {
        int resultSize = 1;
		char* result = (char*)malloc(resultSize);
		memcpy(result, 0, resultSize);
		::secp256k1SigVerify(_msg.data(), _msg.size(), _pubkey.data(), _pubkey.size(), _sig.data(), _sig.size(), result, resultSize);
		std::string ret = result;
		free(result);
		return ret;
    }

    std::string secp256r1SigVerify(const std::string& _msg, const std::string& _pubkey, const std::string& _sig) {
        int resultSize = 1;
		char* result = (char*)malloc(resultSize);
		memcpy(result, 0, resultSize);
		::secp256r1SigVerify(_msg.data(), _msg.size(), _pubkey.data(), _pubkey.size(), _sig.data(), _sig.size(), result, resultSize);
		std::string ret = result;
		free(result);
		return ret;
    }
}