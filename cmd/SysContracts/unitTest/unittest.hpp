//
// Created by zhou.yang on 2018/11/21.
//
#include <vector>
#include "bcwasm/print.hpp"

struct TestResult {
    size_t testcases = 0;
    size_t assertions = 0;
    size_t failures = 0;
    bool isContinue = false;
    bool skip = false;
};


#define TEST_CASE(testName, testGroup)\
  class testGroup##testName##Test {\
  public: testGroup##testName##Test (TestResult& testResult):testResult_(testResult) { testResult_.testcases+=1;} \
            void run (); \
  private: TestResult& testResult_; const std::string name_ = #testName; const std::string group_ = #testGroup; \
            } ;\
	void testGroup##testName##Test::run ()

#define RUN_TEST(testName, testGroup) \
    testGroup##testName##Test testGroup##testName##Instance = testGroup##testName##Test(testResult); testGroup##testName##Instance.run();

#define ASSERT(cond, ...) \
if (!testResult_.skip) {\
    if (!(cond)) {\
        if (!testResult_.isContinue) { \
            testResult_.skip = true; \
        } \
        testResult_.failures += 1; \
        bcwasm::println("assertion failed:", #cond, "line:", __LINE__, "file:", __FILE__, "test name:", name_, "group:", group_, ##__VA_ARGS__); \
    } \
    testResult_.assertions += 1; \
    }

#define ASSERT_EQ(A, B, ...) \
    if (!testResult_.skip) {\
    if (!((A) == (B))) {\
        if (!testResult_.isContinue) { \
            testResult_.skip = true; \
        } \
        testResult_.failures += 1; \
        bcwasm::println("assertion failed:", #A, "==", #B, "line:", __LINE__, "file:", __FILE__, "func:", __func__, ##__VA_ARGS__); \
    } \
    testResult_.assertions += 1; \
    }

#define ASSERT_NE(A, B, ...) \
    if (!testResult_.skip) {\
    if (!((A) != (B))) {\
        if (!testResult_.isContinue) { \
            testResult_.skip = true; \
        } \
        testResult_.failures += 1; \
        bcwasm::println("assertion failed:", #A, "!=", #B, "line:", __LINE__, "file:", __FILE__, "func:", __func__, ##__VA_ARGS__); \
    } \
    testResult_.assertions += 1; \
    }



//#define TEST_SUIT() \
//    void testSuit(TestResult &testResult)

#define UNITTEST_MAIN() \
void testSuit(TestResult &testResult);\
int main(int argc, char *argv[]) { \
    TestResult testResult; \
    testResult.isContinue = true; \
    testSuit(testResult); \
    bcwasm::println(testResult.testcases, "tests,", testResult.assertions, "assertions,", testResult.failures, "failures" ); \
    return 0; \
} \
void testSuit(TestResult &testResult)