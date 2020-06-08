## Qitmeer(PMEER) stratum protocol for Blake2bD

* For other algorithms: visit [examples](https://github.com/leifjacky/qitmeer-gominer-demo/tree/master/examples)

### mining.subscribe

- params: ["agent", null]
- result: [[mining.set_difficulty, mining.notify], "extranonce1", extranonce2 size]

```json
request:
{
	"id": 1,
	"method": "mining.subscribe",
	"params": ["qitmeerminer-v1.0.0", null]
}

response:
{
	"id": 1,
	"result": [
    [["mining.set_difficulty","0.031250"],["mining.notify","aaaabbbbccccdddd"]],
    "f8dd2051",
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
	"params": [0.0078125]
}
// Job target set to 0000007FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF80.
```

> Note that difficulty 1 in mainnet is now defined as 0x1d00fff, which target is 0x00000000FFFF0000000000000000000000000000000000000000000000000000.
>
> For pool, pow limit is 0x00000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF.



### mining.notify

- params: ["jobId", "prev hash", "coinbase1", "coinbase2", "coinbase3", "coinbase4", ["merkle branches"], "version", "nbits", "ntime", fresh job]

```json
{
	"id": null,
	"method": "mining.notify",
	"params": ["3235889","e85cdb99273f82c0e4091333e509f68d70973a6db365a287d08c30d3b13e4472","0100000001","00000000000000000000000000000000000000000000000000000000000000000323060708","0b2f7575706f6f6c2e636e2f","ffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000",[],"0000000b","1c1fffff","5edca5de",true]
}
```



### mining.submit

- params: [ "username", "jobId", "extranonce2", "ntime", "nonce" ]
- result: true / false

```json
{
	"id": 102,
	"method": "mining.submit",
	"params": ["TmdGj68KtaKDbCG5yTMmrB9mFKtzAQpBQqq.worker1","3235889","4d658221","5edca5de","00102dd6"]
}

{"id":102,"result":true,"error":null}    // accepted share response
{"id":102,"result":false,"error":[21,"low difficulty",null]}  // rejected share response
```





```json
In this example

extranonce1 = 0xf8dd2051
extranonce2 = 0x4d658221
extranonce = extranonce1 + extranonce2 = 0xf8dd20514d658221

coinbase1 = 0x0100000001
coinbase2 = 0x00000000000000000000000000000000000000000000000000000000000000000323060708
coinbase3 = 0x0b2f7575706f6f6c2e636e2f
coinbase4 = 0xffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000

coinbaseMidstate = blake2bD(coinbase2 + extranonce + coinbase3) = blake2bD(0x00000000000000000000000000000000000000000000000000000000000000000323060708f8dd20514d6582210b2f7575706f6f6c2e636e2f) = 0xfe9463841f359b39244904530c7c41817b5e365ffcb4cb3b0a6628559a7f9896
coinbaseHash = blake2bD(coinbase1 + coinbaseMidstate + coinbase4) = blake2bD(0x0100000001fe9463841f359b39244904530c7c41817b5e365ffcb4cb3b0a6628559a7f9896ffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000) = 0xb5b708eeed85b305c84273c7e986ad554abad2096bcac787d96ae880e930a334

stateroot = 0x0000000000000000000000000000000000000000000000000000000000000000 // 64 zeros
txRoot = calcMerkleRoot(coinbaseHash, merkleBranches) // join hash func is Blake2bD
header = version + prevHash + txRoot + stateRoot + nbits + ntime + nonce + powType = 0x0b000000e85cdb99273f82c0e4091333e509f68d70973a6db365a287d08c30d3b13e4472b5b708eeed85b305c84273c7e986ad554abad2096bcac787d96ae880e930a3340000000000000000000000000000000000000000000000000000000000000000ffff1f1cdea5dc5ed62d100000 // all concat in Little Endian
headerHash = reverseBytes(blake2bD(header)) = 0x0000004f858c03aabf8a55f588cc5b33b095c3ba0249c1a3c00e320384cfc958  // Big Endian

jobTarget = 0x0000007FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF80

headerHash < jobTarget, valid share
```

