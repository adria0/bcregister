pragma solidity ^0.4.15;

import "./Owned.sol";

contract CourseToken is Owned {

    struct Token {
        address owner;
        uint128 serial;
        string  class;
        uint64  caducity;
        uint64  ownerIndex;
    }
    
    Token[] public tokens;
    mapping(address=>Token[]) public tokenOwners;
    mapping(address=>uint)    public nonces;
    
    function mint(address _owner,string _class, uint64 _caducity)
    onlyOwner public {
        require ( _caducity == 0 || _caducity > uint64(now));

        Token[] storage tokenOwner = tokenOwners[_owner];
        uint128 serial = uint128(tokens.length);

        tokens.push(Token({
            owner      : _owner,
            serial     : serial,
            class      : _class,
            caducity   : _caducity,
            ownerIndex : uint64(tokenOwner.length)
        }));
        
        tokenOwners[_owner].push(tokens[serial]);
    }
    
    function transfer(uint _serial, address _to) public {
        transferInternal(_serial,msg.sender,_to);
    }
    
    function transfer(uint _serial, address _to, uint64 _nonce, uint8 _v, bytes32 _r, bytes32 _s) public {
        
        bytes32 hash = keccak256("bc_coursetoken_transfer",address(this),_serial,_to, _nonce);
        address from = ecrecover(hash,_v,_r,_s);
        
        require(from != 0x0);
        require(nonces[from] < _nonce);
        
        nonces[from] = _nonce;
        
        transferInternal(_serial, from, _to);
    }
    
    function transferInternal(uint _serial, address _from, address _to) internal {
        require ( _from != _to );
        require ( tokens[_serial].owner == _from );

        Token[] storage tokenOwner = tokenOwners[_from];
        
        // safety check if indexes are ok
        assert(tokenOwner[tokens[_serial].ownerIndex].serial == _serial);
        
        /// update the token owner
        tokenOwner[tokens[_serial].ownerIndex] = tokenOwner[tokenOwner.length-1];
        tokenOwner.length--;
        tokenOwners[_to].push(tokens[_serial]);

        /// update the token
        tokens[_serial].owner = _to;
    }
    
    function tokenCount() public constant returns (uint) {
        return tokens.length;
    }
    
    function addressTokens(uint _fromIndex, uint _toIndex) public constant returns (Token[] memory) {

        require(_fromIndex <= _toIndex);

        if (_fromIndex >= tokens.length) _fromIndex=tokens.length;
        if (_toIndex >= tokens.length) _toIndex=tokens.length;
        
        Token[] memory ret = new Token[](_toIndex-_fromIndex);
        for (uint i=_fromIndex;i<_toIndex;i++) {
            ret[i-_fromIndex] = tokens[i];
        }

        return ret;

    }

}
