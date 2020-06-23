## Qitmeer(PMEER) stratum protocol for Cuckaroom

* For other algorithms: visit [examples](https://github.com/leifjacky/pmeer-gominer-demo/tree/master/examples)

### mining.subscribe

- params: ["agent", null]
- result: [[mining.set_difficulty, mining.notify], "extranonce1", extranonce2 size]

```json
request:
{
	"id": 1,
	"method": "mining.subscribe",
	"params": ["pmeerminer-v1.0.0", null]
}

response:
{
	"id": 1,
	"result": [
    [["mining.set_difficulty","32768"],["mining.notify","aaaabbbbccccdddd"]],
    "66976df1",
    4
  ],
	"error": null
}
```



### mining.authorize

- params: ["username", "password"]
- result: true

```json
{
	"id": 2,
	"method": "mining.authorize",
	"params": ["TmdGj68KtaKDbCG5yTMmrB9mFKtzAQpBQqq.worker1", "x"]
}

{"id":2,"result":true,"error":null}
```



### mining.set_difficulty

- params: [difficulty]

```json
{
	"id": null,
	"method": "mining.set_difficulty",
	"params": [32768]
}
```



### mining.notify

- params: ["jobId", "prev hash", "coinbase1", "coinbase2", "coinbase3", "coinbase4", ["merkle branches"], "version", "nbits", "ntime", fresh job]

```json
{
	"id": null,
	"method": "mining.notify",
	"params": ["40205083","bcaa62f248d9f8d540332b82cddfe4bf3c1ff668e1ae9b5cb33ea712fe766f0f","0100010001","000000000000000000000000000000000000000000000000000000000000000002b40f08","0b2f7575706f6f6c2e636e2f","ffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000",[],"0000000c","021ce400","5ef17efd",true]
}
```



### mining.submit

- params: [ "username", "jobId", "extranonce2", "ntime", "nonce", "proofData" ]
- result: true / false

```json
{
	"id": 102,
	"method": "mining.submit",
	"params": ["work01", "40205083", "7cbbc409", "5ef17efd", "00000014", "1d39e822004a119400099df7039a705b0448cd2d06fcc105074a223707b45d56072c8775071c644f08f875fe09142f520a7a3a860c7723580dcaea5e0dc76b860d8c7bff0ddbc1250eff75540ef6c0b50e375acc0ec91f5d11dd0b371353c47514e6d51d155e173b1520359615a318dc1583894b1696bf0f1736c06918368e37191f59bb19006fd91955cc751ae646241b3bac951b66bb161d2f4cb21d3613b71ef69b061f3e1b4d1f"]
}

{"id":102,"result":true,"error":null}    // accepted share response
{"id":102,"result":false,"error":[21,"low difficulty",null]}  // rejected share response
```

> proofData = edgeBits + join(cuckoo nonces in little endian)



```json
In this example

extranonce1 = 0x66976df1
extranonce2 = 0x7cbbc409
extranonce = extranonce1 + extranonce2 = 0x66976df17cbbc409

coinbase1 = 0x0100010001
coinbase2 = 0x000000000000000000000000000000000000000000000000000000000000000002b40f08
coinbase3 = 0x0b2f7575706f6f6c2e636e2f
coinbase4 = 0xffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000

coinbaseMidstate = blake2bD(coinbase2 + extranonce + coinbase3) = blake2bD(0x000000000000000000000000000000000000000000000000000000000000000002b40f0866976df17cbbc4090b2f7575706f6f6c2e636e2f) = 0x2bdf4755eec75a0426ecde553cec6ff799192787efa9b89cc65da93534518d43
coinbaseHash = blake2bD(coinbase1 + coinbaseMidstate + coinbase4) = blake2bD(0x01000000012bdf4755eec75a0426ecde553cec6ff799192787efa9b89cc65da93534518d43ffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000) = 0x1de4cd1b3cbc861a9eaba720c45822082e046b5076a4fe5e7d926f544c00d9de

stateroot = 0x0000000000000000000000000000000000000000000000000000000000000000 // 64 zeros
txRoot = calcMerkleRoot(coinbaseHash, merkleBranches) // join hash func is Blake2bD
headerWithProof = version + prevHash + txRoot + stateRoot + nbits + ntime + nonce + powType + proofData = 0x0c000000bcaa62f248d9f8d540332b82cddfe4bf3c1ff668e1ae9b5cb33ea712fe766f0f1de4cd1b3cbc861a9eaba720c45822082e046b5076a4fe5e7d926f544c00d9de000000000000000000000000000000000000000000000000000000000000000000e41c02fd7ef15e14000000031d39e822004a119400099df7039a705b0448cd2d06fcc105074a223707b45d56072c8775071c644f08f875fe09142f520a7a3a860c7723580dcaea5e0dc76b860d8c7bff0ddbc1250eff75540ef6c0b50e375acc0ec91f5d11dd0b371353c47514e6d51d155e173b1520359615a318dc1583894b1696bf0f1736c06918368e37191f59bb19006fd91955cc751ae646241b3bac951b66bb161d2f4cb21d3613b71ef69b061f3e1b4d1f // all concat in Little Endian
headerHash = reverseBytes(blake2bD(header)) = 0x04de9687cabff80b0339e29164582462fe41eff873efc3b45dcc23e0c7b694bb  // Big Endian

graphWeight: 1856, headerHashDiff: 97574
jobDiff: 32768

headerHashDiff > jobDiff, valid share
```

>graph weight definition: https://github.com/Qitmeer/qitmeer/blob/9433545f617ef69c6b68a45ba63c0ff7c3345429/core/types/pow/cuckaroom.go#L104
>
>cuckoo diff definition: https://github.com/Qitmeer/qitmeer/blob/9433545f617ef69c6b68a45ba63c0ff7c3345429/core/types/pow/diff.go#L194


