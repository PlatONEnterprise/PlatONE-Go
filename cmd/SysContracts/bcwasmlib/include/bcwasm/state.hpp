//
// Created by zhou.yang on 2018/10/19.
//

#pragma once

#include "fixedhash.hpp"


extern "C" {
    int64_t gasPrice();
    void blockHash(int64_t num,  uint8_t[32]);
    uint64_t number();
    uint64_t gasLimit();
    int64_t timestamp();
    void coinbase(uint8_t hash[20]);
    void balance(uint8_t amount[32]);
    void origin(uint8_t hash[20]);
    void caller(uint8_t hash[20]);
    int64_t isOwner(const uint8_t *contract, size_t contractLen, const uint8_t *account, size_t accountLen);
    int64_t isFromInit();
    void callValue(uint8_t val[32]);
    void address(uint8_t hash[20]);
    void sha3(const uint8_t *src, size_t srcLen, uint8_t *dest, size_t destLen);
    void ecrecover(const uint8_t *h, const uint8_t *sig, uint8_t *addr);
    
    int64_t getCallerNonce();
	int64_t callTransfer(const uint8_t* to, size_t toLen, uint8_t amount[32]);
}

namespace bcwasm {
    /**
     * @brief Get block hash
     * 
     * @param number Block height
     * @return h256 Block hash
     */
    h256 blockHash(int64_t number){
        byte hash[32];
        ::blockHash(number, hash);
        return h256(hash, sizeof(hash));
    }

    /**
     * @brief Current block miner’s address
     * 
     * @return h160 
     */
    h160 coinbase(){
        byte hash[20];
        ::coinbase(hash);
        return h160(hash, sizeof(hash));
    }

    /**
     * @brief Get the balance
     * 
     * @return u256 balance
     */
    u256 balance() {
        byte amount[32];
        ::balance(amount);
        return fromBigEndian<u256>(amount);
    }

    /**
     * @brief Sender of the transaction (full call chain)
     * 
     * @return h160 
     */
    h160 origin(){
        byte hash[20];
        ::origin(hash);
        return h160(hash, sizeof(hash));
    }

    /**
     * @brief Caller address
     *
     * @return h160
     */
    h160 caller(){
        byte hash[20];
        ::caller(hash);
        return h160(hash, sizeof(hash));
    }

    /**
     * @brief whether account is contract's creator
     *
     * @return int64_t
     */
    int64_t isOwner(const Address &contract, const Address &account)
    {
        return ::isOwner(contract.data(), contract.size(), account.data(), account.size());
    }

    /**
     * @brief Caller address
     *
     * @return h160
     */
    int64_t isFromInit()
    {
        return ::isFromInit();
    }

    /**
     * @brief Number of wei sent with the message
     * 
     * @return u256 
     */
    u256 callValue()
    {
        byte val[32];
        ::callValue(val);
        return fromBigEndian<u256>(val);
    }

    /**
     * @brief Contract address
     * 
     * @return h160 
     */
    h160 address(){
        byte hash[20];
        ::address(hash);
        return h160(hash, sizeof(hash));
    }

    /**
     * @brief Compute the Keccak-256 hash of the input
     * 
     * @param data Binary data
     * @return h256 
     */
    h256 sha3(bytes & data) {
        byte hash[32];
        ::sha3(data.data(), data.size(), hash, sizeof(hash));
        return h256(hash, sizeof(hash));
    }

    /**
     * @brief Compute the Keccak-256 hash of the input
     * 
     * @param data String
     * @return h256 
     */
    h256 sha3(const std::string &data) {
        byte hash[32];
        ::sha3((const byte*)data.data(), data.length(), hash, sizeof(hash));
        return h256(hash, sizeof(hash));
    }


    /**
     * @brief Compute the Keccak-256 hash of the input
     * 
     * @param data Byte pointer
     * @param len Length
     * @return h256 
     */
    h256 sha3(const byte *data, size_t len) {
        byte hash[32];
        ::sha3(data, len, hash, sizeof(hash));
        return h256(hash, sizeof(hash));
    }

    /**
     * @brief Send given amount of Wei to Address
     * 
     * @param to Destination address
     * @param amount Amount
     * @return int64_t 0 success non-zero failure
     */
    int64_t callTransfer(const Address& to, u256 amount) {
        bytes bs(32);
        toBigEndian(amount, bs);
        return ::callTransfer(to.data(), to.size(), bs.data());
    }

    /**
     * @brief ecrecover
     * @param h, hash of the message
     * @sig, signature of the message
     * @return h160 
     */
    h160 ecrecover(const byte* h, const byte* sig){
        byte addr[20];
        ::ecrecover(h,sig,addr);
        return h160(addr, sizeof(addr));
    }
    h160 ecrecover(bytes & h, bytes & sig){
        byte addr[20];
        ::ecrecover(h.data(),sig.data(),addr);
        return h160(addr, sizeof(addr));
    }
}


