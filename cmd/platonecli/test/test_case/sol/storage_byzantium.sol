pragma solidity >=0.4.22 <0.7.0;

/**
 * @title Storage
 * @dev Store & retreive value in a variable
 */
contract Storage {

    int32 number;

    /**
     * @dev Store value in variable
     * @param num value to store
     */
    function store(int32 num) public {
        number = num;
    }

    /**
     * @dev Return value 
     * @return value of 'number'
     */
    function retreive() public view returns (int32){
        return number;
    }
}