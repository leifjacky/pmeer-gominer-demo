## Qitmeer(PMEER) stratum protocol for Cuckoo

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
    [["mining.set_difficulty","8192"],["mining.notify","aaaabbbbccccdddd"]],
    "25610514",
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
	"params": [8192]
}
```



### mining.notify

- params: ["jobId", "prev hash", "coinbase1", "coinbase2", "coinbase3", "coinbase4", ["merkle branches"], "version", "nbits", "ntime", fresh job]

```json
{
	"id": null,
	"method": "mining.notify",
	"params": ["6676749","f6f06dbdf529f03f148c0f896a8e7dabedcfcd0f89eef614bb32dd178e090000","0100000001","000000000000000000000000000000000000000000000000000000000000000003f41e0708","0b2f7575706f6f6c2e636e2f","ffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000",[],"0000000b","0403fdf1","5edf8fe2",true]
}
```



### mining.submit

- params: [ "username", "jobId", "extranonce2", "ntime", "nonce", "proofData" ]
- result: true / false

```json
{
	"id": 102,
	"method": "mining.submit",
	"params": ["TmdGj68KtaKDbCG5yTMmrB9mFKtzAQpBQqq.worker1","00000000","5edf8fe2","008d09c3","1812f60400702006004cd80600c2600d006f142900fc6d2d0065d72f005b373200978b380017293900dd75420026554300959b45009e405000de3453008c425400f75c5e0042a45e00e4a6610063b1690086ce6b004d096f00efbd710060137e00cc1d8500416e8e00f4a99300049a9500f373a700e334aa00a8ccb100b20fbd00ccd5c7004cfcc9009fcece0083f2d900b783da00587adb003809ee00c437ee00ca6df10073d2f500"]
}

{"id":102,"result":true,"error":null}    // accepted share response
{"id":102,"result":false,"error":[21,"low difficulty",null]}  // rejected share response
```

> proofData = edgeBits + join(cuckoo nonces in little endian)



```json
In this example

extranonce1 = 0x25610514
extranonce2 = 0x00000000
extranonce = extranonce1 + extranonce2 = 0x2561051400000000

coinbase1 = 0x0100000001
coinbase2 = 0x000000000000000000000000000000000000000000000000000000000000000003f41e0708
coinbase3 = 0x0b2f7575706f6f6c2e636e2f
coinbase4 = 0xffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000

coinbaseMidstate = blake2bD(coinbase2 + extranonce + coinbase3) = blake2bD(0x000000000000000000000000000000000000000000000000000000000000000003f41e070825610514000000000b2f7575706f6f6c2e636e2f) = 0x80d94f7bae60ace779a7c1f3a03a338ed10984d81ef143f880485b12dd3b000f
coinbaseHash = blake2bD(coinbase1 + coinbaseMidstate + coinbase4) = blake2bD(0x010000000180d94f7bae60ace779a7c1f3a03a338ed10984d81ef143f880485b12dd3b000fffffffffffffffff01007841cb020000001976a9149d18742b210d55e6819fc2454e6d8c0dac4a8f1c88ac0000000000000000) = 0x1302ff45347414238aa3860069a3544a4a4c799ce75d43e9baad7eaa073b9c4f

stateroot = 0x0000000000000000000000000000000000000000000000000000000000000000 // 64 zeros
txRoot = calcMerkleRoot(coinbaseHash, merkleBranches) // join hash func is Blake2bD
headerWithProof = version + prevHash + txRoot + stateRoot + nbits + ntime + nonce + powType + proofData = 0x0b000000f6f06dbdf529f03f148c0f896a8e7dabedcfcd0f89eef614bb32dd178e0900001302ff45347414238aa3860069a3544a4a4c799ce75d43e9baad7eaa073b9c4f0000000000000000000000000000000000000000000000000000000000000000f1fd0304e28fdf5ec3098d00011812f60400702006004cd80600c2600d006f142900fc6d2d0065d72f005b373200978b380017293900dd75420026554300959b45009e405000de3453008c425400f75c5e0042a45e00e4a6610063b1690086ce6b004d096f00efbd710060137e00cc1d8500416e8e00f4a99300049a9500f373a700e334aa00a8ccb100b20fbd00ccd5c7004cfcc9009fcece0083f2d900b783da00587adb003809ee00c437ee00ca6df10073d2f500 // all concat in Little Endian
headerHash = reverseBytes(blake2bD(header)) = 0x00844fc308b3a4499a1c414253047785cfc2d8f0be0dee46629a1b8ee36daf3e  // Big Endian

graphWeight: 48, headerHashDiff: 23775
jobTarget: 8192

headerHashDiff < jobTarget, valid share
```
