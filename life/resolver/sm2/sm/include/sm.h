
int base64_encode(const char *in_str, int in_len, char *out_str);
int base64_decode(const char *in_str, int in_len, char *out_str);

int sm_generate_sig(const char *msg, const char *privkey, char *output);
int sm_verify_sig(const char *msg, const char *pubkey,const char *signature);
int sm2_sign_with_base64(const char *msg, const char *userid, const char *privkey, char *out);
int sm2_verify_with_base64(const char* msg, const char* userid, const char* pub_data, const char* sig_data);

