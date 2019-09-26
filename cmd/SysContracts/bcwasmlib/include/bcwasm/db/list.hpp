//
// Created by zhou.yang on 2018/11/12.
//

#pragma once
#include "bcwasm/assert.h"
#include "bcwasm/storage.hpp"


namespace bcwasm {
    namespace db {
    /**
     * @brief Implementation list function
     * 
     * @tparam *Name List name, list name is guaranteed to be unique in the same contract
     * @tparam Key List element type
     */
    template <const char *Name, typename Key>
    class List {
    private:
        /**
         * @brief Element state
         * 
         */
        enum State {
            DEL = 0,
            MOD = 1,
            NORMAL = 2
        };

        class Item {
        public:
            Item(){}
            /**
             * @brief Construct a new Item object
             * 
             * @param key Element value
             * @param state status
             */
            Item(const Key &key, State state)
                :key_(key), state_(state) {

            }

            /**
             * @brief Set the State object
             * 
             * @param state Set element status
             */
            void setState(State state) {
                state_ = state;
            }

            /**
             * @brief Get the State object
             * 
             * @return State Get the element status
             */
            State getState() const {
                return state_;
            }

            /**
             * @brief Set the Key object
             * 
             * @param key Set element value
             */
            void setKey(const Key &key) {
                key_ = key;
            }

            /**
             * @brief Get the Key object
             * 
             * @return Key& The obtained element value
             */
            Key & getKey() {
                return key_;
            }

            /**
             * @brief Get the Key object
             * 
             * @return Key The obtained element value
             */
            Key getConstKey() const{
                return key_;
            }

            

        private:
            Key key_;
            State state_;
        };

    public:
        /**
         * @brief Iterator
         * 
         */
        class Iterator : public std::iterator<std::bidirectional_iterator_tag, Key>{
        public:
            friend bool operator == ( const Iterator& a, const Iterator& b ) {
                return a.list_ == b.list_ && a.pos_ = b.pos_;
            }
            friend bool operator != ( const Iterator& a, const Iterator& b ) {
                return a.list_ != b.list_ || a.pos_ != b.pos_;
            }
        public:
            /**
             * @brief Construct a new Iterator object
             * 
             * @param list List
             * @param pos starting point
             */
            Iterator(List<Name, Key> *list, size_t pos)
                    :list_(list), pos_(pos){
            }

            /**
             * @brief Get element
             * 
             * @return Key& 
             */
            Key& operator*() const {
                return list_->get(pos_-1);
            }

            /**
             * @brief Get element
             * 
             * @return Key& 
             */
            Key& operator->() const {
                return list_->get(pos_-1);
            }

            Iterator& operator--(){
                pos_--;
                return *this;
            }

            Iterator operator --(int) {
                BCWasmAssert(pos_ > 0, "pos can't be negative");
                Iterator tmp(list_, pos_--);
                --tmp;
                return tmp;
            }

            Iterator& operator ++() {
                pos_++;
                return *this;
            }

            Iterator operator ++(int) {
                Iterator tmp(list_, pos_++);
                ++tmp;
                return tmp;
            }
        private:

            List<Name, Key> *list_;
            size_t pos_;
        };

        /**
         * @brief Constant iterator
         * 
         */
        class ConstIterator : public std::iterator<std::bidirectional_iterator_tag, Key>{
        public:
            friend bool operator == ( const ConstIterator &a, const ConstIterator &b ) {
                return a.list_ == b.list_ && a.pos_ = b.pos_;
            }
            friend bool operator != ( const ConstIterator &a, const ConstIterator &b ) {
                return a.list_ != b.list_ || a.pos_ != b.pos_;
            }
        public:
            /**
             * @brief Construct a new Const Iterator object
             * 
             * @param list List
             * @param pos starting point
             */
            ConstIterator(List<Name, Key> *list, size_t pos)
                    :list_(list), pos_(pos){
            }

            /**
             * @brief Get element
             * 
             * @return Key& 
             */
            Key& operator*() {
                key_ = list_->getConst(pos_ -1);
                return key_;
            }

            /**
             * @brief Get element
             * 
             * @return Key& 
             */
            Key& operator->() {
                key_ = list_->getConst(pos_ -1);
                return key_;
            }

            ConstIterator& operator--(){
                pos_--;
                return *this;
            }

            ConstIterator operator --(int) {
                BCWasmAssert(pos_ > 0, "pos can't be negative");
                ConstIterator tmp(list_, pos_--);
                --tmp;
                return tmp;
            }

            ConstIterator& operator ++() {
                pos_++;
                return *this;
            }

            ConstIterator operator ++(int) {
                ConstIterator tmp(list_, pos_++);
                ++tmp;
                return tmp;
            }
        private:

            List<Name, Key> *list_;
            size_t pos_;
            Key key_;
        };
        typedef std::reverse_iterator<Iterator> ReverseIterator;
        typedef std::reverse_iterator<ConstIterator> ConstReverseIterator;
    public:
        /**
         * @brief Construct a new List object
         * 
         */
        List(){
            init();
        }

        List(const List<Name, Key> &) = delete;
        List(const List<Name, Key> &&) = delete;
        List<Name, Key>& operator=(const List<Name, Key> &) = delete;

        /**
         * @brief Destroy the List object. Refresh into the blockchain
         * 
         */
        ~List(){
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
            return Iterator(this, size_+1);
        }

        /**
         * @brief Reverse iterator start position
         * 
         * @return ReverseIterator 
         */
        ReverseIterator rbegin() {
            return std::make_reverse_iterator(end());
        }

        /**
         * @brief Reverse iterator end position
         * 
         * @return ReverseIterator 
         */
        ReverseIterator rend() {
            return make_reverse_iterator(begin());
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
            return ConstIterator(this, size_+1);
        }

        /**
         * @brief Inverse constant iterator start position
         * 
         * @return ConstReverseIterator 
         */
        ConstReverseIterator crbegin() {
            return std::make_reverse_iterator(cend());
        }

        /**
         * @brief Inverse constant iterator end position
         * 
         * @return ConstReverseIterator 
         */
        ConstReverseIterator crend() {
            return make_reverse_iterator(cbegin());
        }

        /**
         * @brief Add element
         * 
         * @param k element
         */
        void push(const Key &k){
            cache_[maxNumber_++] = Item(k, MOD);
            mark_.push_back(true);
            ++size_;
        }

        /**
         * @brief Get the specified position element
         * 
         * @param index position
         * @return Key& element
         */
        Key& get(size_t index) {
            BCWasmAssert(index < size_, "out of range", "index:", index, "size:", size_);


            size_t count = 0;
            size_t i = 0;
            for (; i < mark_.size(); ++i) {
                if (count == index && mark_[i]) {
                    break;
                }

                if (mark_[i]) {
                    count++;
                }
            }

            BCWasmAssert(i != mark_.size());

            auto iter = cache_.find(i);
            if (iter != cache_.end()) {
                return iter->second.getKey();
            }

            Key res;
            std::string key = encodeKey(i);
            if (getState(key, res) == 0) {
                bcwasmThrow("getState error list name:", name_, "index:", index, "mark pos;", i);
            }
            auto iterc = cache_.emplace(std::make_pair(count, Item(res, MOD)));


            return iterc.first->second.getKey();
        }

        /**
         * @brief Delete specified location element
         * 
         * @param index element
         */
        void del(size_t index) {
            BCWasmAssert(index < size_, "out of range index:", index, "size:", size_);
            auto iter = cache_.find(index);
            if (iter != cache_.end()) {
                --size_;
                return iter->second.setState(DEL);
            }
            size_t count = 0;
            size_t i = 0;
            for (; i < mark_.size(); ++i) {
                if (count == index && mark_[i]) {
                    break;
                }
                if (mark_[i]) {
                    count++;
                }
            }
            if (i != mark_.size()) {
                Key res;
                std::string key = encodeKey(count);
                bcwasm::delState(key);
                mark_[i] = false;
                --size_;
            }
        }

        /**
         * @brief Delete the element of the specified value
         * 
         * @param delKey Specified element value
         */
        void del(const Key &delKey) {
            for (size_t i = 0; i < mark_.size(); ++i) {
                if (mark_[i]) {
                    Key res;
                    std::string key = encodeKey(i);
                    getState(key, res);
                    if (res == delKey) {
                        cache_.erase(i);
                        bcwasm::delState(key);
                        --size_;
                    }
                }
            }
        }

        /**
         * @brief Bracket operator
         * 
         * @param i position
         * @return Key& element
         */
        Key& operator[](size_t i) {
            return get(i);
        };


        /**
         * @brief Get the element value, not added to the cache
         * 
         * @param index position
         * @return Key element
         */
        Key getConst(size_t index) const{
            BCWasmAssert(index < size_, "out of range", "index:", index, "size:", size_);


            size_t count = 0;
            size_t i = 0;
            for (; i < mark_.size(); ++i) {
                if (count == index && mark_[i]) {
                    break;
                }

                if (mark_[i]) {
                    count++;
                }
            }

            BCWasmAssert(i != mark_.size());

            auto iter = cache_.find(i);
            if (iter != cache_.end()) {
                return iter->second.getConstKey();
            }

            Key res;
            std::string key = encodeKey(i);
            if (getState(key, res) == 0) {
                bcwasmThrow("getState error list name:", name_, "index:", index, "mark pos;", i);
            }
            return res;
        }

        /**
         * @brief Specify location value for element, no element to cache
         * 
         * @param index position
         * @param key element
         */
        void setConst(size_t index , const Key &key)  {
            BCWasmAssert(index < size_, "out of range", "index:", index, "size:", size_);


            size_t count = 0;
            size_t i = 0;
            for (; i < mark_.size(); ++i) {
                if (count == index && mark_[i]) {
                    break;
                }

                if (mark_[i]) {
                    count++;
                }
            }

            BCWasmAssert(i != mark_.size());

            std::string skey = encodeKey(i);
            setState(skey, key);
            if (cache_.find(i) != cache_.end()) {
                cache_.erase(i);
            }
        }

        /**
         * @brief List length
         * 
         * @return size_t 
         */
        size_t size() const{
            return size_;
        }



    private:
        /**
         * @brief 初始化
         * 
         */
        void init() {
            getMaxNumber();
            getSize();
            getMark();
        }

        /**
         * @brief Refresh data to blockchain
         * 
         */
        void flush() {
            for (auto it : cache_) {
                if (it.second.getState() == DEL) {
                    bcwasm::delState(encodeKey(it.first));
                    mark_[it.first] = false;
                } else if (it.second.getState() == MOD) {
                    std::string key = encodeKey(it.first);
                    setState(key, it.second.getKey());
                }
            }
            setMark();
            setMaxNumber();
            setSize();
        }

        /**
         * @brief Set the Mark object
         * 
         */
        void setMark() {
            setState(name_, mark_);
        }

        /**
         * @brief Get the Mark object
         * 
         */
        void getMark() {
            getState(name_, mark_);
        }

        /**
         * @brief Set the Max Number object
         * 
         */
        void setMaxNumber(){
            setState(maxNumberKey_, maxNumber_);
        }

        /**
         * @brief Get the Max Number object
         * 
         */
        void getMaxNumber() {
            getState(maxNumberKey_, maxNumber_);
        }

        /**
         * @brief Set the Size object
         * 
         */
        void setSize(){
            setState(sizeKey_, size_);
        }

        /**
         * @brief Get the Size object
         * 
         */
        void getSize(){
            getState(sizeKey_, size_);
        }

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
            key.append(1, 'L');
            key.append((char*)&index, sizeof(index));
            return key;
        }
    public:
        static const std::string kType;
    private:
        std::map<int, Item> cache_;
        std::vector<bool> mark_;
        size_t maxNumber_ = 0;
        size_t size_ = 0;
        const std::string name_ = kType + Name;
        const std::string maxNumberKey_ = name_ + "maxNumber";
        const std::string sizeKey_ = name_ + "size";
    };
    template <const char *Name, typename Key>
    const std::string List<Name, Key>::kType = "__list__";
}
}
