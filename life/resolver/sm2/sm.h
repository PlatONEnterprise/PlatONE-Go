#include <openssl/bn.h>
#include <openssl/crypto.h>
#include <openssl/evp.h>
#include <openssl/ec.h>
#include <openssl/pem.h>
#include <openssl/bio.h>
#include <openssl/ossl_typ.h>


#ifdef __cplusplus
extern "C" {
#endif

struct ECDSA_SIG_st {
    BIGNUM *r;
    BIGNUM *s;
};


void dump_memory(const unsigned char *buf, unsigned int len);
int base64_encode(const char *in_str, int in_len, char *out_str);
int base64_decode(const char *in_str, int in_len, char *out_str);

static BIGNUM *sm2_compute_msg_hash(const EVP_MD *digest,
                                    const EC_KEY *key,
                                    const uint8_t *msg, size_t msg_len);

int sm2_compute_digest(uint8_t *out,
                         const EVP_MD *digest,
                         const EC_KEY *key);
int sm_generate_sig(const char *msg, const char *privkey, char *output);
int sm_verify_sig(const char *msg, const char *pubkey,const char *signature);
int sm2_sign_with_base64(const char *msg, const char *userid, const char *privkey, char *out);
int sm2_verify_with_base64(const char* msg, const char* userid, const char* pub_data, const char* sig_data);

#ifdef __cplusplus
}
#endif