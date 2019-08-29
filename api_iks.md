# API - IKS

### `1.` `/version` `GET`
- Version of IKS.
- Request URL:
    ```
    http://localhost:3000/version
    ```

### `2.` `/keys` `GET`
- All keys managed by the keyserver.
- Request URL:
    ```
    http://localhost:3000/keys
    ```

### `3.` `/keys` `POST`
- Add a key.
- Request body:
    ```json
    {
        "name": "{name}",
        "password": "{password}",
        "mnemonic": "{mnemonic/empty}"
    }
    ```
- Request URL:
    ```
    http://localhost:3000/keys
    ```

### `4.` `/keys/{name}?bech=acc` `GET`
- Details of one key.
- Request URL:
    ```
    http://localhost:3000/keys/{name}?bech=acc
    ```

### `5.` `/keys/{name}` `PUT`
- Update the password on a key.
- Request body:
    ```json
    {
        "old_password": "{old_password}",
        "new_password": "{new_password}"
    }
    ```
- Request URL:
    ```
    http://localhost:3000/keys/{name}
    ```

### `6.` `/keys/{name}` `DELETE`
- Delete a key.
- Request body:
    ```json
    {
        "password": "{password}"
    }
    ```
- Request URL:
    ```
    http://localhost:3000/keys/{name}
    ```

### `7.` `/tx/sign` `POST`
- Sign a transaction.
- Request body:
    ```json
    {
        "tx": "{tx-json}",
        "name": "{name}",
        "passphrase": "{passphrase}",
        "chain_id": "{chain_id}",
        "account_number": "{account_number}",
        "sequence": "{account_number}",
        
    }
    ```
- Request URL:
    ```
    http://localhost:3000/tx/sign
    ```

### `8.` `/tx/bank/send` `POST`
- Generate a send transaction.
- Request body:
    ```json
    {
        "sender": "{sender-address}",
        "reciever": "{reciever-address}",
        "amount": "{amount}",
        "chain-id": "{chain-id}",
        "memo": "{memo/empty}",
        "fees": "{fees}",
        "gas_adjustment": "{gas_adjustment/empty}"
    }
    ```
- Request URL:
    ```
    http://localhost:3000/tx/bank/send
    ```

### `9.` `/tx/broadcast` `POST`
- Broadcast a signed transaction.
- Request body:
    ```
    # Json of a signed transaction
    ```
- Request URL:
    ```
    http://localhost:3000/tx/broadcast
    ```