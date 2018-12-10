## 关于wasm合约转账的说明

1.当用户在调用某个具体的wasm子合约时，如果需要使用coin，则需要将coins: account_A账户中的coin
  转账（transferToExec）到该wasm_xxx子合约中，该步操作是通过coins合约实现的，即wasm的外部实现。
  
```
  即：coins： account_A ---->  wasm_xxx: account_A  
```

2.当用户需要把coin从某个wasm子合约账户wasm_xxx中进行提币（withdraw）时，也是通过coins合约完成，通过
  withdraw操作来完成；
  
```
  即：wasm_xxx： account_A ----> coins : account_A  
```

3.在wasm的子合约内部需要进行转账或冻结操作时，wasm合约通过平台提供的import机制会把操作账户的接口暴露给
  合约开发者，但是在进行实际的调用处理前，需要首先确认该笔交易的发起者的地址和转账和提币的账户地址是否一致，
  只有两者一致的情况下，调用才会被执行，否则操作失败。
  
## wasm分级账户体系
Step 1:通过coins合约将balance转账到wasm合约(即wasm平台合约)，如果是平行链则是user.p.xxx.wasm
   mavl-coins-BTY-exec-16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp:14KEKbYtKKQm4wMthSK9J4La4nAiidGozt
   key=mavl-execer-symbol-exec-ExecAddress(wasm):addr
   value=protobuffer encode Account
   
Step 2:当用户需要调用user.wasm.xxx合约时，可以用2种方式将balance转账到自己在user.wasm.xxx合约中的账户名下：
   方法一：直接调用wasm的转账到合约的操作
   进行
   
   方法二：通过call方法，即调用user.wasm.xxx时，可以在真正地调用合约前，将balance转账到自己在user.wasm.xxx合约中的账户名下