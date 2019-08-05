//
// Created by zhou.yang on 2018/11/7.
//

#pragma once

#include <boost/endian/conversion.hpp>
#include "RLP.h"
#include <type_traits>
/**
 * @brief Transaction coding operation
 * 
 */
namespace bcwasm {

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d int8_t type
     */
    inline void txEncodeValue(RLPStream &stream, int8_t d){
        d = boost::endian::endian_reverse(d);

        bytesConstRef ref((byte*)&d, sizeof(d));
        stream.append(ref);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d uint16_t type
     */
    inline void txEncodeValue(RLPStream &stream, uint16_t d){
        d = boost::endian::endian_reverse(d);

        bytesConstRef ref((byte*)&d, sizeof(d));
        stream.append(ref);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d int16_t type
     */
    inline void txEncodeValue(RLPStream &stream, int16_t d){
        d = boost::endian::endian_reverse(d);

        bytesConstRef ref((byte*)&d, sizeof(d));
        stream.append(ref);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d uint32_t type
     */
    inline void txEncodeValue(RLPStream &stream, uint32_t d){
        d = boost::endian::endian_reverse(d);

        bytesConstRef ref((byte*)&d, sizeof(d));
        stream.append(ref);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d int32_t type
     */
    inline void txEncodeValue(RLPStream &stream, int32_t d){
        d = boost::endian::endian_reverse(d);

        bytesConstRef ref((byte*)&d, sizeof(d));
        stream.append(ref);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d int type
     */
    inline void txEncodeValue(RLPStream &stream, int d){
        d = boost::endian::endian_reverse((int32_t)d);

        bytesConstRef ref((byte*)&d, sizeof(d));
        stream.append(ref);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d uint64_t type
     */
    inline void txEncodeValue(RLPStream &stream, uint64_t d){
        d = boost::endian::endian_reverse(d);

        bytesConstRef ref((byte*)&d, sizeof(d));
        stream.append(ref);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d int64_t type
     */
    inline void txEncodeValue(RLPStream &stream, int64_t d){
        d = boost::endian::endian_reverse(d);

        bytesConstRef ref((byte*)&d, sizeof(d));
        stream.append(ref);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d std::string type
     */
    inline void txEncodeValue(RLPStream &stream, const std::string d){
        stream.append(d);
    }

    /**
     * @brief Specified type encoding
     * 
     * @param stream RLP stream
     * @param d Char pointer type
     */
    inline void txEncodeValue(RLPStream &stream, const char *d){
        stream.append(std::string(d));
    }

    /**
     * @brief Empty implementation
     * 
     * @param stream 
     */
    inline void txEncodeValue(RLPStream &stream){
    }

    template <typename T>
    inline void txEncode(RLPStream &stream,T arg){
        txEncodeValue(stream,arg);
    }

    /**
     * @brief Serialize to RLPStream
     * 
     * @tparam Arg Starting element type
     * @tparam Args Variable parameter type
     * @param stream RLP stream
     * @param a Starting parameter
     * @param args Variable parameter
     */
    template<typename Arg, typename... Args>
    void txEncode(RLPStream &stream, Arg&& a, Args&&... args ) {
        txEncodeValue(stream, a);
        txEncode(stream, args...);
    }
}
