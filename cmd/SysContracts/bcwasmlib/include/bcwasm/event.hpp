#pragma once

#include "print.hpp"
#include "common.h"
#include "RLP.h"

#define ARG_COUNT_P1_(\
  _1, _2, _3, _4, _5, _6, _7, _8, _9, _10, N, ...) \
  N

#define ARG_COUNT_P2_ \
  9, 8, 7, 6, 5, 4, 3, 2, 1, 0

#define ARG_COUNT_(...) \
  ARG_COUNT_P1_(__VA_ARGS__)

#define ARG_COUNT(...) \
  ARG_COUNT_(_, ##__VA_ARGS__, ARG_COUNT_P2_)

#define M_CAT(a, b) M_CAT_(a, b)
#define M_CAT_(a, b) a##b

#define VA_F(...) \
  M_CAT(func, ARG_COUNT(__VA_ARGS__)) (__VA_ARGS__)

#define PA_F(...) \
  M_CAT(params, ARG_COUNT(__VA_ARGS__)) (__VA_ARGS__)

#define params1(a)   arg1
#define params2(a, ...)  arg2, params1(__VA_ARGS__)
#define params3(a, ...)  arg3, params2(__VA_ARGS__)
#define params4(a, ...)  arg4, params3(__VA_ARGS__)
#define params5(a, ...)  arg5, params4(__VA_ARGS__)
#define params6(a, ...)  arg6, params5(__VA_ARGS__)
#define params7(a, ...)  arg7, params6(__VA_ARGS__)
#define params8(a, ...)  arg8, params7(__VA_ARGS__)
#define params9(a, ...)  arg9, params8(__VA_ARGS__)
#define params10(a, ...) arg10, params9(__VA_ARGS__)

#define func1(a)  a arg1
#define func2(a, ...) a arg2, func1(__VA_ARGS__)
#define func3(a, ...) a arg3, func2(__VA_ARGS__)
#define func4(a, ...) a arg4, func3(__VA_ARGS__)
#define func5(a, ...) a arg5, func4(__VA_ARGS__)
#define func6(a, ...) a arg6, func5(__VA_ARGS__)
#define func7(a, ...) a arg7, func6(__VA_ARGS__)
#define func8(a, ...) a arg8, func7(__VA_ARGS__)
#define func9(a, ...) a arg9, func8(__VA_ARGS__)
#define func10(a, ...) a arg10, func9(__VA_ARGS__)


/**
 * @brief Defining events
 * 
 */
#define BCWASM_EVENT(NAME, ...) \
    void M_CAT(EVENT, NAME)(VA_F(__VA_ARGS__)) { \
        bcwasm::emitEvent(#NAME, PA_F(__VA_ARGS__)); \
    }

/**
 * @brief trigger event
 * 
 */
#define BCWASM_EMIT_EVENT(NAME, ...) \
    M_CAT(EVENT, NAME)(__VA_ARGS__)


#ifdef __cplusplus
extern "C" {
#endif
    void emitEvent(const char *topic, size_t topicLen, const uint8_t *data, size_t dataLen);
#ifdef __cplusplus
}
#endif

namespace bcwasm {

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num uint64_t type
     */
    void event(RLPStream &stream, uint64_t num) {
        stream << num;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num uint32_t type
     */
    void event(RLPStream &stream, uint32_t num) {
        stream << num;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num uint16_t type
     */
    void event(RLPStream &stream, uint16_t num) {
        stream << num;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num uint8_t type
     */
    void event(RLPStream &stream, uint8_t num) {
        stream << num;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num int64_t type
     */
    void event(RLPStream &stream, int64_t num) {
        stream << (uint64_t)num;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num int32_t type
     */
    void event(RLPStream &stream, int32_t num) {
        stream << (uint64_t)num;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num int16_t type
     */
    void event(RLPStream &stream, int16_t num) {
        stream << (uint64_t)num;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num int8_t type
     */
    void event(RLPStream &stream, int8_t num) {
        stream << (uint64_t)num;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num std::string type
     */
    void event(RLPStream &stream, const std::string &s){
        stream << s;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num const char* type
     */
    void event(RLPStream &stream, const char* topic) {
        stream << topic;
    }

    /**
     * @brief Specify event type field serialization
     * 
     * @param stream RLP stream
     * @param num char* type
     */
    void event(RLPStream &stream, char* topic) {
        stream << topic;
    }

    /**
     * @brief Empty parameter
     * 
     * @param stream 
     */
    void event(RLPStream &stream){

    }

    /**
     * @brief 
     * 
     * @tparam Arg Starting parameter type
     * @tparam Args Variable parameter type
     * @param stream RLP stream
     * @param a Starting parameter
     * @param args Variable parameter
     */
    template<typename Arg, typename... Args>
    void event(RLPStream &stream, Arg &&a, Args &&... args) {
        event(stream, a);
        event(stream, args...);
    }

    /**
     * @brief Serialization parameter
     * 
     * @tparam Args Event parameter type
     * @param topic Topic name
     * @param args Event parameter
     */
    template<typename... Args>
    void emitEvent(const std::string &topic, Args &&... args) {
        bytes data;
        RLPStream stream(sizeof...(args));
        event(stream, args...);
        const bytes& rlpData = stream.out();
        ::emitEvent(topic.data(), topic.length(),rlpData.data(), rlpData.size());
    }
}
