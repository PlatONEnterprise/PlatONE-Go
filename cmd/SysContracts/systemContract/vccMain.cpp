///==================
///------------------
/// vc contract class
///------------------
///==================
#include <bcwasm/bcwasm.hpp>
#include <bcwasm/statevc.hpp>
#include <stdio.h>
#include <stdlib.h>
#include <string>
#include <vector>
using namespace std;

/// const value define
#define INPUT_DELIM "#"
#define RES_DELIM "#"
#define VC_PREFIX "VC_"
#define PROOR_BUF_SIZE 0x400
#define RES_BUF_SIZE 0x100
#define MIN_COST_VALUE 0x10
#define CT_ADDR_LEN 42

/// vc key name
#define KEY_PKEY "VC_PKEY"
#define KEY_VKEY "VC_VKEY"
#define KEY_RES "VC_RES"
#define KEY_PROOF "VC_PROOF"
#define COST_VALUE "VC_CV"

/// vc key value
extern char PKEY_VALUE[];
extern char VKEY_VALUE[];
extern bool GenGadget(std::vector<std::string> &);

/// VCC class
namespace vc {

class VCC : public bcwasm::Contract {
public:
  // Event macro define
  BCWASM_EVENT(start_compute_event, uint64_t, const char *)
  BCWASM_EVENT(set_result_event, uint64_t, const char *)

  /*********************************
   * init
   * save keypair to DB
   ********************************/
  void init() {
    // Write DB
    string strpkey = PKEY_VALUE;
    string strvkey = VKEY_VALUE;
    bcwasm::setState(KEY_PKEY, strpkey);
    bcwasm::setState(KEY_VKEY, strvkey);
    bcwasm::println("init vccMain");
  }

  /*********************************
   * compute()
   * Save input data and return task id
   * input value format:123$45$789...
   ********************************/
  void compute(const char *pinput) {
    // gen task id
    string TaskId = gen_task_id();

    // save user input value
    const string strKey = VC_PREFIX + TaskId;
    string strInput = pinput;
    bcwasm::setState(strKey, strInput);

    // save cost value
    bool bRet = save_cost_value(TaskId.c_str());
    if (!bRet) {
      bcwasm::println("call save_cost_value fail...");
      BCWASM_EMIT_EVENT(start_compute_event, 0,
                        "Insufficient value for compute.");
      // return;
    }

    // // Fire event
    BCWASM_EMIT_EVENT(start_compute_event, 1, TaskId.c_str());

    // //for test
    // string proof, result;
    // string strRet = real_compute(TaskId.c_str());
    // size_t pos = strRet.find_first_of(RES_DELIM, 0);
    // if (pos != string::npos) {
    // 	proof = strRet.substr(0, pos);
    // 	result= strRet.substr(pos+strlen(RES_DELIM));
    // 	bcwasm::println("proof is ", proof.c_str());
    // 	bcwasm::println("result is ", result.c_str());
    // }

    // set_result(TaskId.c_str(), strRet.c_str());

    // int64_t resVal = get_result(TaskId.c_str());
    // bcwasm::println("get result value is ", resVal);
    // //end test
  }


  /*********************************
   * real_compute
   * real compute,gen proof and result
   * Part of the code is generated automatically
   ********************************/
  const char *real_compute(const char *task_id) {
    // get input data
    vector<string> vectInput;
    const string delim = INPUT_DELIM;
    string strKey = VC_PREFIX + string(task_id);
    std::string inputs;
    bcwasm::getState(strKey, inputs);
    bcwasm::println("get user input:", inputs);
    split(inputs, delim, vectInput);

    // get pkey
    string strPKey;
    bcwasm::getState(KEY_PKEY, strPKey);

    // 1) init gadget env
    vc_InitGadgetEnv();

    // 2) create gadgets
    GenGadget(vectInput);

    // 3) gen witness
    vc_GenerateWitness();

    // 4) gen proof & result
    char Proof[PROOR_BUF_SIZE + 1] = {0};
    char Result[RES_BUF_SIZE + 1] = {0};
    int Ret = vc_GenerateProofAndResult(strPKey.c_str(), strPKey.size(), Proof,
                                        PROOR_BUF_SIZE, Result, RES_BUF_SIZE);
    if (!Ret)
      bcwasm::println("Gen proof & result fail...");
    else
      bcwasm::println("Gen proof & result success...");

    // 5) uninit gadget env
    vc_UninitGadgetEnv();
    // -----------------------------------------

    // retrue proof & result
    // str format:proof#result
    static string res = Proof + string(RES_DELIM) + string(Result);
    const char *s = res.c_str();
    bcwasm::println("Gen proof success res:", s);
    return res.c_str();
  }

  /*********************************
   * set_result
   * call verify to check proof & result
   * true:save proof & result to DB
   ********************************/
  void set_result(const char *task_id, const char *result) {

    string _result(result);
    size_t pos = _result.find_first_of(RES_DELIM, 0);
    string pProof, pResult;

    if (pos != string::npos) {
      pProof = _result.substr(0, pos);
      pResult = _result.substr(pos + strlen(RES_DELIM));
      bcwasm::println("proof is ", pProof.c_str());
      bcwasm::println("result is ", pResult.c_str());
    }
    // get input data
    string strKey = VC_PREFIX + string(task_id);
    std::string inputs;
    bcwasm::getState(strKey, inputs);
    bcwasm::println("input is ", inputs);

    // get vkey
    string strVKey;
    bcwasm::getState(KEY_VKEY, strVKey);

    // call vc_Verify
    int Ret = vc_Verify(strVKey.c_str(), strVKey.size(), pProof.c_str(),
                        pProof.size(), inputs.c_str(), inputs.size(),
                        pResult.c_str(), pResult.size());
    if (!Ret) {
      bcwasm::println("verify fail.");
      BCWASM_EMIT_EVENT(set_result_event, 0, "verify fail.");
      return;
    } else {
      bcwasm::println("verify pass.");
      BCWASM_EMIT_EVENT(set_result_event, 1, "verify pass.");
    }

    // save proof & resultm
    string strResKey = KEY_RES + string(task_id);
    string strProof  = KEY_PROOF + string(task_id);
    bcwasm::println("set result taskid :", strResKey.c_str());
    bcwasm::setState(strResKey, pResult);
    bcwasm::setState(strProof, pProof);

    // transfer to calculation of party
    bcwasm::u256 cost_val = get_cost_value(task_id);

    bcwasm::println("get_cost_value to:", cost_val);
    string from = bcwasm::caller().toString();
    char *dest_addr = new char[CT_ADDR_LEN + 1];
    memset(dest_addr, 0, CT_ADDR_LEN + 1);
    strncpy(dest_addr, from.c_str(), CT_ADDR_LEN);
    bcwasm::println("transfer to:", string(dest_addr), " value:", cost_val);
    bcwasm::Address _addr(dest_addr, true);
    int ret = bcwasm::callTransfer(_addr, cost_val);
    delete dest_addr;

    // save result success
    bcwasm::println("set_result success.", ret);
    BCWASM_EMIT_EVENT(set_result_event, 1, "set result success.");
  }

  /*********************************
   * get_result
   * return the calc value to user
   **********************************/
  int64_t get_result(const char *task_id) {
    // get result data
    string strKey = KEY_RES + string(task_id);
    std::string strRes;
    bcwasm::println("get result taskid :", strKey.c_str());
    bcwasm::getState(strKey, strRes);
    bcwasm::println("get result string: ", strRes);
    return strtoll(strRes.c_str(), 0, 16);
  }


private:
  /// gen task id
  string gen_task_id() {
    uint64_t Nonce = getCallerNonce();
    bcwasm::h160 caller_h = bcwasm::caller();
    std::string caller_h_str = caller_h.toString();
    string strTaskId = caller_h_str + std::to_string(Nonce);
    bcwasm::h256 sha3_h = bcwasm::sha3((bcwasm::byte *)strTaskId.c_str(), strTaskId.size());
    strTaskId = sha3_h.toString();
    bcwasm::println("gen vc task id : ", strTaskId.c_str());
    return strTaskId;
  }

  /// save cost value
 bool save_cost_value(const char *task_id) {
      bcwasm::u256 cost_val = bcwasm::callValue();
			bcwasm::println("cost value:", cost_val);
			if (cost_val <= 0) {
				bcwasm::println("cost value mush greater than zero.");
				//return false;
			}
      string strKey = COST_VALUE + string(task_id);
      string strVal = cost_val.convert_to<std::string>();
      bcwasm::println("save_cost_value: ", strKey.c_str(), " - ", strVal.c_str());
      bcwasm::setState(strKey, cost_val);
			return true;
  }

  /// get cost value
  bcwasm::u256  get_cost_value(const char *task_id) {
      string strKey = COST_VALUE + string(task_id);
      bcwasm::u256 value_u;
      bcwasm::getState(strKey, value_u);
      bcwasm::println("get_cost_value: ", value_u.convert_to<std::string>());
      return value_u;
  }

  /// get pkey data
  string get_pkey_data() {
    std::string pkey;
    bcwasm::getState(KEY_PKEY, pkey);
    return pkey;
  }

  /// get vkey data
  string get_vkey_data() {
    std::string vkey;
    bcwasm::getState(KEY_VKEY, vkey);
    return vkey;
  }

  ///	split string by delim
  void split(const string &str, const string &delim, vector<string> &vectRet) {
    size_t nLast = 0;
    size_t nIndex = str.find_first_of(delim, nLast);
    while (nIndex != string::npos) {
      vectRet.push_back(str.substr(nLast, nIndex - nLast));
      nLast = nIndex + delim.size();
      nIndex = str.find_first_of(delim, nLast);
    }
    if ((nIndex - nLast) > 0) {
      vectRet.push_back(str.substr(nLast, nIndex - nLast));
    }
  }
};
} // namespace vc

BCWASM_ABI(vc::VCC, compute)
BCWASM_ABI(vc::VCC, real_compute)
BCWASM_ABI(vc::VCC, set_result)
BCWASM_ABI(vc::VCC, get_result)

//bcwasm autogen begin
extern "C" { 
void compute(const char * pinput) {
vc::VCC VCC_bcwasm;
VCC_bcwasm.compute(pinput);
}
const char * real_compute(const char * task_id) {
vc::VCC VCC_bcwasm;
return VCC_bcwasm.real_compute(task_id);
}
void set_result(const char * task_id,const char * result) {
vc::VCC VCC_bcwasm;
VCC_bcwasm.set_result(task_id,result);
}
long long get_result(const char * task_id) {
vc::VCC VCC_bcwasm;
return VCC_bcwasm.get_result(task_id);
}
void init() {
vc::VCC VCC_bcwasm;
VCC_bcwasm.init();
}

}
//bcwasm autogen end
