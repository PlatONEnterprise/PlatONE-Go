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
     * @brief Implement map operations, Map templates
     * 
     * @tparam *Name The name of the Map, the name of the Map should be unique within each contract.
     * @tparam Key key的类型
     * @tparam Value value的类型
     */
    template <const char *Name, typename Key, typename Value>
    class Map{
    public:

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
                return *second_;           }
        private:
            const Key *first_;
            const Value *second_;
        };

        /**
         * @brief Wrapper
         * @param name first name
         * @param key key
         */
        template <typename WrapKey>
        class Wrapper {
        public:
            /**
             * @brief Construct a new Key Wrapper object
             * 
             * @param name first name
             * @param key key
             */
            Wrapper(const std::string &name, const WrapKey &key) :name_(name), key_(key) {
            }
            /**
             * @brief Construct a new bcwasm serialize object
             * 
             * @param name_ 
             */
            BCWASM_SERIALIZE(Wrapper, (name_)(key_))
        private:
            const std::string &name_;
            const WrapKey& key_;
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
        //template <typename ItemIterator>
        class IteratorType : public std::iterator<std::bidirectional_iterator_tag, Pair> {
        public:
            friend bool operator == ( const IteratorType& a, const IteratorType& b ) {
                return a.map_ == b.map_ && a.index_ == b.index_;
            }
            friend bool operator != ( const IteratorType& a, const IteratorType& b ) {
                bool res =  a.map_ != b.map_ || a.index_ != b.index_;
                return res;
            }
        
        public:
            IteratorType() = default;

            /**
             * @brief Construct a new Iterator Type object
             * 
             * @param map 
             * @param index
             */
            IteratorType(Map<Name, Key, Value> *map, int index)
                :map_(map), index_(index){
            }

            /**
             * @brief The obvious operators
             * 
             * @return Pair& 
             */
            Pair& operator*() {
                Key k;
                Value v;
                size_t len = getState(IndexWrapper(wrapperName1_,index_),k);
                BCWasmAssert(len != 0, "cannot find key!");
                getState(KeyWrapper(wrapperName2_, k), v);
                pair_ = Pair(k,v);
                return pair_;
            }

            Pair* operator->() {
                Key k;
                Value v;
                size_t len = getState(IndexWrapper(wrapperName1_,index_),k);
                BCWasmAssert(len != 0, "cannot find key!");
                getState(KeyWrapper(wrapperName2_,k),v);
                pair_ = Pair(k,v);
                return &pair_;
            }

            IteratorType& operator--(){
                --index_;
                return *this;
            }

            IteratorType operator --(int) {
                IteratorType tmp(map_, index_--);
                //--tmp;
                return tmp;
            }
            
            IteratorType& operator ++() {
                ++index_;
                return *this;
            }
    
            IteratorType operator ++(int) {
                IteratorType tmp(map_, index_++);
                //++tmp;
                return tmp;
            }
        private:
            Pair pair_;
            const std::string wrapperName1_ = kType + Name + "index";
            const std::string wrapperName2_ = kType + Name + "key";
            const std::string wrapperName3_ = kType + Name + "string";
            Map<Name, Key, Value> *map_;
            int  index_;
        };

        /**
         * @brief ReverseIterator
         * 
         */
        class ReverseIteratorType : public std::iterator<std::bidirectional_iterator_tag, Pair> {
        public:
            friend bool operator == ( const ReverseIteratorType& a, const ReverseIteratorType& b ) {
                return a.map_ == b.map_ && a.index_ == b.index_;
            }
            friend bool operator != ( const ReverseIteratorType& a, const ReverseIteratorType& b ) {
                bool res =  a.map_ != b.map_ || a.index_ != b.index_;
                return res;
            }
        
        public:
            ReverseIteratorType() = default;

            /**
             * @brief Construct a new ReverseIterator Type object
             * 
             * @param map 
             * @param index
             */
            ReverseIteratorType(Map<Name, Key, Value> *map, int index)
                :map_(map), index_(index){
            }

            /**
             * @brief The obvious operators
             * 
             * @return Pair& 
             */
            Pair& operator*() {
                Key k;
                Value v;
                size_t len = getState(IndexWrapper(wrapperName1_,index_),k);
                BCWasmAssert(len != 0, "cannot find key!");
                getState(KeyWrapper(wrapperName2_, k), v);
                pair_ = Pair(k,v);
                return pair_;
            }

            Pair* operator->() {
                Key k;
                Value v;
                size_t len = getState(IndexWrapper(wrapperName1_,index_),k);
                BCWasmAssert(len != 0, "cannot find key!");
                getState(KeyWrapper(wrapperName2_,k),v);
                pair_ = Pair(k,v);
                return &pair_;
            }

            IteratorType& operator--(){

                ++index_;
                // index total guanxi
                return *this;
            }

            IteratorType operator --(int) {
                IteratorType tmp(map_, index_++);
                //--tmp;
                return tmp;
            }

            IteratorType& operator ++() {
                --index_;

                return *this;
            }

            IteratorType operator ++(int) {
                IteratorType tmp(map_, index_--);
                //++tmp;
                return tmp;
            }
        private:
            Pair pair_;
            const std::string wrapperName1_ = kType + Name + "index";
            const std::string wrapperName2_ = kType + Name + "key";
            const std::string wrapperName3_ = kType + Name + "string";
            Map<Name, Key, Value> *map_;
            int  index_;
        };
        /**
         * @brief Constant iterator
         * 
         */
        class ConstIteratorType : public std::iterator<std::bidirectional_iterator_tag, ConstPair> {
        public:
            friend bool operator == ( const ConstIteratorType& a, const ConstIteratorType& b ) {
                return a.map_ == b.map_ && a.index_ == b.index_;
            }
            friend bool operator != ( const ConstIteratorType& a, const ConstIteratorType& b ) {
                bool res =  a.map_ != b.map_ || a.index_ != b.index_;
                return res;
            }
        public:

            ConstIteratorType() = default;

            /**
             * @brief Construct a new Const Iterator Type object
             * 
             * @param map 
             * @param index
             */
            ConstIteratorType(const Map<Name, Key, Value> *map, int index)
                    :map_(map), index_(index){
            }

            /**
             * @brief The obvious operators
             * 
             * @return Pair& 
             */
            ConstPair& operator*() {
                Key k;
                Value v;
                size_t len = getState(IndexWrapper(wrapperName1_,index_),k);
                BCWasmAssert(len != 0, "cannot find key!");
                getState(KeyWrapper(wrapperName2_,k),v);
                pair_ = ConstPair(k,v);
                return pair_;
            }

            ConstPair* operator->() {
                Key k;
                Value v;
                size_t len = getState(IndexWrapper(wrapperName1_,index_),k);
                BCWasmAssert(len != 0, "cannot find key!");
                getState(KeyWrapper(wrapperName2_,k),v);
                pair_ = ConstPair(k,v);
                return &pair_;
            }

            ConstIteratorType& operator--(){
                --index_;
                return *this;
            }

            ConstIteratorType operator --(int) {
                ConstIteratorType tmp(map_, index_--);
                //--tmp;
                return tmp;
            }

            ConstIteratorType& operator ++() {
                ++index_;
                return *this;
            }
            ConstIteratorType operator ++(int) {
                ConstIteratorType tmp(map_, index_++);
                //++tmp;
                return tmp;
            }

        private:
            ConstPair pair_;
            const Map<Name, Key, Value> *map_;
            const std::string wrapperName1_ = kType + Name + "index";
            const std::string wrapperName2_ = kType + Name + "key";
            const std::string wrapperName3_ = kType + Name + "string";
            int index_;
        };

        /**
         * @brief Constant iterator
         * 
         * @tparam ItemIterator 
         */
        // template <typename ItemIterator>
        class ConstReverseIteratorType : public std::iterator<std::bidirectional_iterator_tag, ConstPair> {
        public:
            friend bool operator == ( const ConstReverseIteratorType& a, const ConstReverseIteratorType& b ) {
                return a.map_ == b.map_ && a.index_ == b.index_;
            }
            friend bool operator != ( const ConstReverseIteratorType& a, const ConstReverseIteratorType& b ) {
                bool res =  a.map_ != b.map_ || a.index_ != b.index_;
                return res;
            }
        public:

            ConstReverseIteratorType() = default;

            /**
             * @brief Construct a new Const ReverseIterator Type object
             * 
             * @param map 
             * @param index
             */
            ConstReverseIteratorType(const Map<Name, Key, Value> *map, int index)
                    :map_(map), index_(index){
            }

            /**
             * @brief The obvious operators
             * 
             * @return Pair& 
             */
            ConstPair& operator*() {
                Key k;
                Value v;
                size_t len = getState(IndexWrapper(wrapperName1_,index_),k);
                BCWasmAssert(len != 0, "cannot find key!");
                getState(KeyWrapper(wrapperName2_,k),v);
                pair_ = ConstPair(k,v);
                return pair_;
            }

            ConstPair* operator->() {
                Key k;
                Value v;
                size_t len = getState(IndexWrapper(wrapperName1_,index_),k);
                BCWasmAssert(len != 0, "cannot find key!");
                getState(KeyWrapper(wrapperName2_,k),v);
                pair_ = ConstPair(k,v);
                return &pair_;
            }

            ConstReverseIteratorType& operator--(){
                ++index_;
                return *this;
            }

            ConstReverseIteratorType operator --(int) {
                ConstReverseIteratorType tmp(map_, index_++);
                return tmp;
            }

            ConstReverseIteratorType& operator ++() {
                --index_;
                return *this;
            }
            ConstReverseIteratorType operator ++(int) {
                ConstReverseIteratorType tmp(map_, index_--);
                return tmp;
            }

        private:
            ConstPair pair_;
            const Map<Name, Key, Value> *map_;
            const std::string wrapperName1_ = kType + Name + "index";
            const std::string wrapperName2_ = kType + Name + "key";
            const std::string wrapperName3_ = kType + Name + "string";
            int index_;
        };
        /**
        * @define three Wrapper type here
        *
        */
        typedef class Wrapper<Key> KeyWrapper;
        typedef class Wrapper<int> IndexWrapper;
        typedef class Wrapper<std::string> StrWrapper;
        typedef class IteratorType Iterator;
        typedef class ReverseIteratorType ReverseIterator;
        typedef class ConstIteratorType ConstIterator;
        typedef class ConstReverseIteratorType ConstReverseIterator;

    public:

        Map(){init();}
        Map(const Map<Name, Key, Value> &) = delete;
        Map(const Map<Name, Key, Value> &&) = delete;
        Map<Name, Key, Value>& operator=(const Map<Name, Key, Value> &) = delete;
        /**
         * @brief Destroy the Map object Refresh data to the blockchain
         * 
         */
        ~Map(){
            free(v_ptr);
        }


        /**
         * @brief Insert a new key-value pair, Update to leveldb
         * 
         * @param k Key
         * @param v Value
         * @return true Inserted successfully
         * @return false Insert failed
         */
        bool insert(const Key &k, const Value &v) {
            init();
    
            Value v_;
            size_t len = bcwasm::getState(KeyWrapper(wrapperName2_, k), v_);
            if(0 != len){
                    return false;
            }
            setState(IndexWrapper(wrapperName1_, total),k);          
            total += 1;
            setState(KeyWrapper(wrapperName2_, k),v);
            setState(StrWrapper(wrapperName3_, TOTAL),total);
            return true;
        }
        /**
         * @brief Update a new key-value pair, Update to leveldb
         * 
         * @param k Key
         * @param v Value
         * @return true Updated successfully
         * @return false Update failed
         */
         bool update(const Key &k, const Value &v) {
            init();
            Value v_;
            size_t len = bcwasm::getState(KeyWrapper(wrapperName2_, k), v_);
            if(0 == len){
                return false;
            }
            setState(KeyWrapper(wrapperName2_, k), v);
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
            size_t len = bcwasm::getState(KeyWrapper(wrapperName2_, k), *v_ptr);
            if (0 != len) {
                return (const Value*)v_ptr;
            }
            return nullptr;
        }
        /**
         * @brief Get value Ptr
         * 
         * @param k Key
         * @return Value* 
         */
        Value* find(const Key &k) {
            init();
            BCWasmAssert(v_ptr != nullptr, "v_ptr is null.");
            size_t len = bcwasm::getState(KeyWrapper(wrapperName2_, k), *v_ptr);
            if (0 != len) {
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
            Key k1;
            Value v1;
            Key lastKey;
            Value lastValue;
            if(find(k)==nullptr){
                return;
            }
            bcwasm::delState(KeyWrapper(wrapperName2_,k));
            for(int i=0; i<total;i++)
            {
                getState(IndexWrapper(wrapperName1_,i),k1);  
                size_t len_1 = getState(KeyWrapper(wrapperName2_,k1),v1);
                if(0 == len_1)
                {
                    if(i == total-1)
                    {
                        bcwasm::delState(IndexWrapper(wrapperName1_,i));
                        setState(StrWrapper(wrapperName3_, TOTAL),--total);       
                        //this.iter.total = total;
                        break;
                    }
                    getState(IndexWrapper(wrapperName1_, total-1),lastKey);
                    setState(IndexWrapper(wrapperName1_,i),lastKey);
                    bcwasm::delState(IndexWrapper(wrapperName1_, total-1));
                    setState(StrWrapper(wrapperName3_, TOTAL),--total);        
                    break;
                }

            }                                              
        }
         /**
         * @brief Get the length of the map
         * 
         * @return int length
         */
        int size() {
            init();
            return total;
        }

        int size() const {
            return total;
        }

        /**
         * @brief Iterator start position
         * 
         * @return Iterator 
         */
        //Iterator 
        Iterator begin() {
            int begin = 0;
            return Iterator(this, begin);
            // this.iter.index = 0;
            // return this.iter;
        }
        /**
         * @brief const Iterator start position
         * 
         * @return cbegin()
         */ 
        ConstIterator begin() const {
            return cbegin();
        }
          /**
         * @brief ReverseIterator start position
         * 
         * @return ReverseIterator 
         */
        //Iterator 
        ReverseIterator rbegin() {
            int rbegin = total-1;
            return ReverseIterator(this, rbegin);
        }
          /**
         * @brief const ReverseIterator start position
         * 
         * @return crbegin()
         */
        //Iterator 
        ConstReverseIterator rbegin() const {
            return crbegin();
        }

        /**
         * @brief Iterator end position
         * 
         * @return Iterator 
         */
        /*Iterator*/
        Iterator  end() {
           // return Iterator(this,total);
        //    this.iter.index = total;
        //    this.iter.total = total;
        //    if (total == 0) {
        //        return nullptr;
        //    }
        //    return this.iter;
        return Iterator(this,total);
        }
        /**
         * @brief const Iterator end position
         * 
         * @return cend() 
         */
        /*Iterator*/
        ConstIterator end() const {
            return cend();
        }
        /**
         * @brief Reverse Iterator end position
         * 
         * @return ReverseIterator 
         */
        /*Iterator*/
        ReverseIterator  rend() {
            return ReverseIterator(this,rend_);
        }
        /**
         * @brief const ReverseIterator end position
         * 
         * @return crend
         */
        /*Iterator*/
        ConstReverseIterator rend() const {
            return crend();
        }
        /**
         * @brief const Iterator end position
         * 
         * @return const Iterator 
         */
        /*Iterator*/
        ConstIterator cbegin() const {
            int begin = 0;
            return ConstIterator(this, begin);
        }
        /**
         * @brief const ReverseIterator end position
         * 
         * @return ConstReverseIterator 
         */
        /*Iterator*/
        ConstReverseIterator crbegin() const {
            int rbegin = total-1;
            return ConstReverseIterator(this, rbegin);
        }
        /**
         * @brief const Iterator end position
         * 
         * @return const Iterator 
         */
        /*Iterator*/
        ConstIterator cend() const {
            return ConstIterator(this, total);
        }
        /**
         * @brief const ReverseIterator end position
         * 
         * @return ConstReverseIterator 
         */
        /*Iterator*/
        ConstReverseIterator crend() const {
            return ConstReverseIterator(this, rend_);
        }


    public:
        static const std::string kType;
    private:
        /**
         * @brief Initialize, get data from the blockchain
         * 
         */
        void init() {
            
            if (!init_) {
            //iter =  Iterator(this, begin); // index:0, total
            v_ptr = (Value *)malloc(sizeof(Value));
            BCWasmAssert(v_ptr != nullptr, "unable to allocate memory for v_ptr.");
            bcwasm::getState(StrWrapper(wrapperName3_, TOTAL), total);
            init_ = true;
            }
        }

        int total = 0;
        //IteratorType iter;

        const std::string wrapperName1_ = kType + Name + "index";
        const std::string wrapperName2_ = kType + Name + "key";
        const std::string wrapperName3_ = kType + Name + "string";
        const std::string TOTAL = "total";
        bool init_ = false;
        Value *v_ptr;
        const int rend_= -1; 
    };
    
    template <const char *Name, typename Key, typename Value>
    const std::string Map<Name, Key, Value>::kType = "__map__";

}
}