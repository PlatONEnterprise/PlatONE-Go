#include "stdio_impl.h"
//include <sys/uio.h>
#include <bcwasm/print.h>

size_t __stdio_write(FILE *f, const unsigned char *buf, size_t len)
{
   prints_l((char*)(f->wbase), f->wpos-f->wbase);
   prints_l((void*)buf, len);
   return f->wpos-f->wbase + len;
}
