## 使用方法

##### 1，使用go命令下载unichain代码，同时下载本代码

命令  

```
go get github.com/unichainplatform/unichain
go get github.com/unichainplatform/faucet4testnet
```


##### 2，切换到当前目录，编译代码

```
cd path/to/github.com/unichainplatform/faucet4testnet
go build faucetServer.go
```

##### 3，创建程序使用的多个子账户

程序使用多个自账户轮流发送交易。所以在启动这个程序前需要先创建好账户。  
默认使用账户名 walletservice.u1~9 九个账户轮流发送交易  
这个账户名可在启动程序时，通过参数 -pn进行设置  

##### 4，启动程序提供http接口服务

启动程序并传入相应的参数  

```
./faucet4testnet -pn walletservice.u -pk xxxxxxxx -l 5 -h [ipAddr]
```

参数解释：  
-pn 用于创建账户的账户名前缀，实际发送交易时会在最后加上一个1～9的数字  
-pk 发送交易的账户私钥，这里认为pn传入的账户使用同一私钥。  
-l 每个ip创建账户的数量限制。  
-h 指定的rpc节点的IP地址

调试完毕可以做为后台服务启动  

```
nohup ./faucet4testnet -pn walletservice.u -pk xxxxxxxx -l 5 -h 47.115.149.93 &
```

##### 5，调用服务的http接口

默认只能通过本地localhost访问

```
http://localhost:9001/wallet_account_creation?accname=arg1&pubkey=0xXXXXXXXXX&deviceid=deviceid
```

参数：  
accname 新建账户名称  
pubkey 新建账户公钥  
deviceid 设备id字串  

成功则返回交易hash  
```
{"code":200,"msg":"0xXXXXXXXXXXXXXX"}
```

##### 6，代码修改说明  

- 文件 138 ～ 145行，设置rpc节点ip与chainID。 chainID 1为主网，100为测试网。
- 文件 254 行，调整轮流发送交易账户个数
- 文件 306 行，调整本地服务监听端口
