# 数据上链

## 介绍
数据上链服务，主要是为了在QuarkChain链上存放数据,查询数据   


支持以下网络
-   主网区块浏览器: [here](https://mainnet.quarkchain.io/)
-   测试网区块浏览器: [here](https://devnet.quarkchain.io/)
   
支持以下操作  
- 数据上链  
- 数据查询
    
**注意**: 要求 [Go 1.13+](https://golang.org/dl/)


## 服务启动
### 启动前准备 
-   申请账户:到指定的区块浏览器申请(注意保存私钥)    
-   申请QKC:
    *   主网：交易所购买 或者 联系QKC官方人员
    *   测试网：[here](https://devnet.quarkchain.io/faucet)     
-   查看指定网络的可用host   
    *   打开上述区块浏览器，右上角有部分节点的IP(启动时需"--host=***"来指明),[here]()

### 启动方式
```bash
# Clone the repository
mkdir -p $GOPATH/src/github.com/QuarkChain
cd $GOPATH/src/github.com/QuarkChain
git clone https://github.com/QuarkChain/qkcDataService
cd qkcDataService
go build


# 将私钥进行加密处理
./qkcDataService -type=encrypt --private_key=申请到的私钥 --password=qkc

# 上面命令会生成新的私钥
     ./qkcDataService --private_key=上个命令加密后的字符串 --password=qkc --host="http://IP:38391"
或者 ./qkcDataService --private_key=申请到的私钥  --host="http://IP::38391"

```

## 调用方式

### 数据上链
    注意：
        1.数据格式必须为json格式
        2.目前数目长度最大为87926(byte) 
        
    curl -X POST -H 'content-type: application/json' --data '{"序号":123,"材料名称":"螺栓","材料类别":"土建","出库数量":666777888999000}' http://IP:8080/
    

### 数据查询
    
    curl http://IP:8080/?txHash=调用数据上链时返回的data
    
