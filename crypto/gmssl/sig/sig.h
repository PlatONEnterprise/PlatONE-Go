#include <string.h>
#include <openssl/bn.h>
#include <openssl/crypto.h>
#include <openssl/evp.h>
#include <openssl/ec.h>
#include <openssl/pem.h>
#include <openssl/bio.h>
#include <openssl/ossl_typ.h>

#include <openssl/rand.h>
#include <openssl/sm2.h>

void dump_memory(const unsigned char *buf, unsigned int len);
int base64_encode(const char *in_str, int in_len, char *out_str);
int base64_decode(const char *in_str, int in_len, char *out_str);

int sm_generate_sig(const char *msg, const char *privkey, char *output);
int sm_verify_sig(const char *msg, const char *pubkey,const char *signature);
int sm2_sign_with_base64(const char *msg, const char *userid, const char *privkey, char *out);
int sm2_verify_with_base64(const char* msg, const char* userid, const char* pub_data, const char* sig_data);

int p256r1_sign_with_base64(const char *msg, const char *privkey, char *out);
int p256r1_verify_with_base64(const char* msg, const char* pub_data, const char* sig_data);

int p256k1_sign_with_base64(const char *msg, const char *privkey, char *out);
int p256k1_verify_with_base64(const char* msg, const char* pub_data, const char* sig_data);