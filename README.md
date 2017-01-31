# execute-on-opsman

Tool to execute bosh commands from an OpsManager VM

## Usage

```
execute-on-opsman --target <opsman url> \
                  --username <opsman username> \
                  --password <opsman password> \
                  bosh
                  --ssh-key-path <path to ssh key>
                  [--product-name <product name>]
                  --command <bosh command>
```

## Example

```
execute-on-opsman --target https://pcf.opsman.com \
                  --username example_user \
                  --password example_password \
                  bosh
                  --ssh-key-path ./key.pem
                  --product-name cf
                  --command stop
```
