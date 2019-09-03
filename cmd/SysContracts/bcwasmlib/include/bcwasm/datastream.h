#pragma once

#include "varint.hpp"
#include "fixedhash.hpp"
#include "assert.h"

#include <array>
#include <set>
#include <map>
#include <string>
#include <tuple>


#include <boost/pfr.hpp>


namespace bcwasm {

/**
 * @defgroup DataStream Data Stream
 * @brief Defines data stream for reading and writing data in the form of bytes
 * @ingroup serialize
 * @{
 */

/**
 *  %A data stream for reading and writing data in the form of bytes
 *
 *  @brief %A data stream for reading and writing data in the form of bytes.
 *  @tparam T - Type of the DataStream buffer
 */
    template<typename T>
    class DataStream {
    public:
        /**
         * Construct a new DataStream object given the size of the buffer and start position of the buffer
         *
         * @brief Construct a new DataStream object
         * @param start - The start position of the buffer
         * @param s - The size of the buffer
         */
        DataStream( T start, size_t s )
                :_start(start),_pos(start),_end(start+s){}

        /**
         *  Skips a specified number of bytes from this stream
         *
         *  @brief Skips a specific number of bytes from this stream
         *  @param s - The number of bytes to skip
         */
        inline void skip( size_t s ){ _pos += s; }

        /**
         *  Reads a specified number of bytes from the stream into a buffer
         *
         *  @brief Reads a specified number of bytes from this stream into a buffer
         *  @param d - The pointer to the destination buffer
         *  @param s - the number of bytes to read
         *  @return true
         */
        inline bool read( char* d, size_t s ) {
            BCWasmAssert( size_t(_end - _pos) >= (size_t)s, "size_t(_end - _pos):", size_t(_end - _pos), "s:", (size_t)s);
            memcpy( d, _pos, s );
            _pos += s;
            return true;
        }

        /**
         *  Writes a specified number of bytes into the stream from a buffer
         *
         *  @brief Writes a specified number of bytes into the stream from a buffer
         *  @param d - The pointer to the source buffer
         *  @param s - The number of bytes to write
         *  @return true
         */
        inline bool write( const char* d, size_t s ) {
            BCWasmAssert( _end - _pos >= (int32_t)s, "write" );
            memcpy( (void*)_pos, d, s );
            _pos += s;
            return true;
        }

        /**
         *  Writes a byte into the stream
         *
         *  @brief Writes a byte into the stream
         *  @param c byte to write
         *  @return true
         */
        inline bool put(char c) {
            BCWasmAssert( _pos < _end, "put" );
            *_pos = c;
            ++_pos;
            return true;
        }

        /**
         *  Reads a byte from the stream
         *
         *  @brief Reads a byte from the stream
         *  @param c - The reference to destination byte
         *  @return true
         */
        inline bool get( unsigned char& c ) { return get( *(char*)&c ); }

        /**
         *  Reads a byte from the stream
         *
         *  @brief Reads a byte from the stream
         *  @param c - The reference to destination byte
         *  @return true
         */
        inline bool get( char& c )
        {
            BCWasmAssert( _pos < _end, "get" );
            c = *_pos;
            ++_pos;
            return true;
        }

        /**
         *  Retrieves the current position of the stream
         *
         *  @brief Retrieves the current position of the stream
         *  @return T - The current position of the stream
         */
        T pos()const { return _pos; }
        inline bool valid()const { return _pos <= _end && _pos >= _start;  }

        /**
         *  Sets the position within the current stream
         *
         *  @brief Sets the position within the current stream
         *  @param p - The offset relative to the origin
         *  @return true if p is within the range
         *  @return false if p is not within the rawnge
         */
        inline bool seekp(size_t p) { _pos = _start + p; return _pos <= _end; }

        /**
         *  Gets the position within the current stream
         *
         *  @brief Gets the position within the current stream
         *  @return p - The position within the current stream
         */
        inline size_t tellp()const      { return size_t(_pos - _start); }

        /**
         *  Returns the number of remaining bytes that can be read/skipped
         *
         *  @brief Returns the number of remaining bytes that can be read/skipped
         *  @return size_t - The number of remaining bytes
         */
        inline size_t remaining()const  { return _end - _pos; }
    private:
        /**
         * The start position of the buffer
         *
         * @brief The start position of the buffer
         */
        T _start;
        /**
         * The current position of the buffer
         *
         * @brief The current position of the buffer
         */
        T _pos;
        /**
         * The end position of the buffer
         *
         * @brief The end position of the buffer
         */
        T _end;
    };

/**
 * @brief Specialization of DataStream used to help determine the final size of a serialized value.
 * Specialization of DataStream used to help determine the final size of a serialized value
 */
    template<>
    class DataStream<size_t> {
    public:
        /**
         * Construct a new specialized DataStream object given the initial size
         *
         * @brief Construct a new specialized DataStream object
         * @param init_size - The initial size
         */
        DataStream( size_t init_size = 0):_size(init_size){}

        /**
         *  Increment the size by s. This behaves the same as write( const char* ,size_t s ).
         *
         *  @brief Increase the size by s
         *  @param s - The amount of size to increase
         *  @return true
         */
        inline bool     skip( size_t s )                 { _size += s; return true;  }

        /**
         *  Increment the size by s. This behaves the same as skip( size_t s )
         *
         *  @brief Increase the size by s
         *  @param s - The amount of size to increase
         *  @return true
         */
        inline bool     write( const char* ,size_t s )  { _size += s; return true;  }

        /**
         *  Increment the size by one
         *
         *  @brief Increase the size by one
         *  @return true
         */
        inline bool     put(char )                      { ++_size; return  true;    }

        /**
         *  Check validity. It's always valid
         *
         *  @brief Check validity
         *  @return true
         */
        inline bool     valid()const                     { return true;              }

        /**
         * Set new size
         *
         * @brief Set new size
         * @param p - The new size
         * @return true
         */
        inline bool     seekp(size_t p)                  { _size = p;  return true;  }

        /**
         * Get the size
         *
         * @brief Get the size
         * @return size_t - The size
         */
        inline size_t   tellp()const                     { return _size;             }

        /**
         * Always returns 0
         *
         * @brief Always returns 0
         * @return size_t - 0
         */
        inline size_t   remaining()const                 { return 0;                 }
    private:
        /**
         * The size used to determine the final size of a serialized value.
         *
         * @brief The size used to determine the final size of a serialized value.
         */
        size_t _size;
    };



    template <typename Stream, unsigned N>
    inline DataStream<Stream>& operator <<(DataStream<Stream> &ds, const FixedHash<N> &d) {
        ds << unsigned_int( N );
        ds.write((const char*)d.data(), N);
        return ds;
    }

    template <typename Stream, unsigned N>
    inline DataStream<Stream>& operator >> (DataStream<Stream> &ds, FixedHash<N> &d) {
        unsigned_int s;
        ds >> s;
        BCWasmAssert(s.value == N, "s.value:", s.value, "N:", N);
        ds.read((char *)d.data(), N);
        return ds;
    }






/**
 *  Serialize a bool into a stream
 *
 *  @brief  Serialize a bool into a stream
 *  @param ds - The stream to read
 *  @param d - The value to serialize
 *  @tparam Stream - Type of DataStream buffer
 *  @return DataStream<Stream>& - Reference to the DataStream
 */
    template<typename Stream>
    inline DataStream<Stream>& operator<<(DataStream<Stream>& ds, const bool& d) {
        return ds << uint8_t(d);
    }

/**
 *  Deserialize a bool from a stream
 *
 *  @brief Deserialize a bool
 *  @param ds - The stream to read
 *  @param d - The destination for deserialized value
 *  @tparam Stream - Type of DataStream buffer
 *  @return DataStream<Stream>& - Reference to the DataStream
 */
    template<typename Stream>
    inline DataStream<Stream>& operator>>(DataStream<Stream>& ds, bool& d) {
        uint8_t t;
        ds >> t;
        d = t;
        return ds;
    }





/**
 *  Serialize a string into a stream
 *
 *  @brief Serialize a string
 *  @param ds - The stream to write
 *  @param v - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS>
    DS& operator << ( DS& ds, const std::string& v ) {
        ds << unsigned_int( v.size() );
        if (v.size())
            ds.write(v.data(), v.size());
        return ds;
    }

/**
 *  Deserialize a string from a stream
 *
 *  @brief Deserialize a string
 *  @param ds - The stream to read
 *  @param v - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS>
    DS& operator >> ( DS& ds, std::string& v ) {
        std::vector<char> tmp;
        ds >> tmp;
        if( tmp.size() )
            v = std::string(tmp.data(),tmp.data()+tmp.size());
        else
            v = std::string();
        return ds;
    }

/**
 *  Serialize a fixed size array into a stream
 *
 *  @brief Serialize a fixed size array
 *  @param ds - The stream to write
 *  @param v - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the object contained in the array
 *  @tparam N - Size of the array
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::size_t N>
    DS& operator << ( DS& ds, const std::array<T,N>& v ) {
        for( const auto& i : v )
            ds << i;
        return ds;
    }


/**
 *  Deserialize a fixed size array from a stream
 *
 *  @brief Deserialize a fixed size array
 *  @param ds - The stream to read
 *  @param v - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the object contained in the array
 *  @tparam N - Size of the array
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::size_t N>
    DS& operator >> ( DS& ds, std::array<T,N>& v ) {
        for( auto& i : v )
            ds >> i;
        return ds;
    }

    /**
     *  Serialize a u256 type into a stream
     *
     *  @brief Serialize a u256 type
     *  @param ds - The stream to write
     *  @param v - The value to serialize
     *  @tparam DS - Type of DataStream
     *  @return DS& - Reference to the DataStream
     */
    template<typename DS>
    DS& operator << (DS& ds, const u256& v) {
        bytes bs(32);
        toBigEndian(v, bs);
        auto hex = toHex(bs.begin(), bs.end(), std::string());
        return ds << hex;
    }

    /**
     *  Deserialize a u256 from a stream
     *
     *  @brief Deserialize a string from a stream
     *  @param ds - The stream to read
     *  @param v - The destination for deserialized value
     *  @param DS - Type of DataStream
     *  @return DS& - Reference to the DataStream
     */
    template<typename DS>
    DS& operator >> (DS& ds, u256& v) {
        std::string hex;
        ds >> hex;
        auto bs = fromHex(hex);
        v = fromBigEndian<u256>(bs);
        return ds;
    }


    namespace _datastream_detail {
        /**
         * Check if type T is a pointer
         *
         * @brief Check if type T is a pointer
         * @tparam T - The type to be checked
         * @return true if T is a pointer
         * @return false otherwise
         */
        template<typename T>
        constexpr bool is_pointer() {
            return std::is_pointer<T>::value ||
                   std::is_null_pointer<T>::value ||
                   std::is_member_pointer<T>::value;
        }

        /**
         * Check if type T is a primitive type
         *
         * @brief Check if type T is a primitive type
         * @tparam T - The type to be checked
         * @return true if T is a primitive type
         * @return false otherwise
         */
        template<typename T>
        constexpr bool is_primitive() {
            return std::is_arithmetic<T>::value ||
                   std::is_enum<T>::value;
        }
    }

/**
 *  Pointer should not be serialized, so this function will always throws an error
 *
 *  @brief Deserialize a a pointer
 *  @param ds - The stream to read
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the pointer
 *  @return DS& - Reference to the DataStream
 *  @post Throw an exception if it is a pointer
 */
    template<typename DS, typename T, std::enable_if_t<_datastream_detail::is_pointer<T>()>* = nullptr>
    DS& operator >> ( DS& ds, T ) {
        static_assert(!_datastream_detail::is_pointer<T>(), "Pointers should not be serialized" );
        return ds;
    }

/**
 *  Serialize a fixed size array of non-primitive and non-pointer type
 *
 *  @brief Serialize a fixed size array of non-primitive and non-pointer type
 *  @param ds - The stream to write
 *  @param v - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the pointer
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::size_t N,
            std::enable_if_t<!_datastream_detail::is_primitive<T>() &&
                             !_datastream_detail::is_pointer<T>()>* = nullptr>
    DS& operator << ( DS& ds, const T (&v)[N] ) {
        ds << unsigned_int( N );
        for( uint32_t i = 0; i < N; ++i )
            ds << v[i];
        return ds;
    }

/**
 *  Serialize a fixed size array of non-primitive type
 *
 *  @brief Serialize a fixed size array of non-primitive type
 *  @param ds - The stream to write
 *  @param v - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the pointer
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::size_t N,
            std::enable_if_t<_datastream_detail::is_primitive<T>()>* = nullptr>
    DS& operator << ( DS& ds, const T (&v)[N] ) {
        ds << unsigned_int( N );
        ds.write((char*)&v[0], sizeof(v));
        return ds;
    }

/**
 *  Deserialize a fixed size array of non-primitive and non-pointer type
 *
 *  @brief Deserialize a fixed size array of non-primitive and non-pointer type
 *  @param ds - The stream to read
 *  @param v - The destination for deserialized value
 *  @tparam T - Type of the object contained in the array
 *  @tparam N - Size of the array
 *  @tparam DS - Type of DataStream
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::size_t N,
            std::enable_if_t<!_datastream_detail::is_primitive<T>() &&
                             !_datastream_detail::is_pointer<T>()>* = nullptr>
    DS& operator >> ( DS& ds, T (&v)[N] ) {
        unsigned_int s;
        ds >> s;
        BCWasmAssert( N == s.value, "T[] size and unpacked size don't match");
        for( uint32_t i = 0; i < N; ++i )
            ds >> v[i];
        return ds;
    }

/**
 *  Deserialize a fixed size array of non-primitive type
 *
 *  @brief Deserialize a fixed size array of non-primitive type
 *  @param ds - The stream to read
 *  @param v - The destination for deserialized value
 *  @tparam T - Type of the object contained in the array
 *  @tparam N - Size of the array
 *  @tparam DS - Type of DataStream
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::size_t N,
            std::enable_if_t<_datastream_detail::is_primitive<T>()>* = nullptr>
    DS& operator >> ( DS& ds, T (&v)[N] ) {
        unsigned_int s;
        ds >> s;
        BCWasmAssert( N == s.value, "T[] size and unpacked size don't match");
        ds.read((char*)&v[0], sizeof(v));
        return ds;
    }

/**
 *  Serialize a vector of char
 *
 *  @brief Serialize a vector of char
 *  @param ds - The stream to write
 *  @param v - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS>
    DS& operator << ( DS& ds, const std::vector<char>& v ) {
        ds << unsigned_int( v.size() );
        ds.write( v.data(), v.size() );
        return ds;
    }

    template<typename DS>
    DS& operator << ( DS& ds, const std::vector<bool>& v ) {
        size_t size = v.size() / 8 + (v.size() % 8 ? 1 : 0);
        std::string data;
        data.reserve(size);
        size_t num = 0;
        while (num < v.size()) {
            uint8_t item = 0;
            size_t shift = v.size() - num >= 8 ? 8 : v.size() - num;

            for (size_t i = shift; i > 0; i--) {
                if (i != shift) {item <<= 1;}
                if (v[num + i-1]) {
                    item |= 0x01;
                }

            }
            data.append(1, (char)item);
            num += 8;
        }
        ds << unsigned_int(v.size());
        ds << data;
        return ds;
    }

/**
 *  Serialize a std::vector
 *
 *  @brief Serialize a std::vector
 *  @param ds - The stream to write
 *  @param v - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the object contained in the std::vector
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T>
    DS& operator << ( DS& ds, const std::vector<T>& v ) {
        ds << unsigned_int( v.size() );
        for( const auto& i : v )
            ds << i;
        return ds;
    }

/**
 *  Deserialize a std::vector of char
 *
 *  @brief Deserialize a std::vector of char
 *  @param ds - The stream to read
 *  @param v - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS>
    DS& operator >> ( DS& ds, std::vector<char>& v ) {
        unsigned_int s;
        ds >> s;
        v.resize( s.value );
        ds.read( v.data(), v.size() );
        return ds;
    }

    template<typename DS>
    DS& operator >> ( DS& ds, std::vector<bool>& v ) {
        unsigned_int s = 0;
        std::string data;
        ds >> s;
        ds >> data;
        size_t size = s.value;// / 8 + (s.value % 8 ? 1 : 0);
        v.reserve(s.value);
        size_t num = 0;
        while (num < data.length()) {
            uint8_t item = data[num];
            size_t shift = size >= (num+1) *8  ? 8 : size -num * 8;
            for (size_t i = 0; i < shift; i++) {
                if (i !=0) { item >>= 1;}
                v.push_back(item & 0x01);

            }
            num += 1;
        }
        return ds;
    }
/**
 *  Deserialize a std::vector
 *
 *  @brief Deserialize a std::vector
 *  @param ds - The stream to read
 *  @param v - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the object contained in the std::vector
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T>
    DS& operator >> ( DS& ds, std::vector<T>& v ) {
        unsigned_int s;
        ds >> s;
        v.resize(s.value);
        for( auto& i : v )
            ds >> i;
        return ds;
    }

    template<typename DS, typename T>
    DS& operator << ( DS& ds, const std::set<T>& s ) {
        ds << unsigned_int( s.size() );
        for( const auto& i : s ) {
            ds << i;
        }
        return ds;
    }

    template<typename DS, typename T>
    DS& operator >> ( DS& ds, std::set<T>& s ) {
        s.clear();
        unsigned_int sz; ds >> sz;

        for( uint32_t i = 0; i < sz.value; ++i ) {
            T v;
            ds >> v;
            s.emplace( std::move(v) );
        }
        return ds;
    }

/**
 *  Serialize a map
 *
 *  @brief Serialize a map
 *  @param ds - The stream to write
 *  @param m - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam K - Type of the key contained in the map
 *  @tparam V - Type of the value contained in the map
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename K, typename V>
    DS& operator << ( DS& ds, const std::map<K,V>& m ) {
        ds << unsigned_int( m.size() );
        for( const auto& i : m ) {
            ds << i.first << i.second;
        }
        return ds;
    }

/**
 *  Deserialize a map
 *
 *  @brief Deserialize a map
 *  @param ds - The stream to read
 *  @param m - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @tparam K - Type of the key contained in the map
 *  @tparam V - Type of the value contained in the map
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename K, typename V>
    DS& operator >> ( DS& ds, std::map<K,V>& m ) {
        m.clear();
        unsigned_int s; ds >> s;

        for (uint32_t i = 0; i < s.value; ++i) {
            K k; V v;
            ds >> k >> v;
            m.emplace( std::move(k), std::move(v) );
        }
        return ds;
    }

//    template<typename DS, typename T>
//    DS& operator << ( DS& ds, const boost::container::flat_set<T>& s ) {
//        ds << unsigned_int( s.size() );
//        for( const auto& i : s ) {
//            ds << i;
//        }
//        return ds;
//    }
//
//    template<typename DS, typename T>
//    DS& operator >> ( DS& ds, boost::container::flat_set<T>& s ) {
//        s.clear();
//        unsigned_int sz; ds >> sz;
//
//        for( uint32_t i = 0; i < sz.value; ++i ) {
//            T v;
//            ds >> v;
//            s.emplace( std::move(v) );
//        }
//        return ds;
//    }


/**
 *  Serialize a flat map
 *
 *  @brief Serialize a flat map
 *  @param ds - The stream to write
 *  @param m - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam K - Type of the key contained in the flat map
 *  @tparam V - Type of the value contained in the flat map
 *  @return DS& - Reference to the DataStream
 */
//    template<typename DS, typename K, typename V>
//    DS& operator<<( DS& ds, const boost::container::flat_map<K,V>& m ) {
//        ds << unsigned_int( m.size() );
//        for( const auto& i : m )
//            ds << i.first << i.second;
//        return ds;
//    }

/**
 *  Deserialize a flat map
 *
 *  @brief Deserialize a flat map
 *  @param ds - The stream to read
 *  @param m - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @tparam K - Type of the key contained in the flat map
 *  @tparam V - Type of the value contained in the flat map
 *  @return DS& - Reference to the DataStream
 */
//    template<typename DS, typename K, typename V>
//    DS& operator>>( DS& ds, boost::container::flat_map<K,V>& m ) {
//        m.clear();
//        unsigned_int s; ds >> s;
//
//        for( uint32_t i = 0; i < s.value; ++i ) {
//            K k; V v;
//            ds >> k >> v;
//            m.emplace( std::move(k), std::move(v) );
//        }
//        return ds;
//    }

/**
 *  Serialize a tuple
 *
 *  @brief Serialize a tuple
 *  @param ds - The stream to write
 *  @param t - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam Args - Type of the objects contained in the tuple
 *  @return DS& - Reference to the DataStream
 */


    template<typename DS, typename Tuple, typename Func, size_t N>
    struct __tuple_processor
    {
        inline static void __tuple_process(DS& ds, Tuple &t, Func func)
        {
            __tuple_processor<DS, Tuple, Func, N - 1>::__tuple_process(ds, t, func);
            func(ds, std::get<N-1>(t));
        }
    };

    template<typename DS, typename Tuple, typename Func>
    struct __tuple_processor<DS, Tuple, Func, 1>
    {
        inline static void __tuple_process(DS& ds, Tuple &t, Func func)
        {
            func(ds, std::get<0>(t));
        }
    };

    template<typename DS, typename Func, typename... Args>
    void tuple_for_each(DS& ds, std::tuple<Args...> &t, Func func)
    {
        __tuple_processor<decltype(ds), decltype(t), Func, sizeof...(Args)>::__tuple_process(ds, t, func);
    }

    template<typename DS, typename... Args>
    DS& operator<<( DS& ds, const std::tuple<Args...>& t ) {
        tuple_for_each(ds, const_cast<std::tuple<Args...>&>(t), [](DS &ds, const auto &v){
            ds << v;
        });
        return ds;
    }

/**
 *  Deserialize a tuple
 *
 *  @brief Deserialize a tuple
 *  @param ds - The stream to read
 *  @param t - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @tparam Args - Type of the objects contained in the tuple
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename... Args>
    DS& operator>>( DS& ds, std::tuple<Args...>& t ) {
        tuple_for_each(ds, t, [](DS &ds, auto &v){
            ds >> v;
        });
        return ds;
    }

/**
 *  Serialize a class
 *
 *  @brief Serialize a class
 *  @param ds - The stream to write
 *  @param v - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of class
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::enable_if_t<std::is_class<T>::value>* = nullptr>
    DS& operator<<( DS& ds, const T& v ) {
        boost::pfr::for_each_field(v, [&](const auto& field) {
            ds << field;
        });
        return ds;
    }

/**
 *  Deserialize a class
 *
 *  @brief Deserialize a class
 *  @param ds - The stream to read
 *  @param v - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of class
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::enable_if_t<std::is_class<T>::value>* = nullptr>
    DS& operator>>( DS& ds, T& v ) {
        boost::pfr::for_each_field(v, [&](auto& field) {
            ds >> field;
        });
        return ds;
    }

/**
 *  Serialize a primitive type
 *
 *  @brief Serialize a primitive type
 *  @param ds - The stream to write
 *  @param v - The value to serialize
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the primitive type
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::enable_if_t<_datastream_detail::is_primitive<T>()>* = nullptr>
    DS& operator<<( DS& ds, const T& v ) {
        ds.write( (const char*)&v, sizeof(T) );
        return ds;
    }

/**
 *  Deserialize a primitive type
 *
 *  @brief Deserialize a primitive type
 *  @param ds - The stream to read
 *  @param v - The destination for deserialized value
 *  @tparam DS - Type of DataStream
 *  @tparam T - Type of the primitive type
 *  @return DS& - Reference to the DataStream
 */
    template<typename DS, typename T, std::enable_if_t<_datastream_detail::is_primitive<T>()>* = nullptr>
    DS& operator>>( DS& ds, T& v ) {
        ds.read( (char*)&v, sizeof(T) );
        return ds;
    }


/**
 * Unpack data inside a fixed size buffer as T
 *
 * @brief Unpack data inside a fixed size buffer as T
 * @tparam T - Type of the unpacked data
 * @param buffer - Pointer to the buffer
 * @param len - Length of the buffer
 * @return T - The unpacked data
 */
    template<typename T>
    T unpack( const char* buffer, size_t len ) {
        T result;
        DataStream<const char*> ds(buffer,len);
        ds >> result;
        return result;
    }

/**
 * Unpack data inside a variable size buffer as T
 *
 * @brief Unpack data inside a variable size buffer as T
 * @tparam T - Type of the unpacked data
 * @param bytes - Buffer
 * @return T - The unpacked data
 */
    template<typename T>
    T unpack( const std::vector<char>& bytes ) {
        return unpack<T>( bytes.data(), bytes.size() );
    }

/**
 * Get the size of the packed data
 *
 * @brief Get the size of the packed data
 * @tparam T - Type of the data to be packed
 * @param value - Data to be packed
 * @return size_t - Size of the packed data
 */
    template<typename T>
    size_t pack_size( const T& value ) {
        DataStream<size_t> ps;
        ps << value;
        return ps.tellp();
    }

/**
 * Get packed data
 *
 * @brief Get packed data
 * @tparam T - Type of the data to be packed
 * @param value - Data to be packed
 * @return bytes - The packed data
 */
    template<typename T>
    bytes pack( const T& value ) {
        bytes result;
        result.resize(pack_size(value));

        DataStream<byte*> ds( result.data(), result.size() );
        ds << value;
        return result;
    }


/// @} DataStream
}
