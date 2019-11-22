#gcc -fPIC -shared -o sm2.so sm2.c -I/home/panchen/Work/C++/smtest/include  -L/home/panchen/Work/C++/smtest/lib -lssl -lcrypto -ldl -lpthread

gcc -c  sm2.c -I./include  -L../ -lssl -lcrypto -ldl -lpthread

