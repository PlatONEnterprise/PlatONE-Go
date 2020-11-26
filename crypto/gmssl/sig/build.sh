
gcc -c  sig.c  -I../gmssl -L../gmssl/lib -lssl -lcrypto -ldl -lpthread

ar -r libsig.a sig.o

cp libsig.a ../gmssl/lib