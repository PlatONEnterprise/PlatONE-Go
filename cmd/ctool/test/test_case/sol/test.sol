pragma solidity ^0.5.2;

contract Test {
    int32 public age = 0;

    constructor() public{}

    function setNum(int32 num) public {
        age = num;
    }

    function getNum() view public returns (int32){
        return age;
    }
}