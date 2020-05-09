#include "sig.h"
#define error(a) printf(a), exit(-1)
#define PUBLEN  65
#define PRIVLEN 32
#define MAXLEN 1000



int base64_decode(const char *in_str, int in_len, char *out_str)
{
    BIO *b64, *bio;
    BUF_MEM *bptr = NULL;
    int counts;
    int size = 0;

    if (in_str == NULL || out_str == NULL)
        return -1;

    b64 = BIO_new(BIO_f_base64());
    BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);

    bio = BIO_new_mem_buf(in_str, in_len);
    bio = BIO_push(b64, bio);

    size = BIO_read(bio, out_str, in_len);
    out_str[size] = '\0';

    BIO_free_all(bio);
    return size;
}



int base64_encode(const char *input, int length, char *out)
{
	BIO * bmem = NULL;
	BIO * b64 = NULL;
	BUF_MEM * bptr = NULL;
 
	b64 = BIO_new(BIO_f_base64());

	BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
	bmem = BIO_new(BIO_s_mem());
	b64 = BIO_push(b64, bmem);
	BIO_write(b64, input, length);
	BIO_flush(b64);
	BIO_get_mem_ptr(b64, &bptr);
 
	memcpy(out, bptr->data, bptr->length);
	out[bptr->length] = 0;
 
	BIO_free_all(b64);
 
	return strlen(out);
}
 

int sm2_sign_with_base64(const char *msg, const char *userid, const char *privkey, char *out)
{
    int siglen = 0;
     EC_KEY *key = EC_KEY_new_by_curve_name(NID_sm2);
    if (key == NULL)
    {
        error("Failed to initial ec key");
    }
    const EC_GROUP *group = EC_KEY_get0_group(key);
    if (group == NULL)
    {
        goto done;
    }
    EC_POINT *pub = EC_POINT_new(group);
    ECDSA_SIG *sig = NULL;
    BIGNUM *priv = NULL;
    BN_CTX *ctx = BN_CTX_new();
    unsigned char *pub_raw = (unsigned char *)malloc(MAXLEN);
    BN_hex2bn(&priv, privkey);
    if (EC_KEY_set_private_key(key, priv) != 1)
    {
        goto done;
    }
    EC_POINT_mul(group, pub, priv, NULL, NULL, NULL);
    EC_KEY_set_public_key(key, pub);
    sig = sm2_do_sign(key, EVP_sm3(), (const uint8_t *)userid, strlen(userid), (const uint8_t *)msg, strlen(msg));
    if (sig == NULL)
    {
        goto done;
    }
    unsigned char *sigd = (unsigned char  *)malloc(MAXLEN);
    memset(sigd, 0, MAXLEN);
    siglen = i2d_ECDSA_SIG(sig, &sigd);
    base64_encode(sigd - siglen, siglen, out);
    free(sigd-siglen);
    
done:
    EC_POINT_free(pub);
    EC_KEY_free(key);
    BN_free(priv);
    ECDSA_SIG_free(sig); 
    return siglen;
}


int sm2_verify_with_base64(const char* msg, const char* userid, const char* pub_data, const char* sig_data)
{
    int ret = 0;
    char pub_raw[MAXLEN];
    char sig_raw[MAXLEN];
    EC_KEY *key = EC_KEY_new_by_curve_name(NID_sm2);
    if (key == NULL)
    {
	goto done;
    }
    const EC_GROUP *group = EC_KEY_get0_group(key);
    if (group == NULL)
    {
        goto done;
    }
    BN_CTX *ctx = BN_CTX_new();
    ECDSA_SIG *sig = NULL;
    int pub_len = base64_decode(pub_data, strlen(pub_data), pub_raw);
    EC_POINT *pub = EC_POINT_new(group);
    if (pub == NULL)
    {
	goto done;
    }
    EC_POINT_oct2point(group, pub, pub_raw, pub_len , ctx);
    EC_KEY_set_public_key(key, pub);
    int sig_len = base64_decode(sig_data, strlen(sig_data), sig_raw);
    sig = ECDSA_SIG_new();
    const unsigned char *tmp = sig_raw;
    sig = d2i_ECDSA_SIG(NULL, &tmp, sig_len);
    ret = sm2_do_verify(key, EVP_sm3(), sig, (const uint8_t *)userid, strlen(userid), (const uint8_t *)msg, strlen(msg));

done:
    EC_POINT_free(pub);
    EC_KEY_free(key);
    ECDSA_SIG_free(sig);
    return ret;
}


int sm2_compute_digest(uint8_t *out,
                         const EVP_MD *digest,
                         const EC_KEY *key)
{
    int rc = 0;
    const EC_GROUP *group = EC_KEY_get0_group(key);
    BN_CTX *ctx = NULL;
    EVP_MD_CTX *hash = NULL;
    BIGNUM *p = NULL;
    BIGNUM *a = NULL;
    BIGNUM *b = NULL;
    BIGNUM *xG = NULL;
    BIGNUM *yG = NULL;
    BIGNUM *xA = NULL;
    BIGNUM *yA = NULL;
    int p_bytes = 0;
    uint8_t *buf = NULL;
    uint16_t entl = 0;

    hash = EVP_MD_CTX_new();
    ctx = BN_CTX_new();
    if (hash == NULL || ctx == NULL) {

        goto done;
    }

    p = BN_CTX_get(ctx);
    a = BN_CTX_get(ctx);
    b = BN_CTX_get(ctx);
    xG = BN_CTX_get(ctx);
    yG = BN_CTX_get(ctx);
    xA = BN_CTX_get(ctx);
    yA = BN_CTX_get(ctx);

    if (yA == NULL) {

        goto done;
    }

    if (!EVP_DigestInit(hash, digest)) {

        goto done;
    }

    if (!EC_GROUP_get_curve(group, p, a, b, ctx)) {

        goto done;
    }

    p_bytes = BN_num_bytes(p);
    buf = OPENSSL_zalloc(p_bytes);
    if (buf == NULL) {
       
        goto done;
    }
    uint8_t data = 0;
    EVP_DigestUpdate(hash, &data, 1);
    EVP_DigestUpdate(hash, &data, 1);
    char len[] = "";
    EVP_DigestUpdate(hash, &len, 0);
    if (BN_bn2binpad(a, buf, p_bytes) < 0
            || !EVP_DigestUpdate(hash, buf, p_bytes)
            || BN_bn2binpad(b, buf, p_bytes) < 0
            || !EVP_DigestUpdate(hash, buf, p_bytes)
            || !EC_POINT_get_affine_coordinates(group,
                                                EC_GROUP_get0_generator(group),
                                                xG, yG, ctx)
            || BN_bn2binpad(xG, buf, p_bytes) < 0
            || !EVP_DigestUpdate(hash, buf, p_bytes)
            || BN_bn2binpad(yG, buf, p_bytes) < 0
            || !EVP_DigestUpdate(hash, buf, p_bytes)
            || !EC_POINT_get_affine_coordinates(group,
                                                EC_KEY_get0_public_key(key),
                                                xA, yA, ctx)
            || BN_bn2binpad(xA, buf, p_bytes) < 0
            || !EVP_DigestUpdate(hash, buf, p_bytes)
            || BN_bn2binpad(yA, buf, p_bytes) < 0
            || !EVP_DigestUpdate(hash, buf, p_bytes)
            || !EVP_DigestFinal(hash, out, NULL)) {

        goto done;
    }
    rc = 1;

 done:
    OPENSSL_free(buf);
    BN_CTX_free(ctx);
    EVP_MD_CTX_free(hash);
    return rc;
}

static BIGNUM *sm2_compute_msg_hash(const EVP_MD *digest,
                                    const EC_KEY *key,
                                    const uint8_t *msg, size_t msg_len)
{
    EVP_MD_CTX *hash = EVP_MD_CTX_new();
    const int md_size = EVP_MD_size(digest);
    uint8_t *z = NULL;
    BIGNUM *e = NULL;
    z = OPENSSL_zalloc(md_size);
    sm2_compute_digest(z, digest, key);
    if (!EVP_DigestInit(hash, digest)
            || !EVP_DigestUpdate(hash, z, md_size)
            || !EVP_DigestUpdate(hash, msg, msg_len)
               /* reuse z buffer to hold H(Z || M) */
            || !EVP_DigestFinal(hash, z, NULL)) {
        goto done;
    }
    e = BN_bin2bn(z, md_size, NULL);
 done:
    OPENSSL_free(z);
    EVP_MD_CTX_free(hash);
    return e;
}


int sm_generate_sig(const char *msg, const char *privkey, char *output)
{
    int sig_len = 0;
    ECDSA_SIG *sig = NULL;
    BIGNUM *priv = NULL;
    EC_KEY *key = EC_KEY_new_by_curve_name(NID_sm2);
    if (key == NULL)
    {
        error("Failed to initial ec key");
    }
    const EC_GROUP *group = EC_KEY_get0_group(key);
    if (group == NULL)
    {
        goto done;
    }
    EC_POINT *pub = EC_POINT_new(group);
    BN_CTX *ctx = BN_CTX_new();
    BN_hex2bn(&priv, privkey);
    if (EC_KEY_set_private_key(key, priv) != 1)
    {
        goto done;
    }
    EC_POINT_mul(group, pub, priv, NULL, NULL, NULL);
    EC_KEY_set_public_key(key, pub);
    BIGNUM *e = sm2_compute_msg_hash(EVP_sm3(), key, msg, strlen(msg));
    unsigned char *dgst = (unsigned char *)malloc(MAXLEN);
    unsigned char *sigraw = (unsigned char *)malloc(MAXLEN);
    int e_len = BN_bn2bin(e, dgst);
    sm2_sign(dgst,e_len,
               sigraw, &sig_len, key);
    base64_encode(sigraw, sig_len, output);         
    free(dgst);
    free(sigraw);


done:
        EC_POINT_free(pub);
        EC_KEY_free(key);
        BN_free(priv);
        ECDSA_SIG_free(sig);
        return sig_len;

}

int sm_verify_sig(const char *msg, const char *pubkey,const char *signature)
{
    int ok = 0;
	char pub_raw[MAXLEN];
    char sig_raw[MAXLEN];
    EC_KEY *key = EC_KEY_new_by_curve_name(NID_sm2);
	if (key == NULL)
	{
		goto done;
	}
    const EC_GROUP *group = EC_KEY_get0_group(key);
    if (group == NULL)
    {
        goto done;
    }
    BN_CTX *ctx = BN_CTX_new();
    int pub_len = base64_decode(pubkey, strlen(pubkey), pub_raw);
    EC_POINT *pub = EC_POINT_new(group);
	if (pub == NULL)
	{
		goto done;
	}
    EC_POINT_oct2point(group, pub, pub_raw, pub_len, ctx);
    EC_KEY_set_public_key(key, pub);
    
    int sig_len = base64_decode(signature, strlen(signature), sig_raw);
    const unsigned char *sigtmp = sig_raw;
    unsigned int mdlen = 0;
    unsigned char *md = (unsigned char *)malloc(MAXLEN);
    EVP_Digest(msg, strlen(msg), md, &mdlen, EVP_sm3(), NULL);
    BIGNUM *e = sm2_compute_msg_hash(EVP_sm3(), key, md, 32); 
    unsigned char *dgst = (unsigned char *)malloc(MAXLEN);
    int e_len = BN_bn2bin(e, dgst);
    ok = sm2_verify(dgst, e_len,
               sigtmp, sig_len, key);
done:
        EC_POINT_free(pub);
        EC_KEY_free(key);
        free(md);
        return ok;

}


int p256r1_sign_with_base64(const char *msg, const char *privkey, char *out)
{
    int siglen = 0;
    EC_KEY *key = EC_KEY_new_by_curve_name(NID_X9_62_prime256v1);
    if (key == NULL)
    {
        error("Failed to initial ec key");
    }
    const EC_GROUP *group = EC_KEY_get0_group(key);
    if (group == NULL)
    {
        goto done;
    }
    ECDSA_SIG *sig = NULL;
    BIGNUM *priv = NULL;
    EC_POINT *pub = EC_POINT_new(group);
    BN_CTX *ctx = BN_CTX_new();
    BN_hex2bn(&priv, privkey);
    if (EC_KEY_set_private_key(key, priv) != 1)
    {
        goto done;
    }
    EC_POINT_mul(group, pub, priv, NULL, NULL, NULL);
    EC_KEY_set_public_key(key, pub);
    unsigned int mdlen = 0;
    unsigned char *md = (unsigned char *)malloc(MAXLEN);
    EVP_Digest(msg, strlen(msg), md, &mdlen, EVP_sha384(), NULL);
    sig = ECDSA_do_sign(md, mdlen,key);
    if (sig == NULL)
    {
        goto done;
    }
    unsigned char *sigd = (unsigned char  *)malloc(MAXLEN);
    memset(sigd, 0, MAXLEN);
    siglen = i2d_ECDSA_SIG(sig, &sigd);
    base64_encode(sigd - siglen, siglen, out);
    free(sigd-siglen);
done:
    EC_POINT_free(pub);
    EC_KEY_free(key);
    BN_free(priv);
    ECDSA_SIG_free(sig); 
    free(md);
    return siglen;
}


int p256r1_verify_with_base64(const char* msg, const char* pub_data, const char* sig_data)
{
	int ret = 0;
	char pub_raw[MAXLEN];
    char sig_raw[MAXLEN];
    EC_KEY *key = EC_KEY_new_by_curve_name(NID_X9_62_prime256v1);
	if (key == NULL)
	{
		goto done;
	}
    const EC_GROUP *group = EC_KEY_get0_group(key);
    if (group == NULL)
    {
        goto done;
    }
    BN_CTX *ctx = BN_CTX_new();
    ECDSA_SIG *sig = NULL;
    int pub_len = base64_decode(pub_data, strlen(pub_data), pub_raw);
    EC_POINT *pub = EC_POINT_new(group);
	if (pub == NULL)
	{
		goto done;
	}
    EC_POINT_oct2point(group, pub, pub_raw, pub_len , ctx);
    EC_KEY_set_public_key(key, pub);
    int sig_len = base64_decode(sig_data, strlen(sig_data), sig_raw);
    sig = ECDSA_SIG_new();
    const unsigned char *tmp = sig_raw;
    sig = d2i_ECDSA_SIG(NULL, &tmp, sig_len);
    unsigned int mdlen = 0;
    unsigned char *md = (unsigned char *)malloc(MAXLEN);
    EVP_Digest(msg, strlen(msg), md, &mdlen, EVP_sha384(), NULL);
    ret = ECDSA_do_verify(md, mdlen, sig, key);

	done:
        EC_POINT_free(pub);
        EC_KEY_free(key);
        ECDSA_SIG_free(sig);
        free(md);
    
    return ret;
}



int p256k1_sign_with_base64(const char *msg, const char *privkey, char *out)
{
    int siglen = 0;
    EC_KEY *key = EC_KEY_new_by_curve_name(NID_secp256k1);
    if (key == NULL)
    {
        error("Failed to initial ec key");
    }
    const EC_GROUP *group = EC_KEY_get0_group(key);
    if (group == NULL)
    {
        goto done;
    }
    ECDSA_SIG *sig = NULL;
    BIGNUM *priv = NULL;
    EC_POINT *pub = EC_POINT_new(group);
    BN_CTX *ctx = BN_CTX_new();
    BN_hex2bn(&priv, privkey);
    if (EC_KEY_set_private_key(key, priv) != 1)
    {
        goto done;
    }
    EC_POINT_mul(group, pub, priv, NULL, NULL, NULL);
    EC_KEY_set_public_key(key, pub);
    unsigned int mdlen = 0;
    unsigned char *md = (unsigned char *)malloc(MAXLEN);
    EVP_Digest(msg, strlen(msg), md, &mdlen, EVP_sha256(), NULL);
    sig = ECDSA_do_sign(md, mdlen,key);
    if (sig == NULL)
    {
        goto done;
    }
    unsigned char *sigd = (unsigned char  *)malloc(MAXLEN);
    memset(sigd, 0, MAXLEN);
    siglen = i2d_ECDSA_SIG(sig, &sigd);
    base64_encode(sigd - siglen, siglen, out);
    free(sigd-siglen);
done:
    EC_POINT_free(pub);
    EC_KEY_free(key);
    BN_free(priv);
    ECDSA_SIG_free(sig); 
    free(md);
    return siglen;
}


int p256k1_verify_with_base64(const char* msg, const char* pub_data, const char* sig_data)
{
	int ret = 0;
	char pub_raw[MAXLEN];
    char sig_raw[MAXLEN];
    EC_KEY *key = EC_KEY_new_by_curve_name(NID_secp256k1);
	if (key == NULL)
	{
		goto done;
	}
    const EC_GROUP *group = EC_KEY_get0_group(key);
    if (group == NULL)
    {
        goto done;
    }
    BN_CTX *ctx = BN_CTX_new();
    ECDSA_SIG *sig = NULL;
    int pub_len = base64_decode(pub_data, strlen(pub_data), pub_raw);
    EC_POINT *pub = EC_POINT_new(group);
	if (pub == NULL)
	{
		goto done;
	}
    EC_POINT_oct2point(group, pub, pub_raw, pub_len , ctx);
    EC_KEY_set_public_key(key, pub);
    int sig_len = base64_decode(sig_data, strlen(sig_data), sig_raw);
    sig = ECDSA_SIG_new();
    const unsigned char *tmp = sig_raw;
    sig = d2i_ECDSA_SIG(NULL, &tmp, sig_len);
    ret = ECDSA_do_verify(msg, 32, sig, key);

	done:
        EC_POINT_free(pub);
        EC_KEY_free(key);
        ECDSA_SIG_free(sig);
    
    return ret;
}