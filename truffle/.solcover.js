module.exports = {
    testCommand: 'truffle test --network coverage',
    copyNodeModules: true,
    skipFiles: [
	'truffle/Migrations.sol'
    ]
}
