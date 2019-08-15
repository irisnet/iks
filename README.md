# IKS - Key Server for IRIS Hub

This is a basic key server for IRIS Hub named `iks`. It contains the following routes:

```
GET     /version
GET     /keys
POST    /keys
GET     /keys/{name}?bech=acc
PUT     /keys/{name}
DELETE  /keys/{name}
POST    /tx/sign
POST    /tx/bank/send
POST    /tx/broadcast
```

First, build and start the server:

***For Testnet, please update [`NetworkType = "testnet"`](./cmd/serve.go#L28) manually***

```bash
> make install
> iks config
> iks serve
```

Then you can use the included CLI to create keys, use the mnemonics to create them in `iriscli` as well:

```bash
# Create a new key with generated mnemonic
> iks keys post jack foobarbaz | jq

# Create another key
> iks keys post jill foobarbaz | jq

# Save the mnemonic from the above command and add it to iriscli
> iriscli keys add jack --recover

# Next create a single node testnet
> iris init --moniker JackNode --chain-id iksnet
> iris add-genesis-account $(iks keys show jack | jq -r .address) 10000000000iris
> iris add-genesis-account $(iks keys show jill | jq -r .address) 100000000iris
> iris gentx --name jack
> iris collect-gentxs
> iris start
```

In another window, generate the transaction to sign, sign it and broadcast:
```bash
> mkdir -p test_data
> iks tx bank send $(iks keys show jack | jq -r .address) $(iks keys show jill | jq -r .address) 10000.58iris iksnet "jack to jill" 0.3iris > test_data/unsigned.json
> iks tx sign jack foobarbaz iksnet 0 1 test_data/unsigned.json > test_data/signed.json
> iks tx broadcast test_data/signed.json
{"height":"0","txhash":"84CEF8B7FD04DA6FE9C22A6077D8286FA7775CAA0BB06D1D875AE9527A3D15CB"}
> iriscli tendermint tx 84CEF8B7FD04DA6FE9C22A6077D8286FA7775CAA0BB06D1D875AE9527A3D15CB --trust-node
```
