#!/usr/bin/python
from cmd import Cmd
import os
import sys
import json
import subprocess
import copy
import time
from shutil import copyfile

class Cli(Cmd):
    prompt =  '[group:0]>'
    intro = """Welcom to platonecli!
    Usage:
         <command>[sub command] [command options] [arguments...]
    you can type '?' or 'help' for help"""

    def __init(self):
        Cmd.__init__(self)
    
    def do_one(self,line):
        """create and start a new node
        Usage:
           one  [command options]
           Options::
                    --groupid                   the specified groupid, default 0
                    --chainid                    the specified chainid, default 300
                    --ip                               node ip,default 127.0.0.1
                    --port                          node p2p port,default 16790
                    --rpcport                   node rpc api port,default 6790
                    --wsport                    node websocket port,default 3790
                    --dashport                node dashboard api port,default 1090
        """
        line = self.parse(line)
        rootDir =  config["datadir"]
        groupid = findFlag(line,'--groupid',str(DEFAULT_GROUP_ID))
        chainid = findFlag(line,'--chainid',str(DEFAULT_CHAIN_ID))
        ip = findFlag(line,'--ip',str(DEFAULT_IP))
        p2pPort = findFlag(line,'--port',str(DEFAULT_P2P_PORT))
        rpcPort = findFlag(line,'--rpcport',str(DEFAULT_RPC_PORT))
        wsPort = findFlag(line,'--wsport',str(DEFAULT_WS_PORT))
        dashboardPort = findFlag(line,'--dashport',str(DEFAULT_DASHBOARD_PORT))

        #create node key and account
        print("[INFO]: auto create node key, and create genesis.json")
        nodePriKey,nodePubKey,nodeAddress = self.createNodeKey({"rootDir":rootDir})
        self.createAccount({"rootDir":rootDir})

        #create genesis and  init chain
        self.createGenesis({"rootDir":rootDir,"groupid":groupid,"chainid":chainid,"ip":ip,"p2pPort":p2pPort})
        self.initChain({"rootDir":rootDir,"groupid":groupid})

        #setup console config file
        bootnodes = findFlag(line,'--bootnodes','')
        if bootnodes != '':
            bootstrapNodes = bootnodes.split(',')
        else:
            bootstrapNodes = ["enode://{0}@{1}:{2}".format(nodePubKey,ip,p2pPort)]
        url = "http://{0}:{1}".format(ip,rpcPort)
        self.setupChainConfig({"rootDir":rootDir,"groupid":groupid,"p2pPort":p2pPort,"rpcPort":rpcPort,"wsPort":wsPort,"dashboardPort":dashboardPort,"bootstrapNodes":bootstrapNodes,"url":url,"status":1})
        
        #start node
        self.startNode(groups[groupid])

        #add admin permision
        time.sleep(3)     
        self.unlockAccount({"addr":config["from"],"password":"0","ip":ip,"rpcPort":rpcPort})
        self.setSuperAdmin({})
        self.addChainAdmin({"addr":config["from"]})
        self.addNodeCMD({"name":nodeAddress,"type":1,"publicKey":nodePubKey,"desc":"","externalIP":ip,"internalIP":ip,"rpcPort":rpcPort,"p2pPort":p2pPort,"owner":nodeAddress,"status":1})

    def do_group(self,line):
        """Create,Join,Leave groups
        Usage:
           group [create|add|join|leave] [command options] [arguments...]
           Options::
                    --groupid                   the specified groupid, default 0
                    --chainid                    the specified chainid, default 300 + groupid
                    --ip                               node ip,default 127.0.0.1
                    --port                          node p2p port,default 16790 + groupid
                    --rpcport                   node rpc api port,default 6790 + groupid
                    --wsport                    node websocket port,default 3790 + groupid
                    --dashport                node dashboard api port,default 1090 + groupid

                unique for join:
                    --creator_pubkey  creator public key,required when you want to join a group
                    --bootnodes             enodes,required when you want to join a group

                unique for add:
                    --pubkey                    node public key to add a new node,required when you add a new node to the chain
                    --addr                          node addr,required when you add a new node to the chain
                    --name                        node name
                    --type                          0=observer;1=validator  
                    --desc                          description of the node


        Options: """
        try:
            line = self.parse(line) 
            if len(line) == 0:
                self.do_help("group")
                return

            if line[0] == "create":
                self.createGroup(line[1:])
            elif line[0] == "add":
                self.addNode(line[1:])
            elif line[0] == "join":
                self.joinGroup(line[1:])
            elif line[0] == "leave":
                self.leaveGroup(line[1:])
            else:
                self.do_help("group")
        except Exception as err:
            print("[ERROR]:" + str(err))
            return

    def do_start(self,line):
        line = self.parse(line)
        groupid = findFlag(line,'--groupid','')
        if groupid == '':
            for id in groups:
                self.startNode(groups[id])
        else:
            self.startNode(groups[groupid])

    def do_stop(self,line):
        line = self.parse(line)
        groupid = findFlag(line,'--groupid','')
        if groupid == '':
            for id in groups:
                self.stopNode(groups[id])
        else:
            self.stopNode(groups[groupid])

    def addNode(self,line):
        pubKey = findFlag(line,'--pubkey','')
        addr = findFlag(line,'--addr','')
        ip = findFlag(line,'--ip','')
        p2pPort = findFlag(line,'--port','')
        if pubKey =='' or addr == '' or ip == '' or p2pPort == '':
            print('[ERROR]: miss required flags,please read command help')
            self.do_help('add')
        
        nodeName = findFlag(line,'--name',addr)
        nodeType = findFlag(line,'--type','0')
        desc = findFlag(line,'--desc','')
        rpcPort = findFlag(line,'--rpcport','0')

        self.addNodeCMD({"name":nodeName,"type":nodeType,"publicKey":pubKey,"desc":desc,"externalIP":ip,"internalIP":ip,"rpcPort":rpcPort,"p2pPort":p2pPort,"owner":addr,"status":1})

    def leaveGroup(self,line):
        groupid = findFlag(line,'--groupid',str(GROUP_ID))
        if groupid == "0":
            print("can not leave group 0")
            return

        if str(GROUP_ID) == groupid:
            switch("0")
        self.stopNode(groups[groupid])

    def do_switch(self,line):
        try: 
            groupid = self.parse(line)[0]
            switch(groupid)
            self.prompt = '[group:' + groupid +']>'
        except Exception as err:
            print("[ERROR]: " + str(err))
            return

    def parse(self,args):
        return args.split()

    def createNodeKey(self,args):
        rootDir =  args["rootDir"]
        nodekeyPath = os.path.join(rootDir,"node.prikey")
        
        if os.path.exists(nodekeyPath):
            print("[INFO]: node key file found. File: " +nodekeyPath)
            address = readAllFromFile(os.path.join(rootDir,"node.address"))
            prikey = readAllFromFile(os.path.join(rootDir,"node.prikey"))
            pubkey = readAllFromFile(os.path.join(rootDir,"node.pubkey"))
            return [prikey,pubkey,address]
        keyinfo= subprocess.check_output(BIN_PATH + '/ethkey genkeypair | sed s/[[:space:]]//g',shell=True).decode('utf-8')
        address = keyinfo[10:50]
        prikey = keyinfo[62:126]
        pubkey = keyinfo[137:265]
        writeToNewFile(nodekeyPath,prikey)
        writeToNewFile(os.path.join(rootDir,"node.pubkey"),pubkey)
        writeToNewFile(os.path.join(rootDir,"node.address"),address)
        print("[INFO]: Create node key successfully. File: " +nodekeyPath)
        return [prikey,pubkey,address]
    
    def createAccount(self,args):
        rootDir =  args["rootDir"]
        cmd = "{0}/platone --datadir {1} account new <<EOF\n0\n0\nEOF".format(BIN_PATH,rootDir)
        ret = subprocess.check_output(cmd,shell=True).decode('utf-8')
        address = "0x" + ret.split("Address: {")[1].split("}")[0]

        config["from"] = address
        onConfigUpdate()
        print("[INFO]: create address " + address + " successfully")

    def createGenesis(self,args):
        rootDir =  args["rootDir"]
        groupid = args["groupid"]
        chainid = args["chainid"]
        ip = args["ip"]
        p2pPort = args["p2pPort"]
        groupDir = os.path.join(rootDir,'group_' + groupid)
        genesisPath = os.path.join(groupDir,'genesis.json')
        if os.path.exists(genesisPath):
            print("[INFO]: group " + groupid + " find genesis.json successfully")
            return
        mkdir(groupDir)
        genesis = json.loads(json.dumps(GENESIS_TEMPLATE))
        genesis["config"]["interpreter"] = "all"    
        genesis["config"]["chainId"] = int(chainid)

        creatorPubKey = args.get("creatorPubKey",readAllFromFile(os.path.join(rootDir,"node.pubkey")))
        creator_enode="enode://{0}@{1}:{2}".format(creatorPubKey,ip,p2pPort)    
        genesis["config"]["istanbul"]["validatorNodes"] = [creator_enode]        
        genesis["config"]["istanbul"]["suggestObserverNodes"] = [creator_enode] 
        writeToNewFile(genesisPath,json.dumps(genesis,indent=4))      
        print("[INFO]: Create genesis.json successfully. File: " +genesisPath)

    def initChain(self,args):
        rootDir =  args["rootDir"]
        groupid = args["groupid"]

        datadir = os.path.join(rootDir,'group_' + groupid )
        if os.path.exists(os.path.join(datadir, 'platone')):
            print("[INFO]: Data dir found")
            return
        
        subprocess.call("{0}/platone --datadir {1} --keystore {2} init {3}".format(BIN_PATH,datadir, os.path.join(rootDir, 'keystore'),os.path.join(datadir, 'genesis.json')) ,shell=True)
        print("[INFO]: group " + groupid + " built successfully")
    
    def setupChainConfig(self,args):
        groupid = args["groupid"]
        #keys = {'p2pPort','rpcPort','wsPort','dashboardPort','bootstrapNodes','url'}

        if groups.get(groupid,{}) == {}:
            tmp = {"id":groupid}
            for key,v in config["groups"][0].items():
                       tmp[key] = args[key]
            config["groups"].append(tmp)
        else:
            for i in range(len(config["groups"])):
                if config["groups"][i]["id"] == groupid:
                    for key,v in config["groups"][0].items():
                        if key == "id":
                            continue
                        config["groups"][i][key] = args[key]
        
        onConfigUpdate()

        groupCfgPath = os.path.join(config["datadir"],"group_" + groupid ,"config.toml")
        if not os.path.exists(groupCfgPath):
            writeToNewFile(groupCfgPath,CHAIN_CONF_TEMPLATE)
        print("[INFO]: setup chain config for group_ " + groupid + " successful")

    def startNode(self,group):
        if group["status"] == 0:
            return
            
        dataPath = os.path.join(config["datadir"],"group_" + group["id"])
        configPath = os.path.join(config["datadir"],"group_" + group["id"] ,"config.toml")
        nodeKeyPath = os.path.join(config["datadir"],"node.prikey")
        keystorePath = os.path.join(config["datadir"],"keystore")

        if not os.path.exists(configPath):
            print("[ERROR]: config file for group_" + group["id"] +" not found")\

        logPath = os.path.join(dataPath,"logs")
        mkdir(logPath)

        cmd = "nohup {0}/platone --config {1} --datadir {2} --nodiscover --nodekey {3} --keystore {4} --port {5} --rpc --rpcport {6} --rpccorsdomain \"*\" --ws --wsaddr 0.0.0.0 --wsport {7} --wsorigins \"*\" --dashboard.host {8} --bootnodes {9} --wasmlog  {10} --wasmlogsize {11} --moduleLogParams '{12}' --gcmode  archive  --debug 1>/dev/null 2>{13}/platone_error.log &".format(
        BIN_PATH,
        configPath,  
        dataPath,
        nodeKeyPath,
        keystorePath,
        group["p2pPort"],
        group["rpcPort"],
        group["wsPort"],
        group["dashboardPort"],
        ",".join(group["bootstrapNodes"]),
        os.path.join(logPath,"wasm_log"),
        "67108864",
        json.dumps({"platone_log":["/"],"__dir__":[logPath],"__size__":["67108864"]}),
        logPath)

        subprocess.Popen(cmd,shell=True,preexec_fn=os.setpgrp)
        print("[INFO]: start group_" + group["id"] + " successfully")

    def stopNode(self,group):
        configPath = os.path.join(config["datadir"],"group_" + group["id"] ,"config.toml")
        cmd = "ps -ef --columns 1000 | grep \"platone --config " + configPath +"\" | grep -v grep | awk \'{print $2}\'"
        ret = subprocess.check_output(cmd,shell=True).decode('utf-8')
        if ret != '':
            subprocess.call("kill -9 " + ret,shell=True)
            print("[INFO]: group_" +  group["id"] + " stopped")

    def createGroup(self,args):
        if str(GROUP_ID) != "0":
            print("Please switch to group 0")
            return
        rootDir =  config["datadir"]
        groupid = findFlag(args,"--groupid","0")
        intGroupID = int(groupid)
        chainid = findFlag(args,"--chainid",str(DEFAULT_CHAIN_ID + intGroupID))
        ip = findFlag(args,"--ip",DEFAULT_IP)
        p2pPort = findFlag(args,"--port",str(DEFAULT_P2P_PORT + intGroupID))
        rpcPort = findFlag(args,"--rpcPort",str(DEFAULT_RPC_PORT + intGroupID))
        wsPort = findFlag(args,"--wsport",str(DEFAULT_WS_PORT + intGroupID))
        dashboardPort = findFlag(args,"--dashport",str(DEFAULT_DASHBOARD_PORT + intGroupID))

        self.createGenesis({"rootDir":rootDir,"groupid":groupid,"chainid":chainid,"ip":ip,"p2pPort":p2pPort})
        self.initChain({"rootDir":rootDir,"groupid":groupid})

        nodePriKey,nodePubKey,nodeAddress = self.createNodeKey({"rootDir":rootDir})
        bootstrapNodes = ["enode://{0}@{1}:{2}".format(nodePubKey,ip,p2pPort)]
        url = "http://{0}:{1}".format(ip,rpcPort)
        self.setupChainConfig({"rootDir":rootDir,"groupid":groupid,"p2pPort":p2pPort,"rpcPort":rpcPort,"wsPort":wsPort,"dashboardPort":dashboardPort,"bootstrapNodes":bootstrapNodes,"url":url,"status":1})
        self.startNode(groups[groupid])

        #add admin permision
        time.sleep(3)
        switch(groupid)
        self.unlockAccount({"addr":config["from"],"password":"0","ip":ip,"rpcPort":rpcPort})
        self.setSuperAdmin({})
        self.addChainAdmin({"addr":config["from"]})
        self.addNodeCMD({"name":nodeAddress,"type":1,"publicKey":nodePubKey,"desc":"","externalIP":ip,"internalIP":ip,"rpcPort":rpcPort,"p2pPort":p2pPort,"owner":nodeAddress,"status":1})
        switch(str(GROUP_ID))

    def joinGroup(self,args):
        rootDir =  config["datadir"]
        groupid = findFlag(args,"--groupid","0")
        intGroupID = int(groupid)
        chainid = findFlag(args,"--chainid",str(DEFAULT_CHAIN_ID + intGroupID))
        ip = findFlag(args,"--ip",DEFAULT_IP)
        p2pPort = findFlag(args,"--port",str(DEFAULT_P2P_PORT + intGroupID))
        rpcPort = findFlag(args,"--rpcPort",str(DEFAULT_RPC_PORT + intGroupID))
        wsPort = findFlag(args,"--wsport",str(DEFAULT_WS_PORT + intGroupID))
        dashboardPort = findFlag(args,"--dashport",str(DEFAULT_DASHBOARD_PORT + intGroupID))

        nodePriKey,nodePubKey,nodeAddress = self.createNodeKey({"rootDir":rootDir})
        if groupid == "0":
            self.createAccount({"rootDir":rootDir})

        creatorPubKey = findFlag(args,"--creator_pubkey","")
        if creatorPubKey == "":
            print("[Error]: creator pubkey  can not be empty")
            return
        self.createGenesis({"rootDir":rootDir,"groupid":groupid,"chainid":chainid,"ip":ip,"p2pPort":p2pPort,"creatorPubKey":creatorPubKey})
        self.initChain({"rootDir":rootDir,"groupid":groupid})

        bootstrapNodes = findFlag(args,"--bootnodes","")
        if bootstrapNodes == "":
            print("[ERROR]: bootstrapNodes can not be empty")
            return
        bootstrapNodes = bootstrapNodes.split(",")

        url = "http://{0}:{1}".format(ip,rpcPort)
        self.setupChainConfig({"rootDir":rootDir,"groupid":groupid,"p2pPort":p2pPort,"rpcPort":rpcPort,"wsPort":wsPort,"dashboardPort":dashboardPort,"bootstrapNodes":bootstrapNodes,"url":url,"status":1})
        self.startNode(groups[groupid])
    
    def setSuperAdmin(self,args):
        contractAddr = USER_MANAGER_ADDR
        contractAbiPath = os.path.join(os.path.dirname(configPath),"contracts","userManager.cpp.abi.json")
        cmd = "{0}/ctool invoke --config {1} --abi {2} --addr {3} --func setSuperAdmin".format(BIN_PATH,ctoolConfPath,contractAbiPath,contractAddr)
        print(cmd)
        subprocess.call(cmd,shell=True)

    def addChainAdmin(self,args):
        addr = args.get("addr","")
        if addr == "":
            return
        contractAddr = USER_MANAGER_ADDR
        contractAbiPath = os.path.join(os.path.dirname(configPath),"contracts","userManager.cpp.abi.json")
        cmd = "{0}/ctool invoke --config {1} --abi {2} --addr {3} --func addChainAdminByAddress --param {4}".format(BIN_PATH,ctoolConfPath,contractAbiPath,contractAddr,addr)
        print(cmd)
        subprocess.call(cmd,shell=True)
    
    def unlockAccount(self,args):
        addr = args["addr"]
        password = args["password"]
        url = "http://{0}:{1}".format(args["ip"],args["rpcPort"])
        http_data=json.dumps({"jsonrpc":"2.0","method":"personal_unlockAccount","params":[addr,password,60],"id":1})
        cmd = "curl -H \"Content-Type: application/json\" --data '{0}'  {1}".format(http_data,url)
        print(cmd)
        subprocess.call(cmd,shell=True)

    def addNodeCMD(self,args):
        nodeJson = {"name":args["name"],"type":int(args["type"]),"publicKey":args["publicKey"],"desc":args["desc"],"externalIP":args["externalIP"],"internalIP":args["internalIP"],"rpcPort":int(args["rpcPort"]),"p2pPort":int(args["p2pPort"]),"owner":args["owner"],"status":args["status"]}
        nodeJsonStr =json.dumps(nodeJson)
        cmd = '{0}/ctool invoke --config {1} --addr {2} --abi {3} --func "add" --param \'{4}\''.format(BIN_PATH,ctoolConfPath,NODE_MANAGER_ADDR,
        os.path.join(os.path.dirname(configPath),"contracts","nodeManager.cpp.abi.json"),nodeJsonStr)
        print(cmd)
        subprocess.call( cmd,shell=True)
        print("[INFO]: add node " + args["name"] + " successfully")

    #################################CLASS END#####################################################################################

def switch(groupid):
    global GROUP_ID
    GROUP_ID = int(groupid)
    group =  groups[groupid]
    f = open(ctoolConfPath,'w')
    f.write("{" + "\n")
    f.write("\"url\":\"" +group["url"]  + "\"," + "\n")
    f.write("\"gas\":\"" + config["gas"] + "\"," + "\n")
    f.write("\"gasPrice\":\"" + config["gasPrice"] + "\"," + "\n")
    f.write("\"from\":\""+ config["from"] + "\"" + "\n")
    f.write("}" + "\n")
    f.close()

def findFlag(args,flagName,default):
    for i in range(len(args)):
        arg = args[i]
        if arg == flagName:
            if  type(default) is bool :
                return True
            elif  i < len(args) - 1:
                return  args[i + 1]
            else:
                return default
    return default

def mkdir(path):
    if os.path.exists(path):
        return
    os.mkdir(path)

def writeToNewFile(path,msg):
    f = open(path,'w')
    f.write(msg)
    f.close()

def readAllFromFile(path):
    f = open(path,"r")
    lines = f.read()
    f.close()
    return lines

def loadGroupsFromConfig(config):
    groups = {}
    for group in config["groups"]:
        groups[group["id"]] = group
    return groups

def onConfigUpdate():
        #write to console config
        f = open(configPath,'w')
        f.write(json.dumps(config,indent=4))
        f.close()

        # reload groups from config
        global groups 
        groups = loadGroupsFromConfig(config)

        #write to ctool config
        switch(str(GROUP_ID))


NODE_MANAGER_ADDR = "0x1000000000000000000000000000000000000002"
USER_MANAGER_ADDR = "0x1000000000000000000000000000000000000001"

DEFAULT_GROUP_ID = 0
DEFAULT_CHAIN_ID = 300
DEFAULT_P2P_PORT = 16790
DEFAULT_RPC_PORT = 6790
DEFAULT_WS_PORT = 3790
DEFAULT_DASHBOARD_PORT = 1090
DEFAULT_IP = "127.0.0.1"

CONF_TEMPLATE = {
    "groups":[
        {
        "id":"0",
        "p2pPort":"",
        "rpcPort":"",
        "wsPort":"",
        "dashboardPort":"",
        "bootstrapNodes":"",
        "url":"",
        "status":1
        }
    ],
    "gas":"0x0",
    "gasPrice":"0x0",
    "from":"",
    "datadir":""
}

GENESIS_TEMPLATE = {
        "config": {
            "chainId": 0,
            "homesteadBlock": 1,
            "eip150Block": 2,
            "eip150Hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
            "eip155Block": 3,
            "eip158Block": 3,
            "byzantiumBlock": 4,
            "interpreter": "",
            "istanbul": {
    	        "timeout": 10000,
      	        "period": 1,
      	        "policy": 0,
      	        "epoch": 1000,
      	        "initialNodes": [],
                "validatorNodes": [],
                "suggestObserverNodes": []
            }
        },
        "nonce": "0x0",
        "timestamp": "0x5c074288",
        "extraData": "0x00",
        "gasLimit": "0xffffffffffff",
        "difficulty": "0x40000",
        "mixHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
        "coinbase": "0x0000000000000000000000000000000000000000",
        "alloc": {
            "0x0000000000000000000000000000000000000011": {
                "balance": "0",
            }
        },
        "number": "0x0",
        "gasUsed": "0x0",
        "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000"
    }

CHAIN_CONF_TEMPLATE = '''
[Eth]
SyncMode = "full"
NoPruning = false
LightPeers = 100
DatabaseCache = 768
TrieCache = 256
TrieTimeout = 3600000000000
MinerGasFloor = 3150000000
MinerGasCeil = 3150000000
MinerGasPrice = 1000000000
MinerRecommit = 3000000000
MinerNoverify = false
EnablePreimageRecording = false

[Eth.TxPool]
Locals = []
NoLocals = false
Journal = "transactions.rlp"
Rejournal = 3600000000000
PriceLimit = 1
PriceBump = 10
AccountSlots = 16
GlobalSlots = 4096
AccountQueue = 64
GlobalQueue = 1024
GlobalTxCount = 3000
Lifetime = 10800000000000

[Eth.GPO]
Blocks = 20
Percentile = 60

[Shh]
MaxMessageSize = 1048576
MinimumAcceptedPOW = 2e-01
RestrictConnectionBetweenLightClients = true

[Node]
UserIdent = "platone"
IPCPath = "platone.ipc"
HTTPHost = "0.0.0.0"
HTTPVirtualHosts = ["localhost"]
HTTPModules = ["db", "eth", "net", "web3", "admin", "personal","txpool","istanbul"]
WSModules = ["net", "web3", "eth", "shh"]

[Node.P2P]
MaxPeers = 50
NoDiscovery = true
StaticNodes = []
TrustedNodes = []
EnableMsgEvents = false

[Node.HTTPTimeouts]
ReadTimeout = 30000000000
WriteTimeout = 30000000000
IdleTimeout = 120000000000

[Dashboard]
Host = "localhost"
Refresh = 5000000000
'''

if __name__ == '__main__':
    try :
        # ENV
        BIN_PATH = os.path.join(os.path.abspath('..'),'bin')
        mkdir(os.path.join(os.path.abspath('..'),'data'))
        # read config filecentos 
        configPath = findFlag(sys.argv,"--config",os.path.join(os.path.abspath('..'),'conf/config.json')) 
        if os.path.exists(configPath):
            lines = readAllFromFile(configPath)
            config =  json.loads(lines)
        else:
            config = json.loads(json.dumps(CONF_TEMPLATE))
            config["datadir"] =  os.path.join(os.path.abspath('..'),'data/node_0')
            writeToNewFile(configPath,json.dumps(config,indent=4))
        
        mkdir(config["datadir"])
        groups = loadGroupsFromConfig(config)

        #generate ctool config file
        ctoolConfPath = os.path.join(os.path.dirname(configPath),"ctool.json")
        GROUP_ID = 0
        switch(str(GROUP_ID))
        
        isDirect = findFlag(sys.argv,"--direct",False)
        if not isDirect:
            cli = Cli()
            cli.cmdloop()
        else:
            print(sys.argv)
    except Exception as err:
        print("[ERROR]: " + str(err))