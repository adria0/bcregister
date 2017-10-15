import { default as lightwallet } from 'eth-lightwallet';
import { default as filesaver } from 'file-saver';
import { default as $ } from 'jquery';
import { default as validator } from 'validator';
import { default as toastr } from 'toastr';
import { default as store } from 'store';
import { default as rlp } from 'rlp';

import './/../node_modules/toastr/build/toastr.css';

let keyStore

function showMainSection() {

	if (store.get('bc-address') === undefined) {
		showRegisterSection()
	} else {
		showRegisteredSection()
	}
}

function showRegisterSection() {

	$('#unregisteredsection').show()
	$('#registeredsection').hide()

}

function showRegisteredSection() {

	$('#unregisteredsection').hide()
	$('#registeredsection').show()
	
	const address = store.get('bc-address')
	const pvk = store.get('bc-pvk')
	const email = store.get('bc-email')

	const link = "<a href='https://etherscan.io/address/"+address+"'>"+address+"</a>"
	$('#addressreg').html(link)
	$('#emailreg').text(email)

}

function unlink() {

	const agree = confirm("Esteu segurs?")
	if (agree==true) {
		store.remove('bc-address')
		store.remove('bc-pvk')
		store.remove('bc-email')
		showMainSection()
	}
}

function backup() {
	const pvk = store.get('bc-pvk')
	var blob = new Blob([pvk], {type: "text/json;charset=utf-8"});
	filesaver.saveAs(blob, "bc_identity.json");
}

function signJsonRpc(ks, address, pwDerivedKey, method, params) {

	const nonce = (+ new Date()).toString('16') 			
	const paramsWithNonce = params.concat(nonce)
	const encoded = rlp.encode(paramsWithNonce);

	const sig = lightwallet.signing.signMsg(
		ks, pwDerivedKey, encoded, address
	)
	const paramsWithSignature = params.concat([
		Buffer.from(sig.r).toString('hex'),
		Buffer.from(sig.s).toString('hex'),
		Buffer.from(new Uint8Array([sig.v])).toString('hex')			
	])

	return {
		"jsonrpc": "2.0",
		"method": method,
		"params": paramsWithSignature,
		"id": nonce
	}
}

function register() {

	const firstName = $("#firstname").val()
	const secondname = $("#secondname").val()
	const email = $("#email").val()
	const passwd1 = $("#passwd1").val()
	const passwd2 = $("#passwd2").val()

	if (firstname.length < 2
		|| secondname.length < 2 
		|| !validator.isEmail(email)) {
		toastr.error('Dades invalides');
		return;
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
  	{ password: passwd1 }
    , function (err, ks) {

    	  keyStore = ks
		  ks.keyFromPassword(passwd1, function (err, pwDerivedKey) {
		    if (err) throw err;

		    ks.generateNewAddress(pwDerivedKey, 1);
		    const address = ks.getAddresses()[0]

		    store.set('bc-address' , address)
		    store.set('bc-pvk' , ks.serialize())
		    store.set('bc-email' , email)

			$("#passwd1").val("")
			$("#passwd2").val("")

			/// --- sign proof of posession

			const msg = signJsonRpc(
				ks, address, pwDerivedKey,
				"bc_register",
				[firstName,secondname,email]
			)

			console.log(msg)

			$.ajax({
            	url: '/test/PersonSubmit',
            	type: 'post',
            	dataType: 'json',
            	data: jsonRpcIn
        	}).success((data) => {
				alert("Success")
			}).error(() => {
				alert("Failed")
			})
			
			/// POST /api/v1/register/:<address>
			/// data sent is:
			/// { firstName : firstName, secondName : secondName, 
			///   email: email, pof :{ r,s,v } }
			
			ks.passwordProvider = function (callback) {
      			var pw = prompt("Please enter password", "Password");
      		    callback(null, pw);
    		};

			showRegisteredSection()

	  });
	});
}


window.addEventListener('load', function() {

	$('#register').click(() => register())
	$('#backup').click(() => backup())
	$('#unlink').click(() => unlink())

	showMainSection() 
})

