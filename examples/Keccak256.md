## Qitmeer(PMEER) stratum cpuminer written in golang 

For tests only.

### usage

```bash
go get -v github.com/leifjacky/pmeer-gominer-demo
cd $GOPATH/src/github.com/leifjacky/pmeer-gominer-demo
go run *.go
```



## Qitmeer(PMEER) stratum protocol for Keccak256

- For other algorithms: visit [examples](https://github.com/leifjacky/pmeer-gominer-demo/tree/master/examples)

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
    [["mining.set_difficulty","0.0009765625"],["mining.notify","aaaabbbbccccdddd"]],
    "e2a7258d",
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
	"params": [0.0009765625]
}
// Job target set to 000003FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFC00.
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
	"params": ["26252908","edad24be7f1e2a2dc5ea26885f1baff64dc3f56eea2885714c515482f9070000","0100000001","000000000000000000000000000000000000000000000000000000000000000003210f0708","0b2f7575706f6f6c2e636e2f","ffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000",[],"0000000b","1c1fffff","5eddb44d",true]
}
```



### mining.submit

- params: [ "username", "jobId", "extranonce2", "ntime", "nonce" ]
- result: true / false

```json
{
	"id": 102,
	"method": "mining.submit",
	"params": ["TmdGj68KtaKDbCG5yTMmrB9mFKtzAQpBQqq.worker1","26252908","4d658221","5eddb44d","008caeac"]
}

{"id":102,"result":true,"error":null}    // accepted share response
{"id":102,"result":false,"error":[21,"low difficulty",null]}  // rejected share response
```





```json
In this example

extranonce1 = 0xe2a7258d
extranonce2 = 0x4d658221
extranonce = extranonce1 + extranonce2 = 0xe2a7258d4d658221

coinbase1 = 0x0100000001
coinbase2 = 0x000000000000000000000000000000000000000000000000000000000000000003210f0708
coinbase3 = 0x0b2f7575706f6f6c2e636e2f
coinbase4 = 0xffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000

coinbaseMidstate = blake2bD(coinbase2 + extranonce + coinbase3) = blake2bD(0x000000000000000000000000000000000000000000000000000000000000000003210f0708e2a7258d4d6582210b2f7575706f6f6c2e636e2f) = 0x7c092a91a4f8803803834f2474ddef934c8e2f2a1386dbce4fa727c434881955
coinbaseHash = blake2bD(coinbase1 + coinbaseMidstate + coinbase4) = blake2bD(0x01000000017c092a91a4f8803803834f2474ddef934c8e2f2a1386dbce4fa727c434881955ffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000) = 0x52ea12b949a2386ca18f9bb50824901352a51d1323f084dcb0a93f4dd85c005d

stateroot = 0x0000000000000000000000000000000000000000000000000000000000000000 // 64 zeros
txRoot = calcMerkleRoot(coinbaseHash, merkleBranches) // join hash func is Blake2bD
header = version + prevHash + txRoot + stateRoot + nbits + ntime + nonce + powType = 0x0b000000edad24be7f1e2a2dc5ea26885f1baff64dc3f56eea2885714c515482f907000052ea12b949a2386ca18f9bb50824901352a51d1323f084dcb0a93f4dd85c005d0000000000000000000000000000000000000000000000000000000000000000ffff1f1c4db4dd5eacae8c0006 // all concat in Little Endian
headerHash = reverseBytes(keccak256(header)) = 0x000003276b3b3037ad3deb0da9ff2bc84f56291706f21b00e6500b657d081422  // Big Endian

jobTarget = 0x000003FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFC00

headerHash < jobTarget, valid share
```

