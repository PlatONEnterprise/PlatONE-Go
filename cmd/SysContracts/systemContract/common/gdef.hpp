#pragma once

#include <string>

namespace gdef
{
    const char* cnsManagerAddr = "0x0000000000000000000000000000000000000011";
    const char* zeroAddr       = "0x0000000000000000000000000000000000000000";
    const char* invalidAddr    = "0xffffffffffffffffffffffffffffffffffffffff";

    const char* paramManager = "__sys_ParamManager";
    const char* userManager  = "__sys_UserManager";
    const char* userRegister = "__sys_UserRegister";
    const char* roleManager  = "__sys_RoleManager";
    const char* roleRegister = "__sys_RoleRegister";
    const char* nodeManager  = "__sys_NodeManager";
    const char* nodeRegister = "__sys_NodeRegister";

    const char* chainCreator     = "chainCreator";
    const char* chainAdmin       = "chainAdmin";
    const char* nodeAdmin        = "nodeAdmin";
    const char* contractAdmin    = "contractAdmin";
    const char* contractDeployer = "contractDeployer";
    const char* contractCaller   = "contractCaller";


    enum UserStatus
    {
        Enabled = 0,
        Disabled,
        Deleted
    };
}

