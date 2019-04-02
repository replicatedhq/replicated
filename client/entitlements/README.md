Using the entitlements client
===============================

Set environment variables:


```shell
export REPLICATED_APP=<app id or slug>
export REPLICATED_API_TOKEN=<api token>
```

Create a file `entitlements.yaml`

```
cat <<EOF > entitlements.yaml
---
- name: My Field
  key: my_field
  description: Number of Seats
  type: string
  default: "10"
  labels:
    - owner=somePerson
EOF
```

Create a spec from this file

```
$ replicated entitlements define-fields --file=./entitlements.yaml --name=testing 
{
  "id": "A-o3S2mTHrUxxxx2aQwBQoW3skA",
  "spec": "---\n- name: My Field\n  key: num_seats\n  description: Number of Seats\n  type: string\n  default: \"10\"\n  labels:\n    - owner=somePerson\n",
  "name": "testing",
  "createdAt": "Thu Mar 21 2019 18:50:23 GMT+0000 (UTC)"
}
```

Set a value for a customer (get customer ID from Ship install script or console.replicated.com url segment). Use the spec ID from above.

```
$ replicated entitlements set-value --definitions-id=xxx --customer-id=xxxx--key=my_field --value=50 
{
  "id": "oubq5aj_A2KRi0icxxxx1pFJU8drg6s0",
  "key": "my_field",
  "value": "50"
}
```

Preview the end customer signed payload (get customer id and installation id from ship install command)

```
$ replicated entitlements get-customer-release --customer-id=xxxx --installation-id=xxx

{
  "id": "L9x1J6xOMMpKJpMvNCzteHGHSTLAX55j",
  "channelId": "PYe3sroI2QKkLiOC2b0T2Sv9hhT99gWV",
  "channelName": "new channel",

  ... snip ...

  "entitlements": {
    "meta": {
      "last_updated": "0001-01-01T00:00:00Z",
      "customer_id": ""
    },
    "signature": "opGlbWqI0yi8Mf2vSaF9K70mzoKzAeN+xpGZ5mk1ZJXbkQ0kPW9LrqiSG/d2VmOdlBhp5UhiJ6y9o3qYw0pxrvypKgbJSB90BoZpLv7F6ZjK39MDvuL6WOWAv07TUB6cBQXSKXkoDC4U1phHMNEpJSjfqG7kEuej/6QYyKv4HrWpNm7nQU8/3UeKtrSp2Cv5fK7OGm2N3yq63znnFwyl5+ycaVjqnNGhmh56ckKDKyg0rC8lpcm0Bv9vq624pJ468ukigKT6JaOT7gm6Q+MEHCYWbMLm8FPQt/ggRJtMBF+5WxuCsENftdfGVCjCQrnFiGjuB+NzSATTKKBwCxUpj/waK3FANB84Wg3WJ2KSSTPTaZTqKyFJuigTokrWpyWoQdxPO3ekfDqjQtlQS/CrOTjxx0o0Xp7FZGIG10TOVgSTowHNv502ZBvvDmZurRU6QGnLwQHfbtFO87ML7IQm7lVX4ld/KNcPa9hoEGEPzlfE0IDYLH9xruaGLEnIrpbzvO3TdgusETN4L4MfGIcHqBiiz+JxMzCVORBAAssntdIhBUkbvl9p9iRGNwobZ7ojQbSHH/qSbn2CKetvR+takgYV9X+cRDz+Z1xvrNICkXm3kPUgngsbr1JCkZWFd2VbS16sWlPiT2vKKoKbqi6J4B5YmhXMIbJe1r9OPO72x/4=", "values": [
      {
        "key": "num_seats",
        "value": "50"
      }
    ]
  }


```

### Using with ship

Reference individual fields in your ship yaml with

```
...k8s yaml...
- name: NUM_SEATS
  value: '{{repl EntitlementValue "num_seats"}}'
...k8s yaml...

```

Collect full payload so your app can validate the signature at runtime

```
...k8s yaml...
- name: LICENSE_PAYLOAD_B64
  value: {{repl Entitlements | Base64Encode }}

...k8s yaml...

```


### Verifying the Signature

First step is to serialize the entitlements to create a consistently-hashable payload.

Javascript example:

```javascript
import * as stringify from "json-stable-stringify"; // maintains sorting

// from response above
const entitlements = {
    "meta": {
      "last_updated": "0001-01-01T00:00:00Z",
      "customer_id": ""
    },
    "signature": "opGlbWqI0yi8Mf2vSaF9K70mzoKzAeN+xpGZ5mk1ZJXbkQ0kPW9LrqiSG/d2VmOdlBhp5UhiJ6y9o3qYw0pxrvypKgbJSB90BoZpLv7F6ZjK39MDvuL6WOWAv07TUB6cBQXSKXkoDC4U1phHMNEpJSjfqG7kEuej/6QYyKv4HrWpNm7nQU8/3UeKtrSp2Cv5fK7OGm2N3yq63znnFwyl5+ycaVjqnNGhmh56ckKDKyg0rC8lpcm0Bv9vq624pJ468ukigKT6JaOT7gm6Q+MEHCYWbMLm8FPQt/ggRJtMBF+5WxuCsENftdfGVCjCQrnFiGjuB+NzSATTKKBwCxUpj/waK3FANB84Wg3WJ2KSSTPTaZTqKyFJuigTokrWpyWoQdxPO3ekfDqjQtlQS/CrOTjxx0o0Xp7FZGIG10TOVgSTowHNv502ZBvvDmZurRU6QGnLwQHfbtFO87ML7IQm7lVX4ld/KNcPa9hoEGEPzlfE0IDYLH9xruaGLEnIrpbzvO3TdgusETN4L4MfGIcHqBiiz+JxMzCVORBAAssntdIhBUkbvl9p9iRGNwobZ7ojQbSHH/qSbn2CKetvR+takgYV9X+cRDz+Z1xvrNICkXm3kPUgngsbr1JCkZWFd2VbS16sWlPiT2vKKoKbqi6J4B5YmhXMIbJe1r9OPO72x/4=",
    "values": [
      {
        "key": "num_seats",
        "value": "50"
      }
    ]
  }; 

let builder = "";
builder += stringify(entitlements.meta);
builder += stringify(entitlements.values.sort((ev1, ev2) => ev1.key.localeCompare(ev2.key)));
fs.writeFileSync("./serialized.txt", builder);
fs.writeFileSync("./sig.txt", entitlements.signature);
```

Next step is to grab the cert, and put some files in place. You can then verify the signature with openssl 
(or use the bindings in your preferred language).

```shell
# before: get certificate from vendor.replicated.com, store in cert.pem
# before: write contents of builder (above) to serialized.txt
# before: copy signature to sig.txt

openssl x509 -pubkey -noout -in cert.pem > pubkey.pem
openssl dgst -sha256 -verify pubkey.pem -signature sig.txt serialized.txt
```

