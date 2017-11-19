import rlp from 'rlp';
import web3 from 'web3';
import store from 'store'
import lightwallet from 'eth-lightwallet';
import config from './config';
import $ from 'jquery';

export default class bc {

    constructor( ) {
    	this.jwt = null
    	this.web3 = null
    	this.keystore = null
    	this.userInfo = null
    	console.log(this.initialized)
    }

    get conf() {
		return config;	
	}

	reset() {
		config.reset()
    	this.jwt = null
    	this.web3 = null
    	this.keystore = null
    	this.userInfo = null
	}

	_updateJwt(_jwt) {
		this.jwt = _jwt
		web3 = new Web3(new Web3.providers.HttpProvider("https://localhost:8443/web3",0,"jwt",this.jwt));
	}

	async _postSignedJsonRpc(method, params, keyStore, pwDerivedKey, address) {

		const nonce = + new Date()			

		const paramsWithNonce = params.concat(method).concat(nonce)
		const encoded = rlp.encode(paramsWithNonce);

		const sig = lightwallet.signing.signMsg(
			keyStore, pwDerivedKey, encoded, address
		)
		const sighex = 
			Buffer.from(sig.r).toString('hex')+
			Buffer.from(sig.s).toString('hex')+
			Buffer.from(new Uint8Array([sig.v])).toString('hex')

		const msg = {
			"jsonrpc" : "2.0",
			"method"  : method,
			"params"  : params.concat(address),
			"id"      : nonce
		}

		return await $.ajax({
	    	url: '/rpc',
	    	type: 'POST',
	    	contentType: 'application/json',
	    	dataType: 'json',
	    	data: JSON.stringify(msg),
	    	beforeSend: function(request) {
	    		request.setRequestHeader("Authorization","Signature "+sighex);
	  		}
		}) 
	}

	_createVault (v) {
		return new Promise( (resolve, reject) => {
			lightwallet.keystore.createVault(v, (err, ks) => {
				if (err) reject(err);
				else resolve(ks);
			})
		});
	}

	_keyFromPassword (ks,password) {
		return new Promise( (resolve, reject) => {
			ks.keyFromPassword(password, (err, ks) => {
				if (err) reject(err);
				else resolve(ks);
			})
		});
	}

	async createWallet(passwd, firstName,secondName,email,mode,interest,captcha) {


		const ks = await this._createVault({ password: passwd })
		const pwDerivedKey = await this._keyFromPassword(ks,passwd)

        ks.generateNewAddress(pwDerivedKey, 1);

        const address = "0x"+ks.getAddresses()[0];

        let resp = await this._postSignedJsonRpc(
			"bc_register",
			[firstName,secondName,email,mode,interest,captcha],
			ks,pwDerivedKey,address
		)

		if (resp.error) {
			throw resp.error.message;
		}

        config.address = address;
        config.pvk = ks.serialize();
        config.backedup = false;

        this.keystore = ks;
		this.userInfo = {
			firstName : firstName,
			secondName : secondName,
			email : email
		}
	}

	async login(passwd) {

		const ks = lightwallet.keystore.deserialize(config.pvk)
		const pwDerivedKey = await this._keyFromPassword(ks,passwd)
        let resp = await this._postSignedJsonRpc(
			"bc_auth",
			[],
			ks,pwDerivedKey,config.address
		)
		console.log("resp",resp)

		if (resp.error) {
			throw resp.error.message;
		}

        this.keystore = ks;
		this.userInfo = resp.data.member
		this._updateJwt(resp.data.jwt)

	}

	async recoverAndLogin(backup,passwd) {

		const ks = lightwallet.keystore.deserialize(backup)
		const pwDerivedKey = await this._keyFromPassword(ks,passwd)
		if (!ks.isDerivedKeyCorrect(pwDerivedKey)) {
			throw "Contrassenya incorrecte"
		}
		
		const address = "0x"+ks.getAddresses()[0]

        let resp = await this._postSignedJsonRpc(
			"bc_auth",
			[],
			ks,pwDerivedKey,address
		)

		if (resp.error) {
			throw resp.error.message;
		}

        config.address = address;
        config.pvk = backup;
        config.backedup = true;

        this.keystore = ks;
		this.userInfo = resp.data.member
		this._updateJwt(resp.data.jwt)

	}


}