# IKS - Key Server for IRIS Hub

This is a basic key server for IRIS Hub named `iks`. It contains the following routes:

```
Type    API                      Descriptions
-------------------------------------------------------------------
GET     /version                 Version of IKS
GET     /keys                    All keys managed by the keyserver
POST    /keys                    Add a key
GET     /keys/{name}?bech=acc    Details of one key 
PUT     /keys/{name}             Update the password on a key 
DELETE  /keys/{name}             Delete a key
POST    /tx/sign                 Sign a transaction
POST    /tx/bank/send            Generate a send transaction
POST    /tx/broadcast            Broadcast a signed transaction
```

[Read the API documentation for more information.](api_iks.md)


First, build and start the server.

***For Testnet, please update `NetworkType=testnet` in [makefile](./Makefile#L11) manually***

```bash
# Install iks cli
> make install

# Generate config file (default is $HOME/.iks/config.yaml)
> iks config

# Run the server. Listening on port ':3000'
> iks serve
```

Request your service via api.

```bash
# example:
curl http://localhost:3000/version | jq

# Response:
{
   "version":"v0.1.2",
   "commit":"a913b97cb1210643271f6ca81ef0b1c625d79e41",
   "branch":"master"
}

# jq is a tool for processing JSON inputs, applying the given 
# filter to its JSON text inputs and producing the filter's 
# results as JSON on standard output.
```

You can use the included CLI to manage keys.

#### IKS Command Line Client

```bash
> iks [command] [flags]

# Global Flags:
# Name: --config  
# Type: string  
# Description:  config file (default is $HOME/.iks/config.yaml)
```

Use `iks keys [command] --help` for more information about a command.

<details>
  <summary><b>Brief descriptions of available commands (click to expand)</b></summary>
  
  - config
    ```bash
    # Sets a default config file
    > iks config [flags]
    ```
- help
    ```bash
    # Help provides help for any command in the application.
    > iks help [path to command] [flags]
    ```
- keys
  - delete
    ```bash
    # Delete a key
    > iks keys delete [name] [password] [flags]
    ```
  - get
    ```bash
    # Fetch all keys managed by the keyserver
    > iks keys get [flags]
    ```
  - post 
    ```bash
    # Add a new key to the keyserver, optionally pass a mnemonic to restore the key
    > iks keys post [name] [password] [mnemonic] [flags]
    ```
  - put
    ```bash
    # Update the password on a key
    > iks keys put [name] [oldpass] [newpass] [flags]
    ```
  - show
    ```bash
    # Fetch details for one key
    > iks keys show [name] [flags]
    ```
- server
    ```bash
    # Runs the server
    > iks serve [flags]
    ```
- tx
  - bank send
    ```bash
    # Generate a send transaction
    > iks tx bank send [sender-address] [reciever-address] [amount] [chain-id] [memo] [fees] [gas-adjustment] [flags]
    ```
  - broadcast
    ```bash
    # Broadcast a signed transaction
    > iks tx broadcast [file] [flags]
    ```
  - sign
    ```bash
    # Sign a transaction
    > iks tx sign [name] [password] [chain-id] [account-number] [sequence] [tx-file] [flags]
    ```
- version
    ```bash
    # Prints version information
    > iks version [flags]
    ```
</details></br>

You can use the mnemonics to create keys in `iriscli` as well.

```bash
# Create a new key with generated mnemonic
# iks keys post [name] [password] | jq
> iks keys post jack foobarbaz | jq

# Create another key
# iks keys post [name] [password] | jq
> iks keys post jill foobarbaz | jq

# Save the mnemonic from the above command and add it to iriscli
# iriscli keys add [name] --recover
> iriscli keys add jack --recover
```

Next create a single node testnet. If any question else, you can refer to the following documents:
[1. How to start an IRISnet network locally](https://github.com/irisnet/irishub/blob/master/docs/software/node.md)
[2. IRIS Command Line Client](https://github.com/irisnet/irishub/blob/master/docs/cli-client/README.md)

```bash
# Initialize the configuration files such as genesis.json and config.toml
# iris init --moniker [node-name] --chain-id [chain-id]
> iris init --moniker JackNode --chain-id iksnet

# Use the following command to modify the genesis.json file to assign the initial account balance to the above validator operator account
# iris add-genesis-account [account-address] 10000000000iris
> iris add-genesis-account $(iks keys show jack | jq -r .address) 10000000000iris
> iris add-genesis-account $(iks keys show jill | jq -r .address) 100000000iris

# Create the CreateValidator transaction and sign the transaction by the validator operator account
# iris gentx --name [name]
> iris gentx --name jack

# Configuring validator information
> iris collect-gentxs
> iris start
```

In another window, generate the transaction to sign, sign it and broadcast:
```bash
> mkdir -p test_data

# Generate a send transaction
# iks tx bank send [sender-address] [reciever-address] [amount] [chain-id] [memo] [fees] > [output-file]
> iks tx bank send $(iks keys show jack | jq -r .address) $(iks keys show jill | jq -r .address) 10000.58iris iksnet "jack to jill" 0.3iris > test_data/unsigned.json

# Sign the transaction
# iks tx sign [name] [password] [chain-id] [account-number] [sequence] [tx-file] > [output-file]
> iks tx sign jack foobarbaz iksnet 0 1 test_data/unsigned.json > test_data/signed.json

# Broadcast the signed transaction
# iks tx broadcast [tx-file]
> iks tx broadcast test_data/signed.json
# Response:
{"height":"0","txhash":"84CEF8B7FD04DA6FE9C22A6077D8286FA7775CAA0BB06D1D875AE9527A3D15CB"}

# Search for the transaction which has the same hash in all existing blocks
# iriscli tendermint tx [hash] [flags]
> iriscli tendermint tx 84CEF8B7FD04DA6FE9C22A6077D8286FA7775CAA0BB06D1D875AE9527A3D15CB --trust-node
```


