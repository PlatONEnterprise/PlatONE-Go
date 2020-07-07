solidity ^0.4.21;

contract InfoContract {
    
   string fName;
   uint age;
   
    event Instructor(
       string name,
       uint age
    );
   
   function setInfo(string _fName, uint _age) public {
       fName = _fName;
       age = _age;
       emit Instructor(_fName, _age);
   }
   
   function getInfo() public constant returns (string, uint) {
       return (fName, age);
   }   
}
