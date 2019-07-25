//auto create contract
//auto create contract

#define ENABLE_TRACE
#include <map>

#include "../../bcwasmlib/include/bcwasm/event.hpp"
#include "../../bcwasmlib/include/bcwasm/contract.hpp"
#include "../../bcwasmlib/include/bcwasm/state.hpp"
#include "../../bcwasmlib/include/bcwasm/serialize.hpp"
#include "../../bcwasmlib/include/bcwasm/storage.hpp"

#include "../common/util.hpp"

using namespace bcwasm;

namespace SystemContract {
    class MultiSig : public Contract {
    private:
        const uint8_t kMaxOwnerCount = 50;
        const Address ZeroAddress;
    private:
        BCWASM_EVENT(Confirmation, const char*, int64_t)
        BCWASM_EVENT(Revocation, const char*, int64_t)
        BCWASM_EVENT(Submission, int64_t)
        BCWASM_EVENT(Execution, const char*)
        BCWASM_EVENT(ExecutionFailure, const char*)
        BCWASM_EVENT(Deposit, const char*, const char*)
        BCWASM_EVENT(OwnerAddition, int64_t)
        BCWASM_EVENT(OwnerRemoval, int64_t)
        BCWASM_EVENT(RequirementChange, int64_t)

        struct Transaction {
            Address destination;
            Address from;
            u256 value;
            int64_t time;
            u256 fee;
            std::string data;
            bool executed;
            bool pending;
            BCWASM_SERIALIZE(Transaction, (destination)(from)(value)(time)(fee)(data)(executed)(pending))
        };

        struct TxId {
            TxId(uint64_t i) : id(i){}
            uint64_t id;
            const std::string type = "transaction";
            BCWASM_SERIALIZE(TxId, (id)(type))
        };
        struct ConfirmID {
            ConfirmID(uint64_t i) : id(i){}
            uint64_t id;
            const std::string type = "confirm";
            BCWASM_SERIALIZE(ConfirmID, (id)(type))
        };

#define REQUIRED  "required"
#define TXCOUNT "transactionCount"
#define OWNERS "owners"
#define INIT "init"

//        std::map<int64_t, Transaction> transactions_;
//        std::map<int64_t, std::map<Address, bool>> confirmations_;
//        std::map<Address, bool> isOwner_;
//        std::vector<Address> owners_;
//        uint64_t required_;
//        uint64_t transactionCount_;



    public:
        MultiSig(){}
        void init(){
        }

    private:
        void assertWallet() {
            BCWasmAssert(origin() == address(), "only wallet");
        }

        void ownerDoesNotExist(Address owner) {
            bool exist = false;
            getState(owner, exist);
            BCWasmAssert(exist, "");
        }

        void ownerExists(Address owner) {
            bool exist = false;
            getState(owner, exist);
            println("owner is exist:", exist ? "true":"false");
            BCWasmAssert(exist, "");
        }

        void transactionExists(uint64_t transactionId) {
            TxId txid(transactionId);
            Transaction tx;
            getState(txid, tx);
            BCWasmAssert(tx.destination != ZeroAddress, "dest:", tx.destination.toString());
        }

        bool confirmed(uint64_t transactionId, Address owner) {
            println("confirm transactionid:", transactionId);
            ConfirmID cid(transactionId);
            std::map<Address, bool> addrMap;
            getState(cid, addrMap);
            for (auto& kv : addrMap) {
                println("addrMap first:", kv.first.toString(), " second:", kv.second ? "true" : "false");
            }

            if (addrMap.find(owner) == addrMap.end()) {
                return false;
            }

            return addrMap[owner];
        }

        void notConfirmed(uint64_t transactionId, Address owner) {
            ConfirmID cid(transactionId);
            std::map<Address, bool> addrMap;
            getState(cid, addrMap);
            if (addrMap.find(owner) == addrMap.end()) { return; }
            println("addrMap:", owner.toString(), addrMap[owner] ? "true" : "false");
            BCWasmAssert(!addrMap[owner], "");
        }

        void notExecuted(uint64_t transactionId) {
            TxId txid(transactionId);
            Transaction tx;
            getState(txid, tx);
            println("txid:", transactionId, tx.executed ? "true" : "false");
            BCWasmAssert(!tx.executed, "");
        }

        void notNull(Address address) {
            println("address:", address.toString());
            BCWasmAssert(address != ZeroAddress, "");
        }

        void validRequirement(uint64_t ownerCount, uint64_t required) {
            TRACE("ownerCount:", ownerCount, "required:", required);
            // 根据BCWASM 0.1版本的需求，允许required = ownerCount
            // 修正如下条件为所有条件同时成立
            BCWasmAssert((ownerCount <= kMaxOwnerCount && required <= ownerCount && required != 0 && ownerCount >= 2), "");
        }

        void payable() {
            u256 value = callValue();
            TRACE("value:", value);
            if (value > 0) {
                BCWASM_EMIT_EVENT(Deposit, caller().toString().c_str(), value.convert_to<std::string>().c_str());
            }
        }

    public:
        bool isInitWallet() {
            uint32_t init = 0;
            getState(INIT, init);
            return init == 1;
        }
        void initWallet(const char *owner, uint64_t required) {
            BCWasmAssert(!isInitWallet());
            setState(INIT, (uint32_t)1);
            println("init wallet owner:", owner, "required:", required);
            std::vector<std::string> addresses;
            split(owner, addresses, ":");
            println("addresses size:",addresses.size());
            validRequirement(addresses.size(), required);
            std::vector<Address> ownerAddr;
            ownerAddr.reserve(addresses.size());
            for (size_t i = 0; i < addresses.size(); i++) {
                println("address:", addresses[i]);
                bool exist = false;
                Address addr(addresses[i], true);
                println("address convert:", addr.toString());
                getState(addr, exist);
                BCWasmAssert(exist == false, "");
                setState(addr, true);
                println("push_back:", addr.toString());
                ownerAddr.push_back(addr);
            }
            setState(OWNERS, ownerAddr);
            setState(REQUIRED, required);
            println("end initWallet");


            std::vector<Address> testAddr;
            getState(OWNERS, testAddr);
            for (size_t i = 0; i< testAddr.size(); i++) {
                println("test:", i, " addr:", testAddr[i].toString());
            }
        }


        uint64_t submitTransaction(const char *destination, const char *from, const char *vs, const char *data, uint64_t len,uint64_t time, const char *fs) {
            println("input args", destination, from, vs, len, time, fs);
            u256 value(vs);
            u256 fee(fs);
            uint64_t transactionId = addTransaction(destination, from, value, data, len, time, fee);
            println("submitTransaction transactionId:", transactionId);
            return transactionId;
        }

        void confirmTransaction(uint64_t transactionId) {
            Address sender = origin();
            println("sender:", sender.toString(), "transactionId:", transactionId, "blocknum:", ::number());
            ownerExists(sender);
            transactionExists(transactionId);
            notConfirmed(transactionId, sender);

            std::map<Address, bool> confirms;
            ConfirmID cid(transactionId);
            getState(cid, confirms);
            confirms[sender] = true;
            setState(cid, confirms);
            BCWASM_EMIT_EVENT(Confirmation, sender.toString().c_str(), transactionId);
            executeTransaction(transactionId);
        };

        // a confirmation can be revoked only when:
        // a) the transaction is pending
        // b) the transaction has not been executed
        void revokeConfirmation(uint64_t transactionId) {
            Address sender = origin();
            ownerExists(sender);
            
            /*
            if (confirmed(transactionId, sender)) {
                TRACE("alread confirmed");
                return;
            }
            */

            notExecuted(transactionId);

            Transaction tx;
            getState(TxId(transactionId), tx);
            if (!tx.pending) { // only when transaction is pending, can revoke a confirmation
                return;
            }

            ConfirmID cid(transactionId);
            std::map<Address, bool> confirms;
            getState(cid, confirms);

            auto it = confirms.find(sender);
            if(it == confirms.end()) {
                return;
            }

            confirms[sender] = false;
            setState(cid, confirms);
            println("Multsigtx.pending:",toString(tx.pending));
            if(isConfirmed(transactionId)){
                println("isConfirmed:", toString(isConfirmed(transactionId)));
                tx.pending = false;
                setState(TxId(transactionId), tx);
            }
            BCWASM_EMIT_EVENT(Revocation, sender.toString().c_str(), transactionId);
        }

        void executeTransaction(uint64_t transactionId) {
            notExecuted(transactionId);

            if (isConfirmed(transactionId)) {
                println("confirm success id:", transactionId);
                Transaction tx;
                getState(TxId(transactionId), tx);
                tx.executed = true;
                tx.pending = false;
                //transfer
                std::string addr = tx.destination.toString();
                int res = callTransfer(Address(addr), tx.value);
                //未来返回值1是失败，0是成功，待底层修改后再来修改合约
                if (res == 0){
                    BCWASM_EMIT_EVENT(Execution, "Execution");
                } else {
                    BCWASM_EMIT_EVENT(ExecutionFailure, "ExecutionFailure");
                    tx.executed = false;
                }
                setState(TxId(transactionId), tx);
            }
        }

        int isConfirmed(uint64_t transactionId) const {
            uint64_t count = 0;
            std::vector<Address> owners;
            getState(OWNERS, owners);
            std::map<Address, bool> confirms;
            getState(ConfirmID(transactionId), confirms);
            for (size_t i = 0; i < owners.size(); i++){
                if (confirms[owners[i]]) {
                    count += 1;
                }
                uint64_t required = 0;
                getState(REQUIRED, required);
                if (count == required) {
                    return true;
                }
            }
            return false;
        }

        uint64_t getRequired() const {
            uint64_t required = 0;
            getState(REQUIRED, required);
            println("getRequired required:", required);
            return required;
        }

         uint64_t getListSize() const {
             uint64_t transactionCount = 0;
             getState(TXCOUNT, transactionCount);
             println("getTransactionCount count:", transactionCount);
             return transactionCount;
         }


        uint64_t addTransaction(const char *destination, const char *from, u256 value, const char *data, uint64_t len, uint64_t time, u256 fee){
            println("addTransaction");
            Address dest(destination, true);
            notNull(dest);
            uint64_t transactionId = 0;
            getState(TXCOUNT, transactionId);
            println("get transaction id:", transactionId);
            Transaction transaction;
            transaction.destination = dest;
            transaction.from = Address(std::string(from), true);
            transaction.value = value;
            transaction.time = time;
            transaction.fee = fee;
            transaction.data.append(data, len);
            transaction.executed = false;
            transaction.pending = true;
            TxId txId(transactionId);

            setState(txId, transaction);
            setState(TXCOUNT, transactionId+1);
            println("emit event: Submission", transactionId);
            BCWASM_EMIT_EVENT(Submission, transactionId);
            return transactionId;
        }

        uint64_t getConfirmationCount(uint64_t transactionId) const {
            println("getConfirmationCount:", transactionId);
            // std::vector<Address> owners;
            // getState(OWNERS, owners);
            uint64_t count = 0;

            ConfirmID cid(transactionId);
            std::map<Address, bool> confirms;

            getState(cid, confirms);
            
            /*
            for (size_t i = 0; i < owners.size(); i++) {
                if (confirms[owners[i]]){
                    TRACE("is confirm ", owners[i].toString());
                    count += 1;
                }
                TRACE("is not confirm ", owners[i].toString());
            }
            */

            for (std::map<Address, bool>::iterator iter = confirms.begin(); iter != confirms.end(); iter++) {
                    if (iter->second) {
                        println("is confirm:", "0x" + iter->first.toString());
                        ++count;
                    }
                }

            println("getConfirmationCount count:", count);
            return count;
        }

        uint64_t getTransactionCount(int pending, int executed) const {
            println("getTransactionCount pending:", pending, "executed:", executed);
            uint64_t count = 0;
            uint64_t transactionCount = 0;
            getState(TXCOUNT, transactionCount);
            for (size_t i = 0; i < transactionCount; i++) {
                Transaction transaction;
                getState(TxId(i), transaction);
                println("getTransactionCount :", i, "pending:", pending, "executed:", executed ? "true":"false");
                if (pending && !transaction.executed
                    || executed && transaction.executed){
                    count++;
                }
            }
            println("getTransactionCount count:", count);
            return count;
        }

        const char *getTransactionList(uint64_t from, uint64_t to) const {
        uint64_t transactionCount = 0;
             getState(TXCOUNT, transactionCount);
             if (from > to) { return ""; }
             std::string result;
             size_t end = transactionCount > to ? to : transactionCount;
             for(size_t i = from; i < end; i++){
                 println("get transaction:", i);
                 Transaction transaction;
                 if (getState(TxId(i), transaction) == 0) {
                     println("get transaction:", i, "failed");
                     break;
                 }

                 println("from:",transaction.from.toString(), "dest:", transaction.destination.toString() ,"value:", transaction.value, "time:", transaction.time, "fee:", transaction.fee, "pending:", transaction.pending, "executed:", transaction.executed);
                 result.append(transaction.from.toString())
                         .append("|").append(transaction.destination.toString())
                         .append("|").append(transaction.value.convert_to<std::string>())
                         .append("|").append(toString(transaction.time))
                         .append("|").append(transaction.data)
                         .append("|").append(transaction.fee.convert_to<std::string>())
                         .append("|").append(toString(transaction.pending))
                         .append("|").append(toString(transaction.executed))
                         .append("|").append(toString(i))
                         .append(":");
             }

             println("getTransactionList:", result);
             return util::makeReturnedStr(result);
         }


        const char *getOwners()const {
            std::vector<Address> owners;
            getState(OWNERS, owners);
            println("owners size:", owners.size());
            std::string address;
            for (size_t i = 0; i < owners.size(); i++) {
                println("owner[", i, "]:", owners[i].toString());
                if (i != 0){
                    address += ":";
                }
                address += "0x" + owners[i].toString();
            }
            println("owners:", address);
            return util::makeReturnedStr(address);
        }

        const char * getConfirmations(uint64_t transactionId) const {
            println("getConfirmations id:", transactionId);
            // std::vector<Address> owners;
            // getState(OWNERS, owners);

            // uint64_t count = 0;
            ConfirmID cid(transactionId);
            std::map<Address, bool> confirms;
            std::string address;
            getState(cid, confirms);

            for (std::map<Address, bool>::iterator iter = confirms.begin(); iter != confirms.end(); iter++) {
                    if (iter->second) {
                        println("is confirm:", "0x" + iter->first.toString());
                        if (iter != confirms.begin()) { address += ":"; }
                        address += "0x" + iter->first.toString();
                    }
                }

            println("addresses:", address);
            return util::makeReturnedStr(address);
        }

        const char * getTransactionIds(uint64_t from, uint64_t to, int pending, int executed) const {
            println("from:", from, "to:", to, "pending:", pending, "executed:", executed);
            uint64_t count = 0;
            uint64_t transactionCount = 0;
            getState(TXCOUNT, transactionCount);
            std::string transactionIds;
            for (size_t i = 0; i < transactionCount; i++) {
                Transaction transaction;
                getState(TxId(i), transaction);
                println("transaction.executed:", transaction.executed ? "true":"false");
                if (pending && !transaction.executed
                    || executed && transaction.executed){
                    Transaction transaction;
                    if (i >= from) {
                        transactionIds += ":" + toString(i);
                    }
                    if (i == to) {
                        break;
                    }
                }
            }
            println("transactionIds:", transactionIds);
            return util::makeReturnedStr(transactionIds);
        }

		const char * getMultiSigList(const char *transactionIds)const {
            std::vector<std::string> ids;
            split(transactionIds, ids, ",");
            std::string res;
            println("transactionIds size:", ids.size());
            for (size_t i = 0; i < ids.size(); i++) {
                if (i != 0) { res += "|"; }
                res += ids[i] + ":";
                println("transactionId:", ids[i]);
                ConfirmID cid(stouint64(ids[i]));
                std::map<Address, bool> addrMap;
                getState(cid, addrMap);
                for ( std::map<Address, bool>::iterator iter = addrMap.begin(); iter != addrMap.end(); iter++) {
                    if (iter->second) {
                        if (iter != addrMap.begin()) { res += ","; }
                        res += "0x" + iter->first.toString();
                    }
                }
                res +=  ":";
                for ( std::map<Address, bool>::iterator iter = addrMap.begin(); iter != addrMap.end(); iter++) {
                    if (!iter->second) {
                        if (iter != addrMap.begin()) { res += ","; }
                        res += "0x" + iter->first.toString();
                    }
                }
            }
            
            return util::makeReturnedStr(res);
        }

        uint64_t stouint64(const std::string &num) const {
            uint64_t res = 0;
            for (size_t i = 0; i < num.length(); i++) {
                res = res * 10 + (num[i] - '0');
            }
            bcwasm::println("stouint64", num, "->", res);
            return res;
        }

        std::string toString(uint64_t num) const {
            if (num == 0) { return "0";}
            std::string res;
            while (num != 0) {
                char c = num % 10 + '0';
                num /= 10;
                res.insert(0, 1, c);
            }
            return res;
        }

        int split( const std::string & srcStr, std::vector<std::string> & destArray, const std::string & delimiter ) const {
            if( srcStr.empty() ){
                return 0;
            }
            std::string::size_type startPos = srcStr.find_first_not_of( delimiter );
            size_t lengthOfDelimiter = delimiter.length();
            while( std::string::npos != startPos ){
                std::string::size_type nextPos = srcStr.find( delimiter, startPos );
                std::string str;
                if( std::string::npos != nextPos ){
                    str = srcStr.substr( startPos, nextPos - startPos );
                    nextPos += lengthOfDelimiter;
                }
                else{
                    str = srcStr.substr( startPos );
                }
                startPos = nextPos;
                if( !str.empty() ){
                    destArray.push_back( str );
                }
            }
            return destArray.size();
        }


    };
}
BCWASM_ABI(SystemContract::MultiSig, initWallet)
BCWASM_ABI(SystemContract::MultiSig, confirmTransaction)
BCWASM_ABI(SystemContract::MultiSig, executeTransaction)
BCWASM_ABI(SystemContract::MultiSig, revokeConfirmation)
BCWASM_ABI(SystemContract::MultiSig, submitTransaction)
BCWASM_ABI(SystemContract::MultiSig, isConfirmed)
BCWASM_ABI(SystemContract::MultiSig, getConfirmationCount)
BCWASM_ABI(SystemContract::MultiSig, getTransactionCount)
BCWASM_ABI(SystemContract::MultiSig, getTransactionList)
BCWASM_ABI(SystemContract::MultiSig, getOwners)
BCWASM_ABI(SystemContract::MultiSig, getConfirmations)
BCWASM_ABI(SystemContract::MultiSig, getTransactionIds)
BCWASM_ABI(SystemContract::MultiSig, getRequired)
BCWASM_ABI(SystemContract::MultiSig, getMultiSigList)
BCWASM_ABI(SystemContract::MultiSig, getListSize)
//bcwasm autogen begin
extern "C" { 
void initWallet(const char * owner,unsigned long long required) {
SystemContract::MultiSig MultiSig_bcwasm;
MultiSig_bcwasm.initWallet(owner,required);
}
unsigned long long submitTransaction(const char * destination,const char * from,const char * vs,const char * data,unsigned long long len,unsigned long long time,const char * fs) {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.submitTransaction(destination,from,vs,data,len,time,fs);
}
void confirmTransaction(unsigned long long transactionId) {
SystemContract::MultiSig MultiSig_bcwasm;
MultiSig_bcwasm.confirmTransaction(transactionId);
}
void revokeConfirmation(unsigned long long transactionId) {
SystemContract::MultiSig MultiSig_bcwasm;
MultiSig_bcwasm.revokeConfirmation(transactionId);
}
void executeTransaction(unsigned long long transactionId) {
SystemContract::MultiSig MultiSig_bcwasm;
MultiSig_bcwasm.executeTransaction(transactionId);
}
int isConfirmed(unsigned long long transactionId) {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.isConfirmed(transactionId);
}
unsigned long long getRequired() {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getRequired();
}
unsigned long long getListSize() {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getListSize();
}
unsigned long long getConfirmationCount(unsigned long long transactionId) {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getConfirmationCount(transactionId);
}
unsigned long long getTransactionCount(int pending,int executed) {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getTransactionCount(pending,executed);
}
const char * getTransactionList(unsigned long long from,unsigned long long to) {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getTransactionList(from,to);
}
const char * getOwners() {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getOwners();
}
const char * getConfirmations(unsigned long long transactionId) {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getConfirmations(transactionId);
}
const char * getTransactionIds(unsigned long long from,unsigned long long to,int pending,int executed) {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getTransactionIds(from,to,pending,executed);
}
const char * getMultiSigList(const char * transactionIds) {
SystemContract::MultiSig MultiSig_bcwasm;
return MultiSig_bcwasm.getMultiSigList(transactionIds);
}
void init() {
SystemContract::MultiSig MultiSig_bcwasm;
MultiSig_bcwasm.init();
}

}
//bcwasm autogen end