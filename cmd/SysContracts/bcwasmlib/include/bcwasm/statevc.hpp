//
// Created by zhou.yang on 2018/10/19.
//

#pragma once

#include "fixedhash.hpp"


extern "C" {
    //for vcc

    void vc_InitGadgetEnv();

    void vc_UninitGadgetEnv();

    void vc_CreatePBVar(int64_t varAddr);

    uint8_t vc_CreateGadget(int64_t input0, int64_t input1, int64_t input2, int64_t res, int32_t Type);

    void vc_SetVar(int64_t varAddr, int64_t Val, unsigned char is_unsigned);

    void vc_SetRetIndex(int64_t RetAddr);

    void vc_GenerateWitness();

    uint8_t vc_GenerateProofAndResult(const char *pPKEY, int32_t pkSize, char *pProof, int32_t prSize, char *pResult, int32_t resSize);

    uint8_t vc_Verify(const char *pVKEY, int32_t pkSize, const char *pPoorf, int32_t prSize, const char *pInput, int32_t inSize, const char *pOutput, int32_t outSize); 
}

