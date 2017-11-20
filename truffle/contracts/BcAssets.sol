pragma solidity ^0.4.15;

import "./Owned.sol";

interface IBcAssetFallback {
    function isBcAssetFallback() returns (uint);
    function onBcAssetFallback(uint _serial);
}

contract BcAssets is Owned {

    /// --- events --------------------------------------------------

    event LogTransfer(uint indexed serial, address from, address to);

    /// --- structures ----------------------------------------------

    struct Asset {
        address owner;         // owner of the object

        uint128 serial;        // serial number of the object
        uint64  class;         // class of object , 1 == membership
        bool    transferable;  // can be transfered

        uint64  caducity;      // unix caducity time        
        string  description;   // description

        uint256 customAttr1;
        uint256 customAttr2;

        uint64  ownerIndex;
        
    }
    
    /// --- state mutable variables --------------------------------

    Asset[]                   public assets;
    mapping(address=>Asset[]) public assetOwners;
    mapping(address=>uint)    public nonces;

    /// --- public functions ---------------------------------------
    
    function mint(address _owner, uint16 _class, bool _transferable, uint64 _caducity, string _description)
    onlyOwner public returns (uint){
        Asset[] storage assetOwner = assetOwners[_owner];
        uint128 serial = uint128(assets.length);

        assets.push(Asset({
            owner         : _owner,
            serial        : serial,
            class         : _class,
            transferable  : _transferable,
            caducity      : _caducity,
            description   : _description,
            customAttr1   : 0,
            customAttr2   : 0,
            ownerIndex    : uint64(assetOwner.length)
        }));
        
        assetOwners[_owner].push(assets[serial]);

        LogTransfer(serial,0x0,_owner);

        return serial;
    }

    function burn(uint _serial) 
    public returns (uint){
        require (
            msg.sender == assets[_serial].owner
            || msg.sender == owner
        );
        transferInternal(_serial,assets[_serial].owner,0xdead);
    }
    
    function transfer(uint _serial, address _to, bool _notify) public {
        transferInternal(_serial,msg.sender,_to);

        if (_notify && IBcAssetFallback(_to).isBcAssetFallback()
            == 0x5ff3af2a56b585d12453d7073276716a9a19e813ee0e5ef2e8996ca10985f0f4) {
            IBcAssetFallback(_to).onBcAssetFallback(_serial);
        }
    }
    
    function setCustomAttr1(uint _serial, uint256 _value) public {
        require(msg.sender == assets[_serial].owner);
        assets[_serial].customAttr1 = _value;
    }

    function setCustomAttr2(uint _serial, uint256 _value) public {
        require(msg.sender == owner);
        assets[_serial].customAttr2 = _value;
    }
    
    function transferOffchain(uint _serial, address _to, uint64 _nonce, uint8 _v, bytes32 _r, bytes32 _s) public {
        
        bytes32 hash = keccak256(msg.sig,address(this),_nonce,_serial,_to, _nonce);
        address from = ecrecover(hash,_v,_r,_s);
        
        require(from != 0x0);
        require(_nonce > nonces[from]);
        
        nonces[from] = _nonce;
        
        transferInternal(_serial, from, _to);
    }
    
    /// --- web3 helpers --------------------------------------------

    function assetCount() public view returns (uint) {
        return assets.length;
    }

    function ownerAssetCount(address _addr) public view returns (uint) {
        return assetOwners[_addr].length;
    }

    /// --- internal functions --------------------------------------

    function transferInternal(uint _serial, address _from, address _to) internal {
        require ( _from != _to );
        require ( assets[_serial].owner == _from );

        Asset[] storage assetOwner = assetOwners[_from];
        
        // safety check if indexes are ok
        assert(assetOwner[assets[_serial].ownerIndex].serial == _serial);
        
        /// update the token owner
        assetOwner[assets[_serial].ownerIndex] = assetOwner[assetOwner.length-1];
        assetOwner.length--;
        assetOwners[_to].push(assets[_serial]);

        /// update the token
        assets[_serial].owner = _to;

        LogTransfer(_serial,_from,_to);

    }
    
}


