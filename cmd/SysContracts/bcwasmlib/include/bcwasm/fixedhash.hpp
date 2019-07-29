#pragma once

#include <vector>
#include <string>
#include <array>
#include <initializer_list>
#include "common.h"

namespace bcwasm {

    template <unsigned N>
    class FixedHash{
    public:
        /// Construct an empty hash.
        FixedHash() { m_data.fill(0); }

        template <unsigned M> explicit FixedHash(FixedHash<M> const& h){
            m_data.fill(0);
            unsigned c = std::min(M, N);
            for (unsigned i = 0; i < c; ++i){
                m_data[i] = h[i];
            }
        }
        template <typename T>
        FixedHash(const std::initializer_list<T> &l) {
            m_data.fill(0);
            unsigned i= 0;
            for (auto it = l.begin(); it != l.end() && i < N; ++it, ++i) {
                m_data[i] = *it;
            }
        }

        explicit FixedHash(unsigned _u) { toBigEndian(_u, m_data); }

        explicit FixedHash(const byte *h, size_t len){
            memcpy(m_data.data(), h, N);
        }
        explicit FixedHash(const bytes &&b) {
            std::copy(b.begin(), b.end(), m_data.begin());
        }

        explicit FixedHash(const std::string &s, bool isHex = true): FixedHash(isHex ? fromHex(s) : asBytes(s)){}

        std::string toString() const {
            return toHex(m_data);
        }
        byte const* data() const { return m_data.data(); }
        byte* data() { return m_data.data(); }
        bytesRef ref() { return bytesRef(m_data.data(), N); }
        /// @returns true iff this is the empty hash.
        explicit operator bool() const { return std::any_of(m_data.begin(), m_data.end(), [](byte _b) { return _b != 0; }); }

        // The obvious comparison operators.
        bool operator==(FixedHash const& c) const { return m_data == c.m_data; }
        bool operator!=(FixedHash const& c) const { return m_data != c.m_data; }
        bool operator<(FixedHash const& c) const { for (unsigned i = 0; i < N; ++i) if (m_data[i] < c.m_data[i]) return true; else if (m_data[i] > c.m_data[i]) return false; return false; }
        bool operator>=(FixedHash const& c) const { return !operator<(c); }
        bool operator<=(FixedHash const& c) const { return operator==(c) || operator<(c); }
        bool operator>(FixedHash const& c) const { return !operator<=(c); }

        // The obvious binary operators.
        FixedHash& operator^=(FixedHash const& c) { for (unsigned i = 0; i < N; ++i) m_data[i] ^= c.m_data[i]; return *this; }
        FixedHash operator^(FixedHash const& c) const { return FixedHash(*this) ^= c; }
        FixedHash& operator|=(FixedHash const& c) { for (unsigned i = 0; i < N; ++i) m_data[i] |= c.m_data[i]; return *this; }
        FixedHash operator|(FixedHash const& c) const { return FixedHash(*this) |= c; }
        FixedHash& operator&=(FixedHash const& c) { for (unsigned i = 0; i < N; ++i) m_data[i] &= c.m_data[i]; return *this; }
        FixedHash operator&(FixedHash const& c) const { return FixedHash(*this) &= c; }
        FixedHash operator~() const { FixedHash ret; for (unsigned i = 0; i < N; ++i) ret[i] = ~m_data[i]; return ret; }

        /// @returns a particular byte from the hash.
        byte& operator[](unsigned _i) { return m_data[_i]; }
        /// @returns a particular byte from the hash.
        byte operator[](unsigned _i) const { return m_data[_i]; }

        // Big-endian increment.
        FixedHash& operator++() { for (unsigned i = size(); i > 0 && !++m_data[--i]; ) {} return *this; }

        /// @returns true if all one-bits in @a _c are set in this object.
        bool contains(FixedHash const& _c) const { return (*this & _c) == _c; }

        const size_t size() const { return N; }
    private:
        std::array<byte, N> m_data;
    };

    using h256 = FixedHash<32>;
    using h160 = FixedHash<20>;
    using h128 = FixedHash<16>;
    using h64 = FixedHash<8>;
    using Address = FixedHash<20>;
}
