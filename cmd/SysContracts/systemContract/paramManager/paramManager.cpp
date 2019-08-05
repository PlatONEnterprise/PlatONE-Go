//auto create contract
#include <stdlib.h>
#include <string.h>
#include <string>
#include <bcwasm/bcwasm.hpp>

#include <rapidjson/document.h>
#include <rapidjson/writer.h>
#include <rapidjson/stringbuffer.h>

#include "../common/gdef.hpp"
#include "../common/util.hpp"

using namespace std;

namespace SystemContract
{
    char gasContractNameKey[] = "GasContractName";
    char isProduceEmptyBlockKey[] = "IsProduceEmptyBlock";
    char txGasLimitKey[] = "TxGasLimit";
    char blockGasLimitKey[] = "BlockGasLimit";
    char isAllowAnyAccountDeployContractKey[] = "IsAllowAnyAccountDeployContract";
    char isCheckContractDeployPermission[]="isCheckContractDeployPermission";
    char isApproveDeployedContractKey[] = "IsApproveDeployedContract";
    char isTxUseGasKey[] = "IsTxUseGas";
    char cbftTimeParamKey[] = "CBFTTimeParam";

    /* txGasLimitMinValue 需要保证系统合约能够正常部署和调用，尤其是 SetTxGasLimit 和 SetBlockGasLimit 这两个调用要能正常调用
     * 12771596 为所有系统合约部署及上述两个接口调用所需 gas 的最大值（按照当前的 gas table），txGasLimitMinValue 的取值再放大到 100 倍
     *
     * txGasLimitMaxValue 目前定为 2s 对应的 gas 值
     *
     * txGasLimitDefaultValue 目前定为 1.5s 对应的 gas 值
     */
    const unsigned long long txGasLimitMinValue = 12771596*100; // 12771596 大致相当于 0.012772s
    const unsigned long long txGasLimitMaxValue = 2e9;          // 相当于 2s
    const unsigned long long txGasLimitDefaultValue = 1.5e9;          // 相当于 1.5s


    /* blockGasLimitMinValue 需要保证系统合约能够正常部署和调用，尤其是 SetTxGasLimit 和 SetBlockGasLimit 这两个调用要能正常调用
     * 12771596 为所有系统合约部署及上述两个接口调用所需 gas 的最大值（按照当前的 gas table），blockGasLimitMinValue 的取值再放大到 100 倍
     *
     * blockGasLimitMaxValue 目前定为 20s 对应的 gas 值
     *
     * blockGasLimitDefaultValue 目前定为 10s 对应的 gas 值
     */
    const unsigned long long blockGasLimitMinValue = 12771596*100; // 12771596 大致相当于 0.012772s
    const unsigned long long blockGasLimitMaxValue = 2e10;         // 相当于 20s
    const unsigned long long blockGasLimitDefaultValue = 1e10;         // 相当于 10s

    const int produceDurationMaxValue = 60; 
    const int produceDurationDefaultValue = 10; 

    const int blockIntervalMinValue = 1; 
    const int blockIntervalDefaultValue = 1; 

    // 03.08测试发现，如果在init()中设置string类型的状态变量的默认值为空串时，此时直接调用get接口会导致 runtime error: slice bound out of range
    // 测试发现，string类型的状态变量的默认值就是空串
    // const char* GasContractName_DefaultValue = "";


    struct CBFTTimeParam
    {
        int produceDuration;
        int blockInterval;
        BCWASM_SERIALIZE(CBFTTimeParam, (produceDuration)(blockInterval));
    };


    class ParamManager : public bcwasm::Contract
    {
    public:
        ParamManager() {}

        /// 实现父类: bcwasm::Contract 的虚函数
        /// 该函数在合约首次发布时执行，仅调用一次
        void init()
        {
            // register paramManager to cnsManager
            util::registerContractFromInit(gdef::paramManager, "1.0.0.0");

            /* set default values for params */
            struct CBFTTimeParam cbftTimeParam;
            cbftTimeParam.produceDuration = produceDurationDefaultValue;
            cbftTimeParam.blockInterval = blockIntervalDefaultValue;
            bcwasm::setState(cbftTimeParamKey, cbftTimeParam); // produceDuration 和 blockInterval 的默认值

            bcwasm::setState(txGasLimitKey, txGasLimitDefaultValue); // txGasLimit 的默认值
            bcwasm::setState(blockGasLimitKey, blockGasLimitDefaultValue); // blockGasLimit 的默认值

            bcwasm::println("ParamManager init success...");
        }

        /// 定义Event
        BCWASM_EVENT(Notify, uint64_t, const char *)
        enum Code
        {
            SUCCESS = 0,
            FAILURE
        };

    public:
        // 设置作为gas消耗的合约名称
        // @contractName: 
        // 特定合约代币的名称，根据此名称可以从合约管理合约获取地址
        //如果为空，表示不消耗特定合约代币作为gas
        int setGasContractName(const char *contractName)
        {
            if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set GasContractName failure, no permission");
                return FAILURE; // 设置不成功
            }

            if(nullptr == contractName)
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "contractName is invalid: null pointer");
                return FAILURE; // 设置不成功
            }

            bcwasm::setState(gasContractNameKey, string(contractName));

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set GasContractName success");
            return SUCCESS; // 设置成功
        }

        // 获取作为gas消耗的合约名称
        const char * getGasContractName() const
        {
            string contractName;
            bcwasm::getState(gasContractNameKey, contractName);
            
            return util::makeReturnedStr(contractName);
        }

        // 设置每个节点出块时长和相邻区块间隔
        // @produceDuration:  每个节点出块时长
        // @blockInterval:        相邻区块间隔
        // produceDuration 和 blockInterval 都为整数
        // blockInterval 大于等于1,小于或等于 ProduceDuration，且可整除 produceDuration
        // ProduceDuration小于或等于60
        int setCBFTTimeParam(int produceDuration, int blockInterval)
        {
            if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set CBFT time param failure, no permission");
                return FAILURE; // 设置不成功
            }

            if(blockInterval < blockIntervalMinValue) // blockInterval 需大于等于1
            {
                bcwasm::println("blockInterval should be greater than 1");
                BCWASM_EMIT_EVENT(Notify, FAILURE, "blockInterval should be greater than 1");
                return FAILURE;
            }

             if(blockInterval > produceDuration) // blockInterval 需小于或等于 produceDuration
            {
                bcwasm::println("blockInterval should not be greater than produceDuration");
                BCWASM_EMIT_EVENT(Notify, FAILURE, "blockInterval should not be greater than produceDuration");
                return FAILURE;
            }

             if(produceDuration > produceDurationMaxValue) // produceDuration 需小于或等于60
            {
                bcwasm::println("produceDuration should not be greater than 60");
                BCWASM_EMIT_EVENT(Notify, FAILURE, "produceDuration should not be greater than 60");
                return FAILURE; 
            }

            if(produceDuration % blockInterval != 0) // blockInterval 需整除 produceDuration
            {
                bcwasm::println("blockInterval should divide produceDuration");
                BCWASM_EMIT_EVENT(Notify, FAILURE, "blockInterval should divide produceDuration");
                return FAILURE;
            }
            
            struct CBFTTimeParam cbftTimeParam;
            cbftTimeParam.produceDuration = produceDuration;
            cbftTimeParam.blockInterval = blockInterval;
        
            bcwasm::setState(cbftTimeParamKey, cbftTimeParam);

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set CBFT time param success");
            return SUCCESS;
        }

        // 获取CBFT的时间参数：每个节点出块时长和相邻区块间隔
        // 返回JSON格式：{"ProduceDuration": xxx, "BlockInterval": xxx}
        const char *getCBFTTimeParam() const
        {
            rapidjson::Document jsonDoc;                                                                     //生成一个dom元素Document
            rapidjson::Document::AllocatorType &allocator = jsonDoc.GetAllocator();   //获取分配器
            jsonDoc.SetObject();                                                                                 //将当前的Document设置为一个object，也就是说，整个Document是一个Object类型的dom元素

            struct CBFTTimeParam cbftTimeParam;
            bcwasm::getState(cbftTimeParamKey, cbftTimeParam);

            //添加属性
            jsonDoc.AddMember("ProduceDuration", cbftTimeParam.produceDuration, allocator); 
            jsonDoc.AddMember("BlockInterval", cbftTimeParam.blockInterval, allocator);

            //生成字符串
            rapidjson::StringBuffer buffer;
            rapidjson::Writer<rapidjson::StringBuffer> writer(buffer);
            jsonDoc.Accept(writer);

            std::string strJson = buffer.GetString();
            return util::makeReturnedStr(strJson);
        }

        // 本参数根据最新的讨论（2019.03.06之前）不再需要，即需要出空块，这个机制结合 CBFT 的间接确认机制，用来保证共识的 liveness
        // 设置是否出空块
        // @isProduceEmptyBlock: 
        // 1: 出空块  0：不出空块
        int setIsProduceEmptyBlock(int isProduceEmptyBlock)
        {
            if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set isProduceEmptyBlock failure, no permission");
                return FAILURE; // 设置不成功 
            }

            if(0 != isProduceEmptyBlock && 1 != isProduceEmptyBlock)
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set isProduceEmptyBlock failure: param should be 0 or 1");
                return FAILURE;
            }

            bcwasm::setState(isProduceEmptyBlockKey, isProduceEmptyBlock);

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set isProduceEmptyBlock success");
            return SUCCESS;
        }

        int getIsProduceEmptyBlock() const
        {
            int isProduceEmptyBlock = 0;
            bcwasm::getState(isProduceEmptyBlockKey, isProduceEmptyBlock);

            return isProduceEmptyBlock;
        }

        // 设置每笔交易 gaslimit
        // @txGasLimit: 每笔交易 gaslimit
        int setTxGasLimit(unsigned long long txGasLimit)
        {
             if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set txGasLimit failure, no permission");
                return FAILURE;  // 设置不成功 
            }

            if(txGasLimit < txGasLimitMinValue || txGasLimit > txGasLimitMaxValue)
            {
                bcwasm::println("txGasLimit value is beyond the limitation");
                BCWASM_EMIT_EVENT(Notify, FAILURE, "txGasLimit value is beyond the limitation");
                return FAILURE;
            }

            // 获取区块 gas limit，其值应大于或等于每笔交易 gas limit 参数的值
            unsigned long long blockGasLimit;
            bcwasm::getState(blockGasLimitKey, blockGasLimit);

            if(txGasLimit > blockGasLimit)
            {
                bcwasm::println("txGasLimit value should not be greater than blockGasLimit");
                BCWASM_EMIT_EVENT(Notify, FAILURE, "txGasLimit value should not be greater than blockGasLimit");
                return FAILURE;
            }

            bcwasm::setState(txGasLimitKey, txGasLimit);

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set txGasLimit success");
            return SUCCESS;
        }

        // 获取每笔交易 gaslimit
        unsigned long long getTxGasLimit() const
        {
            unsigned long long txGasLimit;
            bcwasm::getState(txGasLimitKey, txGasLimit);

            return txGasLimit;
        }

        // 设置区块 gaslimit
        // @blockGasLimit: 区块 gaslimit
        int setBlockGasLimit(unsigned long long blockGasLimit)
        {
             if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set blockGasLimit failure, no permission");
                return FAILURE; // 设置不成功
            }
 
            if(blockGasLimit < blockGasLimitMinValue || blockGasLimit > blockGasLimitMaxValue)
            {
                bcwasm::println("blockGasLimit value is beyond the limitation");
                BCWASM_EMIT_EVENT(Notify, FAILURE, "blockGasLimit value is beyond the limitation");
                return FAILURE;
            }

            unsigned long long txGasLimit;
            bcwasm::getState(txGasLimitKey, txGasLimit);
            if(txGasLimit > blockGasLimit)
            {
                bcwasm::println("txGasLimit value should not be greater than blockGasLimit");
                BCWASM_EMIT_EVENT(Notify, FAILURE, "txGasLimit value should not be greater than blockGasLimit");
                return FAILURE;
            }

            bcwasm::setState(blockGasLimitKey, blockGasLimit);

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set blockGasLimit success");
            return SUCCESS;
        }

        // 获取区块 gaslimit
        unsigned long long getBlockGasLimit() const
        {
            unsigned long long blockGasLimit;
            bcwasm::getState(blockGasLimitKey, blockGasLimit);

            return blockGasLimit;
        }

        // 设置是否允许任意用户部署合约
        // @isAllowAnyAccountDeployContract:
        // 0: 允许任意用户部署合约  1: 用户具有相应权限才可以部署合约
        // 默认为0，即允许任意用户部署合约
        int setAllowAnyAccountDeployContract(int isAllowAnyAccountDeployContract)
        {
             if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set isAllowAnyAccountDeployContract failure, no permission");
                return FAILURE; // 设置不成功
            }

            if(0 != isAllowAnyAccountDeployContract && 1 != isAllowAnyAccountDeployContract)
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set AllowAnyAccountDeployContract failure: param should be 0 or 1");
                return FAILURE;
            }

            bcwasm::setState(isAllowAnyAccountDeployContractKey, isAllowAnyAccountDeployContract);

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set isAllowAnyAccountDeployContract success");
            return SUCCESS;
        }

        // 设置是否检查合约部署权限
        // 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
        // 默认为0，不检查合约部署权限，即允许任意用户部署合约
        int setCheckContractDeployPermission(int checkPermission)
        {
             if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set isCheckContractDeployPermission failure, no permission");
                return FAILURE; // 设置不成功
            }

            if(0 != checkPermission && 1 != checkPermission)
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set AllowAnyAccountDeployContract failure: param should be 0 or 1");
                return FAILURE;
            }

            bcwasm::setState(isCheckContractDeployPermission, checkPermission);

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set isCheckContractDeployPermission success");
            return SUCCESS;
        }
        // 获取是否是否检查合约部署权限
        // 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
        // 默认为0，不检查合约部署权限，即允许任意用户部署合约        
        int getCheckContractDeployPermission() const
        {
            int checkPermission = 0;
            bcwasm::getState(isCheckContractDeployPermission, checkPermission);

            return checkPermission;
        }

        // 获取是否允许任意用户部署合约的标志
        int getAllowAnyAccountDeployContract() const
        {
            int isAllowAnyAccountDeployContract = 0;
            bcwasm::getState(isAllowAnyAccountDeployContractKey, isAllowAnyAccountDeployContract);

            return isAllowAnyAccountDeployContract;
        }

        // 设置是否审核已部署的合约
        // @isApproveDeployedContract:
        // 1: 审核已部署的合约  0: 不审核已部署的合约
        int setIsApproveDeployedContract(int isApproveDeployedContract)
        {
             if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set isApproveDeployedContract failure, no permission");
                return FAILURE; // 设置不成功
            }

            if(0 != isApproveDeployedContract && 1 != isApproveDeployedContract)
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set isApproveDeployedContract failure: param should be 0 or 1");
                return FAILURE;
            }

            bcwasm::setState(isApproveDeployedContractKey, isApproveDeployedContract);

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set isApproveDeployedContract success");
            return SUCCESS;
        }

        // 获取是否审核已部署的合约的标志
        int getIsApproveDeployedContract() const
        {
            int isApproveDeployedContract = 0;
            bcwasm::getState(isApproveDeployedContractKey, isApproveDeployedContract);

            return isApproveDeployedContract;
        }

        // 本参数根据最新的讨论（2019.03.06之前）不再需要，即交易需要消耗gas。但是计费相关如消耗特定合约代币的参数由 setGasContractName 进行设置
        // 设置交易是否消耗 gas
        // @isTxUseGas:
        // 1:  交易消耗 gas  0: 交易不消耗 gas
        int setIsTxUseGas(int isTxUseGas)
        {
             if(!hasPermission())
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set isTxUseGas failure, no permission");
                return FAILURE; // 设置不成功
            }

            if(0 != isTxUseGas && 1 != isTxUseGas)
            {
                BCWASM_EMIT_EVENT(Notify, FAILURE, "set isTxUseGas failure: param should be 0 or 1");
                return FAILURE;
            }

            bcwasm::setState(isTxUseGasKey, isTxUseGas);

            BCWASM_EMIT_EVENT(Notify, SUCCESS, "set isTxUseGas success");
            return SUCCESS;
        }

        // 获取交易是否消耗 gas
        int getIsTxUseGas() const
        {
            int isTxUseGas = 0;
            bcwasm::getState(isTxUseGasKey, isTxUseGas);

            return isTxUseGas;
        }

    private:
        // 判断用户是否有权限设置参数管理合约中管理的参数
     	bool hasPermission()
        {
           int userStatus = util::getUserStatus(); // 获取用户状态
                if (0 != userStatus)
                {
                     bcwasm::println("userStatus: ", userStatus);
                    return false; // 用户状态不正确
                }

                vector<string> rolesList;
                util::getRoles(rolesList); // 获取用户角色列表

                if(rolesList.empty())
                {
                    bcwasm::println("Role list query result is empty!");
                    return false;
                }

                for(vector<string>::iterator iter = rolesList.begin(); iter != rolesList.end(); iter++)
                {
                    bcwasm::println("paramManager: got role: ", (*iter));
                    if(util::CHAIN_CREATOR == (*iter) ||  util::CHAIN_ADMIN == (*iter))
                    {
                        return true; // 只有链创建者和管理员才有权限设置参数管理合约中管理的参数
                    }
                        
                }

                return false;

       }

    };
} // namespace SystemContract

// 此处定义的函数会生成ABI文件供外部调用
BCWASM_ABI(SystemContract::ParamManager, setGasContractName)
BCWASM_ABI(SystemContract::ParamManager, getGasContractName)
BCWASM_ABI(SystemContract::ParamManager, setCBFTTimeParam)
BCWASM_ABI(SystemContract::ParamManager, getCBFTTimeParam)
BCWASM_ABI(SystemContract::ParamManager, setIsProduceEmptyBlock)
BCWASM_ABI(SystemContract::ParamManager, getIsProduceEmptyBlock)
BCWASM_ABI(SystemContract::ParamManager, setTxGasLimit)
BCWASM_ABI(SystemContract::ParamManager, getTxGasLimit)
BCWASM_ABI(SystemContract::ParamManager, setBlockGasLimit)
BCWASM_ABI(SystemContract::ParamManager, getBlockGasLimit)
BCWASM_ABI(SystemContract::ParamManager, setAllowAnyAccountDeployContract)
BCWASM_ABI(SystemContract::ParamManager, setCheckContractDeployPermission)
BCWASM_ABI(SystemContract::ParamManager, getCheckContractDeployPermission)
BCWASM_ABI(SystemContract::ParamManager, getAllowAnyAccountDeployContract)
BCWASM_ABI(SystemContract::ParamManager, setIsApproveDeployedContract)
BCWASM_ABI(SystemContract::ParamManager, getIsApproveDeployedContract)
BCWASM_ABI(SystemContract::ParamManager, setIsTxUseGas)
BCWASM_ABI(SystemContract::ParamManager, getIsTxUseGas)
//bcwasm autogen begin
extern "C" { 
int setGasContractName(const char * contractName) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setGasContractName(contractName);
}
const char * getGasContractName() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getGasContractName();
}
int setCBFTTimeParam(int produceDuration,int blockInterval) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setCBFTTimeParam(produceDuration,blockInterval);
}
const char * getCBFTTimeParam() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getCBFTTimeParam();
}
int setIsProduceEmptyBlock(int isProduceEmptyBlock) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setIsProduceEmptyBlock(isProduceEmptyBlock);
}
int getIsProduceEmptyBlock() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getIsProduceEmptyBlock();
}
int setTxGasLimit(unsigned long long txGasLimit) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setTxGasLimit(txGasLimit);
}
unsigned long long getTxGasLimit() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getTxGasLimit();
}
int setBlockGasLimit(unsigned long long blockGasLimit) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setBlockGasLimit(blockGasLimit);
}
unsigned long long getBlockGasLimit() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getBlockGasLimit();
}
int setAllowAnyAccountDeployContract(int isAllowAnyAccountDeployContract) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setAllowAnyAccountDeployContract(isAllowAnyAccountDeployContract);
}
int setCheckContractDeployPermission(int checkPermission) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setCheckContractDeployPermission(checkPermission);
}
int getCheckContractDeployPermission() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getCheckContractDeployPermission();
}
int getAllowAnyAccountDeployContract() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getAllowAnyAccountDeployContract();
}
int setIsApproveDeployedContract(int isApproveDeployedContract) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setIsApproveDeployedContract(isApproveDeployedContract);
}
int getIsApproveDeployedContract() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getIsApproveDeployedContract();
}
int setIsTxUseGas(int isTxUseGas) {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.setIsTxUseGas(isTxUseGas);
}
int getIsTxUseGas() {
SystemContract::ParamManager ParamManager_bcwasm;
return ParamManager_bcwasm.getIsTxUseGas();
}
void init() {
SystemContract::ParamManager ParamManager_bcwasm;
ParamManager_bcwasm.init();
}

}
//bcwasm autogen end