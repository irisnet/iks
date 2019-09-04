# API - IKS

### `1.` `/version` `GET`
- Version of IKS.

### `2.` `/keys` `GET`
- All keys managed by the keyserver.

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

### `4.` `/keys/{name}?bech=acc` `GET`
- Details of one key.

### `5.` `/keys/{name}` `PUT`
- Update the password on a key.
- Request body:
    ```json
    {
        "old_password": "{old_password}",
        "new_password": "{new_password}"
    }
    ```

### `6.` `/keys/{name}` `DELETE`
- Delete a key.
- Request body:
    ```json
    {
        "password": "{password}"
    }
    ```

### `7.` `/tx/sign` `POST`
- Sign a transaction.
- Request body:
    ```json
    {
        "tx": "{tx_json}",
        "name": "{name}",
        "password": "{password}",
        "chain_id": "{chain_id}",
        "account_number": "{account_number}",
        "sequence": "{account_number}",
        
    }
    ```

### `8.` `/tx/bank/send` `POST`
- Generate a send transaction.
- Request body:
    ```json
    {
        "sender": "{sender_address}",
        "reciever": "{reciever_address}",
        "amount": "{amount}",
        "chain_id": "{chain_id}",
        "memo": "{memo/empty}",
        "fees": "{fees}",
        "gas_adjustment": "{gas_adjustment/empty}"
    }
    ```

### `9.` `/tx/broadcast` `POST`
- Broadcast a signed transaction.
- Request body:
    ```
    # Json of a signed transaction
    ```
