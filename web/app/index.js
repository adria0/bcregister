import { default as lightwallet } from 'eth-lightwallet';
import { default as filesaver } from 'file-saver';
import { default as $ } from 'jquery';
import { default as validator } from 'validator';
import { default as toastr } from 'toastr';
import { default as store } from 'store';
import { default as rlp } from 'rlp';
import { default as web3 } from 'web3';

import './/../node_modules/toastr/build/toastr.css';

const ERC20ABI = '[{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"totalSupply","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"remaining","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_from","type":"address"},{"indexed":true,"name":"_to","type":"address"},{"indexed":false,"name":"_value","type":"uint256"}],"name":"Transfer","type":"event"},{"anonymous":false,"inputs":[{"indexed":true,"name":"_owner","type":"address"},{"indexed":true,"name":"_spender","type":"address"},{"indexed":false,"name":"lue","type":"uint256"}],"name":"Approval","type":"event"}]'

let keyStore
let pwDerivedKey
let userInfo = null

function showSection(section) {

	$('#authsection').hide()
	$('#mainsection').hide()
	$('#createsection').hide()
	$('#registeredsection').hide()
	$('#restoresection').hide()

	$(section).show()

	if (store.get('bc-address') !== undefined) {
		let html = store.get('bc-address')
		html += " <a href=# id=unlinkid>desvincular</a>"
		$("#footer").html(html)
		$('#unlinkid').click(() => unlinkid())
	} else {
	}

}

function uiMainSection_show() {

	if (store.get('bc-address') === undefined) {
		showSection('#mainsection')
	} else {
		if (userInfo == null) {
			uiAuthSection_show()
		} else {
			uiRegisteredSection_show()
		}
	}
}

function uiCreateSection_show() {

	showSection('#createsection')

}

function uiRestoreSection_show() {

	showSection('#restoresection')

}

function uiAuthSection_show() {

	const address = store.get('bc-address')
	const link = "<a href='https://etherscan.io/address/"+address+"'>"+address+"</a>"
	$('#authaddress').html(link)

	showSection('#authsection')
	
}

function uiRegisteredSection_show() {

	showSection('#registeredsection')
	console.log("userInfo",userInfo)
	
	let info = userInfo.firstName+" "+userInfo.secondName+"<br>"+userInfo.email

	if (userInfo.emailVerified) {
		info += " (Verificat)"
	} else {
		info += " (No verificat)"
		toastr.error("Consulteu el correu per verificar l'email");
	}

	if ( store.get('bc-backupdone') == false ) {
		$('#backupdiv').show()
	}

	$('#emailreg').html(info+"<br>")

	const address = store.get('bc-address')
	const link = "<a href='https://etherscan.io/address/"+address+"'>"+address+"</a>"

}

function uiAuthSection_auth() {

	const authpasswd = $("#authpasswd").val()
	const ks = lightwallet.keystore.deserialize(store.get('bc-pvk'))
	ks.keyFromPassword(authpasswd, function (err, _pwDerivedKey) {

		if (err) {
			toastr.error('Contrassenya incorrecte');
			return;
		}

		keyStore = ks
		pwDerivedKey = _pwDerivedKey;

		postJsonRpc(
			"bc_auth",
			[]
		).done((data) => {
			if (data.error) {
				toastr.error(data.error.message);
				return
			}
			userInfo = data.data
			console.log(userInfo)
			uiRegisteredSection_show()
		})
	})

}

function unlinkid() {
	const agree = confirm("Esteu segurs?")
	if (agree==true) {
		store.remove('bc-address')
		store.remove('bc-pvk')
		store.remove('bc-backupdone')		
		userInfo = null
		uiMainSection_show()
	}
}

function uiBackupSection_backup() {
	const pvk = store.get('bc-pvk')
	var blob = new Blob([pvk], {type: "text/json;charset=utf-8"});
	filesaver.saveAs(blob, "bc_identity.json");
	store.set('bc-backupdone',true)
	$('#backupdiv').hide()

}

function uiRestoreSection_restore() {

	const passwd = $("#restorepasswd").val()
	const file = $("#restorefile")[0].files[0]

	var reader = new FileReader();
	reader.onload = function(e) {
	  const ks = lightwallet.keystore.deserialize(reader.result)
	  ks.keyFromPassword(passwd, function (err, _pwDerivedKey) {

	  	if (ks.isDerivedKeyCorrect(_pwDerivedKey)) {

		    pwDerivedKey = _pwDerivedKey
	    	keyStore = ks

		    const address = "0x"+ks.getAddresses()[0]
		    store.set('bc-address' , address)
		    store.set('bc-pvk' , ks.serialize())
		    store.set('bc-backupdone' , true)

			toastr.info('Identitat importada');

			postJsonRpc(
				"bc_auth",
				[]
			).done((data) => {
				if (data.error) {
					toastr.error(data.error.message);
					return
				}
				userInfo = data.data
				console.log(userInfo)
				uiRegisteredSection_show()
			})

	  	} else {
			toastr.error('Contrassenya invalida');	  		
	  	}

	    if (err) throw err;

	  })

	}
    reader.readAsText(file);
}


function postJsonRpc(method, params) {

	const address = store.get('bc-address')

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
		"jsonrpc": "2.0",
		"method": method,
		"params": params.concat(address).concat(sighex),
		"id": nonce
	}

	console.log(JSON.stringify(msg))

	return $.ajax({
    	url: 'http://localhost:8080/rpc',
    	type: 'POST',
    	contentType: 'application/json',
    	dataType: 'json',
    	data: JSON.stringify(msg)
	}).fail((err) => {
		toastr.error("Error conectant amb el servidor")
		console.log(err)
	})
}

function uiCreateSection_create() {

	const firstName = $("#firstname").val()
	const secondName = $("#secondname").val()
	const email = $("#email").val()
	const passwd1 = $("#passwd1").val()
	const passwd2 = $("#passwd2").val()
	const mode = $("#mode").val()
	const interest = $("#interest").val()

	if (firstname.length < 2) {
		toastr.error('Nom massa curt');		
	}
	if (secondName.length < 2) {
		toastr.error('Cognom massa curt');		
	}
	if (!validator.isEmail(email)) {
		toastr.error('Email invalid');		
	}

	if (passwd1.length < 4) {
		toastr.error('Contrassenya massa curta');
		return;
	}

	if (passwd1 != passwd2) {
		toastr.error('Les contrassenyes no coincideixen');
		return;
	}

    lightwallet.keystore.createVault(
  		{ password: passwd1 },
  		function (err, ks) {

		  ks.keyFromPassword(passwd1, function (err, _pwDerivedKey) {
		    if (err) throw err;

		    pwDerivedKey = _pwDerivedKey
	    	keyStore = ks

		    ks.generateNewAddress(pwDerivedKey, 1);
		    const address = "0x"+ks.getAddresses()[0]

		    store.set('bc-address' , address)
		    store.set('bc-pvk' , ks.serialize())
		    store.set('bc-backupdone' , false)

			$("#passwd1").val("")
			$("#passwd2").val("")

			/// --- sign proof of posession

			postJsonRpc(
				"bc_register",
				[firstName,secondName,email,mode,interest]
			).done((data) => {
				
				if (data.error) {
					toastr.error(data.error.message);
					return
				}

				userInfo = {
					firstName : firstName,
					secondName : secondName,
					email : email
				}

				uiMainSection_show()
			})
/*						
			ks.passwordProvider = function (callback) {
      			var pw = prompt("Please enter password", "Password");
      		    callback(null, pw);
    		};
*/

	  });
	});
}


window.addEventListener('load', function() {

	$('#uiCreateSection_show').click(() => uiCreateSection_show())
	$('#uiRestoreSection_show').click(() => uiRestoreSection_show())
	$('#uiRestoreSection_restore').click(() => uiRestoreSection_restore())
	$('#uiCreateSection_create').click(() => uiCreateSection_create())
	$('#uiBackupSection_backup').click(() => uiBackupSection_backup())
	$('#uiAuthSection_auth').click(() => uiAuthSection_auth())
	$('#uiRestoreSection_back').click(() => uiMainSection_show());
	$('#uiCreateSection_back').click(() => uiMainSection_show());

	uiMainSection_show() 

	const contract = new web3.eth.Contract(ERC20ABI, "0x9a642d6b3368ddc662CA244bAdf32cDA716005BC")
	
})
