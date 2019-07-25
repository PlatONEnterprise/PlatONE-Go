#include "../unittest.hpp"

using namespace bcwasm;

TEST_CASE(convert, floatToStr) {
    {
        float f = 1.000001f;
        char buf[10];
        ::snprintf(buf, sizeof(buf), "%f", f);
        ASSERT(strcmp(buf, "1.000001") == 0);
    }

    {
        float f = 2e-6f / 3.0f;
        char buf[10];
        ::snprintf(buf, sizeof(buf), "%f", f);
        ASSERT(strcmp(buf, "0.000001") == 0);
    }
}

TEST_CASE(convert, doubleToStr) {
    {
        double f = 1.000001;
        char buf[10];
        ::snprintf(buf, sizeof(buf), "%lf", f);
        ASSERT(strcmp(buf, "1.000001") == 0);
    }

    {
        double f = 2e-6 / 3.0;
        char buf[10];
        ::snprintf(buf, sizeof(buf), "%lf", f);
        ASSERT(strcmp(buf, "0.000001") == 0);
    }

}

TEST_CASE(convert, longDoubleToStr) {
    {
        long double f = 1.000001l;
        char buf[10];
        ::snprintf(buf, sizeof(buf), "%Lf", f);
        ASSERT(strcmp(buf, "1.000001") == 0, buf);
    }

    {
        long double f = 2e-6l / 3.0l;
        char buf[10];
        ::snprintf(buf, sizeof(buf), "%Lf", f);
        ASSERT(strcmp(buf, "0.000001") == 0);
    }

}

UNITTEST_MAIN() {
    RUN_TEST(convert, floatToStr);
    RUN_TEST(convert, doubleToStr);
    RUN_TEST(convert, longDoubleToStr);
}
