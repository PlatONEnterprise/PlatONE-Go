//
// Created by zhou.yang on 2018/11/10.
//

#pragma once

#include <map>
#include <vector>
#include <set>
#include "bcwasm/storage.hpp"
#include "bcwasm/serialize.hpp"
#include "bcwasm/print.hpp"

/**
 * @brief Implement map operation
 * 
 */
namespace bcwasm {
    namespace db {
        /**
         * @brief Traverse is a variable that can be made, NoTraverse is an immutable
         * 
         */
    enum class MapType {
        Traverse = 0,
        NoTraverse = 1
    };

    /**
     * @brief Implement map operations, Map templates
     * 
     * @tparam *Name The name of the Map, the name of the Map should be unique within each contract.
     * @tparam Key key的类型
     * @tparam Value value的类型
     * @tparam MapType::Traverse The default is Traverse, when Traverse needs extra data structure to operate, set to NoTraverse when no traversal operation is needed. 
     */
    template <const char *Name, typename Key, typename Value, MapType type = MapType::Traverse>
    class Map{
    public:
        //template <const char *Name, typename Key, typename Value>

        class Pair {
        public:
            /**
             * @brief Construct a new Pair object
             * 
             */
            Pair():first_(nullptr), second_(nullptr){}
            /**
             * @brief Construct a new Pair object
             * 
             * @param key key
             * @param val value
             */
            Pair(const Key &key, const Value &val)
                    :first_(&key), second_(&val){
            }
//            Pair(const Key &&key, const Value &&val)
//                    :first_(&key), second_(&val){
//            }
            /**
             * @brief Construct a new Pair object
             * 
             * @param pair 
             */
            Pair(const Pair& pair):first_(pair.first_), second_(pair.second_) {
            }
            /**
             * @brief Assignment operator
             * 
             * @param pair 
             * @return Pair& 
             */
            Pair & operator = (const Pair& pair) {
                first_ = pair.first_;
                second_ = pair.second_;
                return *this;
            }
            /**
             * @brief Get key
             * 
             * @return const Key& 
             */
            const Key& first() const {
                BCWasmAssert(first_ != nullptr, "first is nullptr");
                return *first_;
            }

            /**
             * @brief Get value
             * 
             * @return const Value& 
             */
            const Value& second() const {
                BCWasmAssert(second_ != nullptr, "second is nullptr");
                return *second_;
            }
        private:
            const Key *first_;
            const Value *second_;
        };

        /**
         * @brief KeyWrapper
         * 
         */
        class KeyWrapper {
        public:
            /**
             * @brief Construct a new Key Wrapper object
             * 
             * @param name first name
             * @param key key
             */
            KeyWrapper(const std::string &name, const Key &key) :name_(name), key_(key) {
            }
            /**
             * @brief Construct a new bcwasm serialize object
             * 
             * @param name_ 
             */
            BCWASM_SERIALIZE(KeyWrapper, (name_)(key_))
        private:
            const std::string &name_;
            const Key& key_;
        };

        /**
         * @brief Constant Pair
         * 
         */
        class ConstPair {
        public:
            ConstPair() = default;
           /**
            * @brief Construct a new Const Pair object
            * 
            * @param key key
            * @param val value
            */
            ConstPair(const Key &key, const Value &val)
                    :first_(key), second_(val){
            }
//            Pair(const Key &&key, const Value &&val)
//                    :first_(&key), second_(&val){
//            }

            /**
             * @brief Construct a new Const Pair object
             * 
             * @param pair 
             */
            ConstPair(const ConstPair& pair):first_(pair.first_), second_(pair.second_) {
            }
            /**
             * @brief Assignment operator
             * 
             * @param pair 
             * @return ConstPair& 
             */
            ConstPair & operator = (const ConstPair& pair) {
                first_ = pair.first_;
                second_ = pair.second_;
                return *this;
            }
            /**
             * @brief Get key
             * 
             * @return const Key& 
             */
            const Key& first() const {
                return first_;
            }
            /**
             * @brief Get value
             * 
             * @return const Value& 
             */
            const Value& second() const {
                return second_;
            }
        private:
            Key first_;
            Value second_;
        };
        /**
         * @brief Iterator
         * 
         * @tparam ItemIterator 
         */
        template <typename ItemIterator>
        class IteratorType : public std::iterator<std::bidirectional_iterator_tag, Pair> {
        public:
            friend bool operator == ( const IteratorType& a, const IteratorType& b ) {
                return a.map_ == b.map_ && a.iter_ == b.iter_;
            }
            friend bool operator != ( const IteratorType& a, const IteratorType& b ) {
                bool res =  a.map_ != b.map_ || a.iter_ != b.iter_;
                return res;
            }
        public:


            IteratorType() = default;

            /**
             * @brief Construct a new Iterator Type object
             * 
             * @param map 
             * @param iter 
             */
            IteratorType(Map<Name, Key, Value, type> *map, ItemIterator iter)
                :map_(map), iter_(iter){
            }

            /**
             * @brief The obvious operators
             * 
             * @return Pair& 
             */
            Pair& operator*() {
                pair_ = Pair(*iter_, (*map_)[*iter_]);
                return pair_;
            }

            Pair* operator->() {
                pair_ = Pair(*iter_, (*map_)[*iter_]);
                return &pair_;
            }

            IteratorType& operator--(){
                --iter_;
                return *this;
            }

            IteratorType operator --(int) {
                IteratorType tmp(map_, iter_--);
                --tmp;
                return tmp;
            }

            IteratorType& operator ++() {
                ++iter_;
                return *this;
            }

            IteratorType operator ++(int) {
                IteratorType tmp(map_, iter_++);
                ++tmp;
                return tmp;
            }
        private:
            Pair pair_;
            Map<Name, Key, Value, type> *map_;
            ItemIterator iter_;
        };

        /**
         * @brief Constant iterator
         * 
         * @tparam ItemIterator 
         */
        template <typename ItemIterator>
        class ConstIteratorType : public std::iterator<std::bidirectional_iterator_tag, ConstPair> {
        public:
            friend bool operator == ( const ConstIteratorType& a, const ConstIteratorType& b ) {
                return a.map_ == b.map_ && a.iter_ == b.iter_;
            }
            friend bool operator != ( const ConstIteratorType& a, const ConstIteratorType& b ) {
                bool res =  a.map_ != b.map_ || a.iter_ != b.iter_;
                return res;
            }
        public:

            ConstIteratorType() = default;

            /**
             * @brief Construct a new Const Iterator Type object
             * 
             * @param map 
             * @param iter 
             */
            ConstIteratorType(const Map<Name, Key, Value, type> *map, ItemIterator iter)
                    :map_(map), iter_(iter){
            }

            /**
             * @brief The obvious operators
             * 
             * @return Pair& 
             */
            ConstPair& operator*() {
                pair_ = ConstPair(*iter_, map_->find(*iter_));
                return pair_;
            }

            ConstPair* operator->() {
                pair_ = ConstPair(*iter_, *(map_->find(*iter_)));
                return &pair_;
            }

            ConstIteratorType& operator--(){
                --iter_;
                return *this;
            }

            ConstIteratorType operator --(int) {
                ConstIteratorType tmp(map_, iter_--);
                --tmp;
                return tmp;
            }

            ConstIteratorType& operator ++() {
                ++iter_;
                return *this;
            }

            ConstIteratorType operator ++(int) {
                ConstIteratorType tmp(map_, iter_++);
                ++tmp;
                return tmp;
            }
        private:
            ConstPair pair_;
            const Map<Name, Key, Value, type> *map_;
            ItemIterator iter_;
        };

        typedef class IteratorType<typename std::set<Key>::iterator> Iterator;
        typedef class IteratorType<typename std::set<Key>::reverse_iterator> ReverseIterator;
        typedef class ConstIteratorType<typename std::set<Key>::iterator> ConstIterator;
        typedef class ConstIteratorType<typename std::set<Key>::reverse_iterator> ConstReverseIterator;
    public:

        Map(){init();}
        Map(const Map<Name, Key, Value, type> &) = delete;
        Map(const Map<Name, Key, Value, type> &&) = delete;
        Map<Name, Key, Value, type>& operator=(const Map<Name, Key, Value, type> &) = delete;
        /**
         * @brief Destroy the Map object Refresh data to the blockchain
         * 
         */
        ~Map(){
            flush();
            free(v_ptr);
        }


        /**
         * @brief Insert a new key-value pair, Update to cache
         * 
         * @param k Key
         * @param v Value
         * @return true Inserted successfully
         * @return false Insert failed
         */
        bool insert(const Key &k, const Value &v) {
            init();
            map_[k] = v;             
            modify_.insert(k);

            if (type == MapType::Traverse) {
                keySet_.insert(k);             
            }
            // setState(curIndex, k);
            // setState(k, v);
            // total += 1;
            // setState(TOTAL, total);
            return true;
        }

        /**
         * @brief Insert a new key-value pair that will not be updated to the cache. Suitable for large number of inserts, no updates after insertion
         * 
         * @param k Key
         * @param v Value
         * @return true Inserted successfully
         * @return false Insert failed
         */
        bool insertConst(const Key &k, const Value &v) {
            init();
            if (type == MapType::Traverse) {
                keySet_.insert(k);
            }

            if (map_.find(k) != map_.end()) {
                map_[k] = v;
            }

            setState(KeyWrapper(keySetName_, k), v);
            return true;
        }

        /**
         * @brief Get const value Ptr
         * 
         * @param k Key
         * @return Value* 
         */
        const Value* find(const Key &k) const {
            BCWasmAssert(v_ptr != nullptr, "v_ptr is null.");
            size_t len = bcwasm::getState(k, *v_ptr);
            if (0 != len) {
                return (const Value*)v_ptr;
            }
            return nullptr;
        }

        Value* find(const Key &k) {
            init();
            auto iter = map_.find(k);
            if (iter != map_.end()) {
                return &iter->second;
            }
            size_t len = bcwasm::getState(k, *v_ptr);
            if (0 != len) {
                map_[k] = *v_ptr;
                keySet_.insert(k);
                return v_ptr;
            }
            return nullptr;
        }

        /**
         * @brief Delete key-value pairs
         * 
         * @param k Key
         */
        void del(const Key &k) {
            init();
            auto iter = map_.find(k);
            if (iter != map_.end()) {
                map_.erase(iter);
            }
            if (type == MapType::Traverse) {
                //0424: 删除元素的处理，将total-1位置的key放入删除的位置
                auto it = keySet_.find(k);
                if (it == keySet_.end()) {
                    //elem to delete is not found.
                    return;
                }
                const int keyIndex = std::distance(keySet_.begin(),it);
                Key lastKey;
                getState(total-1, lastKey);
                setState(keyIndex, lastKey);

                bcwasm::delState(total-1);
                bcwasm::delState(k);

                bcwasm::setState(TOTAL, total-1);

                keySet_.erase(k);
            }
            modify_.insert(k);
        }

        /**
         * @brief Bracket operator
         * 
         * @param k Key
         * @return Value& Get Value 
         */
        Value& operator[](const Key& k) {
            init();
            return *find(k);
        }

        const Value& operator[](const Key& k) const{
            return *find(k);
        }

        /**
         * @brief Get the length of the map, only allowed when the MapType is Traverse
         * 
         * @return size_t length
         */
        size_t size() {
            init();
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return keySet_.size();
        }

        size_t getSize() const 
        {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return keySet_.size();
        }

        const std::set<Key>& getKeys() const
        {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return keySet_;
        }

        /**
         * @brief Refresh the modified data in memory to the blockchain
         * 
         */
        void flush() {
            std::for_each(
                    modify_.begin(),
                    modify_.end(),
                    [this](const Key &k) {
                        auto iter = map_.find(k);
                        if (iter != map_.end()) {
                            Value v;
                            if (0 == getState(k, v)) {
                                int curIndex = total;
                                setState(curIndex, k);
                                setState(k, iter->second);
                                setState(TOTAL, ++total);
                            }
                            else {
                                bcwasm::setState(k, iter->second);
                            }
                        } else {
                            bcwasm::delState(k);
                            if (type == MapType::Traverse) {
                                keySet_.erase(k);
                            }
                        }
                    }
            );
        }

        /**
         * @brief Iterator start position
         * 
         * @return Iterator 
         */
        Iterator begin() {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return Iterator(this, keySet_.begin());
        }

        ConstIterator begin() const {
            return cbegin();
        }

        /**
         * @brief Iterator end position
         * 
         * @return Iterator 
         */
        Iterator end() {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return Iterator(this, keySet_.end());
        }

        ConstIterator end() const {
            return cend();
        }

        /**
         * @brief Reverse iterator start position
         * 
         * @return ReverseIterator 
         */
        ReverseIterator rbegin() {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return ReverseIterator(this, keySet_.rbegin());
        }

        /**
         * @brief Reverse iterator end position
         * 
         * @return ReverseIterator 
         */
        ReverseIterator rend() {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return ReverseIterator(this, keySet_.rend());
        }

        /**
         * @brief Constant iterator start position
         * 
         * @return ConstIterator 
         */
        ConstIterator cbegin() const {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return ConstIterator(this, keySet_.begin());
        }

        /**
         * @brief Constant iterator end position
         * 
         * @return ConstIterator 
         */
        ConstIterator cend() const {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return ConstIterator(this, keySet_.end());
        }

        /**
         * @brief Inverse constant iterator start position
         * 
         * @return ConstReverseIterator 
         */
        ConstReverseIterator crbegin() {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return ConstReverseIterator(this, keySet_.crbegin());
        }

        /**
         * @brief Inverse constant iterator end position
         * 
         * @return ConstReverseIterator 
         */
        ConstReverseIterator crend() {
            BCWasmAssert(type == MapType::Traverse, "NoTraverse of Map", keySetName_);
            return ConstReverseIterator(this, keySet_.crend());
        }

    public:
        static const std::string kType;
    private:
        /**
         * @brief Initialize, get data from the blockchain
         * 
         */
        void init() {
            if (!init_ && type == MapType::Traverse) {

            v_ptr = (Value *)malloc(sizeof(Value));
            BCWasmAssert(v_ptr != nullptr, "unable to allocate memory for v_ptr.");

            bcwasm::getState(TOTAL, total);

            if (!keySet_.empty()) {
                keySet_.clear();
            }
            for (int i=0; i<total;i++) {
                Key k;
                Value v;
                getState(i, k);
                keySet_.insert(k);
            }
            init_ = true;
            }
        }

        int total = 0;
        std::map<Key, Value> map_;
        std::set<Key> keySet_;
        std::set<Key> modify_;
        const std::string keySetName_ = kType + Name;
        const std::string TOTAL = "total";
        bool init_ = false;
        Value *v_ptr;
    };

    template <const char *Name, typename Key, typename Value, MapType type>
    const std::string Map<Name, Key, Value, type>::kType = "__map__";
}
}