/* global artifacts */
/* global contract */
/* global web3 */
/* global assert */

const assertFail = require("./helpers/assertFail.js");

const BcAssets = artifacts.require("../contracts/BcAssets.sol");

contract("BcAssets", (accounts) => {

    const now = () => Math.floor(Date.now() / 1000)

    const mapAsset = asset => {
        return {
            owner       : asset[0],
            serial      : asset[1].toNumber(),
            creation    : asset[2].toNumber(),
            description : asset[3],
            ipfs        : asset[4].toNumber(),
            ownerIndex  : asset[5]
        }
    }

    const {
        0: owner,
        1: acc1,
        2: acc2
    } = accounts;

    let bcassets;

    beforeEach(async () => {
        bcassets = await BcAssets.new();
    });

    /// --- mint

    it("Minting creates a new asset and event is generated", async () => {
        
        const result = await bcassets.mint(acc1, "asset1", 9191, { from: owner });

        assert.equal(result.logs.length, 1);
        assert.equal(result.logs[ 0 ].event, "LogTransfer");
        assert.equal(result.logs[ 0 ].args.serial, 0);
        assert.equal(result.logs[ 0 ].args.from, 0);
        assert.equal(result.logs[ 0 ].args.to, acc1);

        const asset = mapAsset(await bcassets.assets(result.logs[ 0 ].args.serial))

        assert.equal(asset.owner, acc1);
        assert.equal(asset.serial, 0);
        assert.isOk(asset.creation-now()<5);
        assert.equal(asset.description, "asset1");
        assert.equal(asset.ipfs, 9191);

    });

    it("Non-owner cannot mint", async () => {
        
        try {
            await bcassets.mint(acc1, "asset1", 9191, { from: acc1 });
        } catch (error) {
            return assertFail(error);
        }

    });

    /// --- transfer

    it("An asset can be tranferred to other account and event is generated", async () => {
        
        let result = await bcassets.mint(acc1, "asset1", 9191, { from: owner });
        const serial = result.logs[ 0 ].args.serial
        
        result = await bcassets.transfer(serial, acc2, { from: acc1 });
        assert.equal(result.logs.length, 1);
        assert.equal(result.logs[ 0 ].event, "LogTransfer");
        assert.equal(result.logs[ 0 ].args.serial.toNumber(), serial.toNumber());
        assert.equal(result.logs[ 0 ].args.from, acc1);
        assert.equal(result.logs[ 0 ].args.to, acc2);

    });

    it("An asset cannot be tranferred by non-owner", async () => {
        
        await bcassets.mint(acc1, "asset1", 9191, { from: owner });
        try {
            result = await bcassets.transfer(serial, acc2, { from: acc2 });
        } catch (error) {
            return assertFail(error);
        }

    });

});