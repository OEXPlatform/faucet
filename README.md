## 使用方法

##### 1，使用go命令下载oexchain代码，同时下载本代码

命令  

```
go get github.com/oexplatform/oexchain
go get github.com/oexplatform/faucet
```


##### 2，切换到当前目录，编译代码

```
cd path/to/github.com/oexplatform/faucet
go build faucetServer.go
```

##### 3，创建程序使用的多个子账户

程序使用多个自账户轮流发送交易。所以在启动这个程序前需要先创建好账户。  
默认使用账户名 walletservice.u1~9 九个账户轮流发送交易  
这个账户名可在启动程序时，通过参数 -pn进行设置  

##### 4，启动程序提供http接口服务

启动程序并传入相应的参数  

```
./faucet4 -pn walletservice.u -pk xxxxxxxx -l 5
```

参数解释：  
-pn 用于创建账户的账户名前缀，实际发送交易时会在最后加上一个1～9的数字  
-pk 发送交易的账户私钥，这里认为pn传入的账户使用同一私钥。  
-l 每个ip创建账户的数量限制。  

调试完毕可以做为后台服务启动  

```
nohup ./faucet -pn walletservice.u -pk xxxxxxxx -l 5 &
```

##### 5，调用服务的http接口

默认只能通过本地localhost访问

```
http://localhost:9001/wallet_account_creation?accname=arg1&pubkey=0xXXXXXXXXX&deviceid=deviceid&rpchost=47.115.149.93&rpcport=8080&chainid=100
```

参数：  
accname: 新建账户名称  
pubkey: 新建账户公钥  
deviceid: 设备id字串  
rpchost: rpc节点IP  
rpcport: rpc节点Port  
chainid: rpc节点所在链的chainId


成功则返回交易hash  
```
{"code":200,"msg":"0xXXXXXXXXXXXXXX"}
```

