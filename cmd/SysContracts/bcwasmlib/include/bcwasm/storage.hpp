//
// Created by zhou.yang on 2018/10/22.
//
#pragma once

#include "common.h"
#include "datastream.h"
#include <string>

#ifdef __cplusplus
extern "C" {
#endif
    void setState(const uint8_t* key, size_t klen, const uint8_t *value, size_t vlen);
    size_t getStateSize(const uint8_t* key, size_t klen);
    void getState(const uint8_t* key, size_t klen, uint8_t *value, size_t vlen);
#ifdef __cplusplus
}
#endif


namespace bcwasm {
    /**
     * @brief Set the State object
     * 
     * @tparam KEY Key type
     * @tparam VALUE Value type
     * @param key Key
     * @param value Value
     */
    template <typename KEY, typename VALUE>
    inline void setState(const KEY &key, const VALUE &value) {
        std::vector<char> vecKey(pack_size(key));
        std::vector<char> vecValue(pack_size(value));
        DataStream<char*> keyStream(vecKey.data(), vecKey.size());
        DataStream<char*> valueStream(vecValue.data(), vecValue.size());
        keyStream << key;
        valueStream << value;
        ::setState((const byte*)vecKey.data(), vecKey.size(),  (const byte*)vecValue.data(), vecValue.size());
    }
    /**
     * @brief Get the State object
     * 
     * @tparam KEY Key type
     * @tparam VALUE Value type
     * @param key Key
     * @param value Value
     * @return size_t Get the length of the data
     */
    template <typename KEY, typename VALUE>
    inline size_t getState(const KEY &key, VALUE &value) {
        std::vector<char> vecKey(pack_size(key));
        DataStream<char*> keyStream(vecKey.data(), vecKey.size());
        keyStream << key;
        size_t len = ::getStateSize((const byte*)vecKey.data(), vecKey.size());
        if (len == 0){ return 0; }
        std::vector<char> vecValue(len);
        ::getState((const byte*)vecKey.data(), vecKey.size(), (byte*)vecValue.data(), vecValue.size());

        DataStream<char*> valueStream(vecValue.data(), vecValue.size());
        valueStream >> value;
        return len;
    }

    /**
     * @brief delete State Object
     * 
     * @tparam KEY Key type
     * @param key Key
     */
    template <typename KEY>
    inline void delState(const KEY &key) {
        byte del = 0;
        std::vector<char> vecKey(pack_size(key));
        DataStream<char*> keyStream(vecKey.data(), vecKey.size());
        keyStream << key;
        ::setState((const byte*)vecKey.data(), vecKey.size(),  (const byte*)&del, 0);
    }

}
