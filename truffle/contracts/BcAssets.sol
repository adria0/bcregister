pragma solidity ^0.4.15;

import "./Owned.sol";

contract BcAssets is Owned {

    /// --- events --------------------------------------------------

    event LogTransfer(uint indexed serial, address from, address to);

    /// --- structures ----------------------------------------------

    struct Asset {
        address owner;       // owner of the object
        uint128 serial;      // serial number of the object
        uint64  creation;    // unix creation time
        string  description; // description
        uint256 ipfs;        // the keccak256 hash of the object for ipfs
        uint64  ownerIndex;   
    }
    
    /// --- state mutable variables --------------------------------

    Asset[]                   public assets;
    mapping(address=>Asset[]) public assetOwners;
    mapping(address=>uint)    public nonces;

    /// --- public functions ---------------------------------------
    
    function mint(address _owner, string _description, uint256 _ipfshash)
    onlyOwner public returns (uint){
        Asset[] storage assetOwner = assetOwners[_owner];
        uint128 serial = uint128(assets.length);

        assets.push(Asset({
            owner       : _owner,
            serial      : serial,
            creation    : uint64(now),
            ipfs        : _ipfshash,
            ownerIndex  : uint64(assetOwner.length),
            description : _description
        }));
        
        assetOwners[_owner].push(assets[serial]);

        LogTransfer(serial,0x0,_owner);

        return serial;
    }
    
    function transfer(uint _serial, address _to) public {
        transferInternal(_serial,msg.sender,_to);
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
