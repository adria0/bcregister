pragma solidity ^0.4.15;

import "./AclControlled.sol";

interface IAssetsFallback {
    function isAssetsFallback() public returns (uint);
    function onAssetsFallback(uint _serial) public;
}

/// @title Assets
/// @author Adrià Massanet <adria@codecontext.io>
contract Assets is AclControlled {

    /// --- constants -----------------------------------------------

    uint constant public ACL_ASSETADMIN = 2;

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

        uint256 customAttr1;   // custom attribute, set by  
        uint256 customAttr2;

        uint64  ownerIndex;
        
    }
    
    /// --- state mutable variables --------------------------------

    Asset[]                   public assets;
    mapping(address=>Asset[]) public assetOwners;
    mapping(address=>uint)    public nonces;

    /// --- public functions ---------------------------------------
    
    function mint(address _owner, uint16 _class, bool _transferable, uint64 _caducity, string _description)
    onlyAcl(ACL_OWNER|ACL_ASSETADMIN) public returns (uint){
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
    onlyAcl(ACL_BYPASS) public returns (uint){

        require (
            msg.sender == assets[_serial].owner
            || checkAcl(msg.sender,ACL_OWNER|ACL_ASSETADMIN)
        );
        transferInternal(_serial,assets[_serial].owner,0xdead);

    }
    
    function transfer(uint _serial, address _to, bool _notify)
    onlyAcl(ACL_BYPASS) public {

        transferInternal(_serial,msg.sender,_to);

        if (_notify && IAssetsFallback(_to).isAssetsFallback()
            == 0x5ff3af2a56b585d12453d7073276716a9a19e813ee0e5ef2e8996ca10985f0f4) {
            IAssetsFallback(_to).onAssetsFallback(_serial);
        }
    }
    
    function setCustomAttr1(uint _serial, uint256 _value)
    onlyAcl(ACL_BYPASS) public {

        require(assets[_serial].owner == msg.sender);
        assets[_serial].customAttr1 = _value;

    }

    function setCustomAttr2(uint _serial, uint256 _value)
    onlyAcl(ACL_OWNER|ACL_ASSETADMIN) public {

        require(assets[_serial].owner != 0x0);
        assets[_serial].customAttr2 = _value;

    }
    
    function transferOffchain(uint _serial, address _to, uint64 _nonce, uint8 _v, bytes32 _r, bytes32 _s)
    onlyAcl(ACL_BYPASS) public {
        
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
        require ( assets[_serial].transferable );

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
