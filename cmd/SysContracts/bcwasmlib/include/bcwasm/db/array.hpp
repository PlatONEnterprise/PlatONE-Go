//
// Created by zhou.yang on 2018/11/14.
//

#pragma once

#include <map>
#include "bcwasm/storage.hpp"

namespace bcwasm {
namespace db {
    /**
     * @brief Implement Array function
     * 
     * @tparam *Name Array name, in the same contract, the name should be unique
     * @tparam Key Array element type
     * @tparam Size Array length
     */
    template <const char *Name, typename Key, unsigned Size>
    class Array {
    public:
        /**
         * @brief Iterator
         * 
         */
        class Iterator: public std::iterator<std::bidirectional_iterator_tag, Key> {
        public:
            friend bool operator == ( const Iterator& a, const Iterator& b ) {
                return a.array_ == b.array_ && a.pos_ = b.pos_;
            }
            friend bool operator != ( const Iterator& a, const Iterator& b ) {
                return a.array_ != b.array_ || a.pos_ != b.pos_;
            }
        public:
            /**
             * @brief Construct a new Iterator object
             * 
             * @param array Array
             * @param pos position
             */
            Iterator(Array<Name, Key, Size> *array, size_t pos)
                    :array_(array), pos_(pos){
            }

            /**
             * @brief Get the element value
             * 
             * @return Key& 
             */
            Key& operator*() const {
                return array_->at(pos_-1);
            }

            /**
             * @brief Get the element value
             * 
             * @return Key& 
             */
            Key& operator->() const {
                return array_->at(pos_-1);
            }

            Iterator& operator--(){
                pos_--;
                return *this;
            }

            Iterator operator --(int) {
                BCWasmAssert(pos_ > 0, "pos can't be negative");
                Iterator tmp(array_, pos_--);
                --tmp;
                return tmp;
            }

            Iterator& operator ++() {
                pos_++;
                return *this;
            }

            Iterator operator ++(int) {
                Iterator tmp(array_, pos_++);
                ++tmp;
                return tmp;
            }
        private:

            Array<Name, Key, Size> *array_;
            size_t pos_;
        };

        /**
         * @brief Constant iterator
         * 
         */
        class ConstIterator: public std::iterator<std::bidirectional_iterator_tag, const Key> {
        public:
            friend bool operator == ( const ConstIterator& a, const ConstIterator& b ) {
                return a.array_ == b.array_ && a.pos_ = b.pos_;
            }
            friend bool operator != ( const ConstIterator& a, const ConstIterator& b ) {
                return a.array_ != b.array_ || a.pos_ != b.pos_;
            }
        public:
            /**
             * @brief Construct a new Const Iterator object
             * 
             * @param array Array
             * @param pos position
             */
            ConstIterator(Array<Name, Key, Size> *array, size_t pos)
                    :array_(array), pos_(pos) {
            }

            /**
             * @brief Get the element value
             * 
             * @return Key& 
             */
            const Key& operator*() {
                key_ = array_->getConst(pos_ - 1);
                return key_;
            }

            /**
             * @brief Get the element value
             * 
             * @return Key& 
             */
            const Key& operator->() {
                key_ = array_->getConst(pos_ - 1);
                return key_;
            }

            ConstIterator& operator--(){
                pos_--;
                return *this;
            }

            ConstIterator operator --(int) {
                BCWasmAssert(pos_ > 0, "pos can't be negative");
                ConstIterator tmp(array_, pos_--);
                --tmp;
                return tmp;
            }

            ConstIterator& operator ++() {
                pos_++;
                return *this;
            }

            ConstIterator operator ++(int) {
                ConstIterator tmp(array_, pos_++);
                ++tmp;
                return tmp;
            }
        private:

            Array<Name, Key, Size> *array_;
            size_t pos_;
            Key key_;
        };

        /**
         * @brief Constant reverse iterator
         * 
         */
        class ConstReverseIterator: public std::iterator<std::bidirectional_iterator_tag, const Key> {
        public:
            friend bool operator == ( const ConstReverseIterator& a, const ConstReverseIterator& b ) {
                return a.array_ == b.array_ && a.pos_ = b.pos_;
            }
            friend bool operator != ( const ConstReverseIterator& a, const ConstReverseIterator& b ) {
                return a.array_ != b.array_ || a.pos_ != b.pos_;
            }
        public:
            /**
             * @brief Construct a new Const Reverse Iterator object
             * 
             * @param array Array
             * @param pos position
             */
            ConstReverseIterator(Array<Name, Key, Size> *array, size_t pos)
                    :array_(array), pos_(pos){
            }

            /**
             * @brief Get the element value
             * 
             * @return Key& 
             */
            const Key& operator*() {
                key_ = array_->getConst(pos_ - 1);
                return key_;
            }

            /**
             * @brief Get the element value
             * 
             * @return Key& 
             */
            const Key& operator->() {
                key_ = array_->getConst(pos_ - 1);
                return key_;
            }

            ConstReverseIterator& operator--(){
                pos_++;
                return *this;
            }

            ConstReverseIterator operator --(int) {
                BCWasmAssert(pos_ > 0, "pos can't be negative");
                ConstReverseIterator tmp(array_, pos_++);
                ++tmp;
                return tmp;
            }

            ConstReverseIterator& operator ++() {
                pos_--;
                return *this;
            }

            ConstReverseIterator operator ++(int) {
                ConstReverseIterator tmp(array_, pos_--);
                --tmp;
                return tmp;
            }
        private:

            Array<Name, Key, Size> *array_;
            size_t pos_;
            Key key_;
        };

        typedef std::reverse_iterator<Iterator> ReverseIterator;
    public:
        static_assert(Size != 0, "array no support size = 0");
        Array () {
        }

        Array(const Array<Name, Key, Size> &) = delete;
        Array(const Array<Name, Key, Size> &&) = delete;
        Array<Name, Key, Size>& operator=(const Array<Name, Key, Size> &) = delete;

        ~Array() {
            flush();
        }

        /**
         * @brief Iterator start position
         * 
         * @return Iterator 
         */
        Iterator begin() {
            return Iterator(this, 1);
        }

        /**
         * @brief Iterator end position
         * 
         * @return Iterator 
         */
        Iterator end() {
            return Iterator(this, Size+1);
        }

        /**
         * @brief Reverse iterator start position
         * 
         * @return ReverseIterator 
         */
        ReverseIterator rbegin() {
            return ReverseIterator(end());
        }

        /**
         * @brief Reverse iterator end position
         * 
         * @return ReverseIterator 
         */
        ReverseIterator rend() {
            return ReverseIterator(begin());
        }

        /**
         * @brief Constant iterator start position
         * 
         * @return ConstIterator 
         */
        ConstIterator cbegin() {
            return ConstIterator(this, 1);
        }

        /**
         * @brief Constant iterator end position
         * 
         * @return ConstIterator 
         */
        ConstIterator cend() {
            return ConstIterator(this, Size+1);
        }

        /**
         * @brief Inverse constant iterator start position
         * 
         * @return ConstReverseIterator 
         */
        ConstReverseIterator crbegin() {
            return ConstReverseIterator(this, Size);
        }

        /**
         * @brief Inverse constant iterator end position
         * 
         * @return ConstReverseIterator 
         */
        ConstReverseIterator crend() {
            return ConstReverseIterator(this, 0);
        }

        /**
         * @brief Get the specified position element
         * 
         * @param pos Element position
         * @return Key& Element value
         */
        Key& at(size_t pos) {
            BCWasmAssert(pos < Size, "out of range pos:", pos, "size:", Size);
            auto iter = cache_.find(pos);
            if (iter != cache_.end()) {
                return iter->second;
            }
            Key key;
            getState(encodeKey(pos), key);
            auto iterc = cache_.emplace(std::make_pair(pos, key));
            BCWasmAssert(iterc.second, "cache emplace failed");
            return iterc.first->second;
        }

        /**
         * @brief Bracket operator
         * 
         * @param pos position
         * @return Key& element
         */
        Key& operator[](size_t pos) {
            return at(pos);
        }

        /**
         * @brief 数组长度
         * 
         * @return size_t 
         */
        size_t size() const {
            return Size;
        }
        
        /**
         * @brief Get the Const object. Do not flush to cache
         * 
         * @param pos position
         * @return Key Element value
         */
        Key getConst(size_t pos) const{
            auto iter = cache_.find(pos);
            if (iter != cache_.end()) {
                return iter->second;
            }
            Key key;
            getState(encodeKey(pos), key);
            return key;
        }

        /**
         * @brief Set the Const object, Do not flush to cache
         * 
         * @param pos 
         * @param key 
         */
        void setConst(size_t pos, const Key &key) {
            auto iter = cache_.find(pos);
            if (iter != cache_.end()) {
                cache_[pos] = key;
            }
            setState(encodeKey(pos), key);
        }

    private:
        /**
         * @brief Generate the key of the specified index
         * 
         * @param index 
         * @return std::string 
         */
        std::string encodeKey(size_t index) const {
            std::string key;
            key.reserve(name_.length() + 1 + sizeof(index));
            key.append(name_);
            key.append(1, 'A');
            key.append((char*)&index, sizeof(index));
            return key;
        }

    private:
        /**
         * @brief Refresh data to blockchain
         * 
         */
        void flush() {
            for (auto iter : cache_) {
                std::string key = encodeKey(iter.first);
                setState(key, iter.second);
            }
        }

    public:
        static const std::string kType;
    private:
        std::map<size_t, Key> cache_;

        const std::string name_ = kType + Name;
    };

//    const std::string kType = "array"
    template <const char *Name, typename Key, unsigned Size>
    const std::string Array<Name, Key, Size>::kType = "__array__";
}
}
