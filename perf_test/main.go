package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	// "github.com/shirou/gopsutil/cpu"
	// "github.com/shirou/gopsutil/mem"
	// "github.com/shirou/gopsutil/net"

	"github.com/BCOSnetwork/BCOS-Go/core/types"
	cli "github.com/BCOSnetwork/BCOS-Go/ethclient"
)

var (
	// 公共参数
	rpcURL     = flag.String("url", "ws://127.0.0.1:6790", "节点url")
	configPath = flag.String("configPath", "", "配置文件")
	// 性能测试参数
	contractAddress = flag.String("contractAddress", "0x0000000000000000000000000000000000000011", "合约地址，用于合约压测,当地址不为空时，启用合约压测")
	abiPath         = flag.String("abiPath", "./", "待测合约的abi文件相对路径")
	//funcParams            = flag.String("funcParams", "", "待测合约的接口及参数")
	txType                = flag.Int("txType", 0, "指定发送的交易类型")
	benchmark             = flag.Bool("benchmark", false, "是否开启benchmark")
	blockDuration         = flag.Int("blockDuration", 10, "性能测试的区块区间数")
	chanValue             = flag.Uint("chanValue", 1000, "每秒最大压力")
	deployContractAddress = flag.String("deployContractAddress", "", "部署合约地址")
	registerContractNum   = flag.Int("registerContractNum", 10, "注册合约总数")
	stressTest            = flag.Int("stressTest", 0, "是否开启压力测试, 1:简单合约测试(setter&getter) 2. 复杂合约测试(CNS)")
	consensusTest         = flag.Bool("consensusTest", false, "是否开启共识测试")
	realtimeTps           = flag.Bool("realtimeTps", false, "是否开启实时压力测试")
)

const (
	consensusLogFile      = "./consensus_data.txt"
	simpleContractLogFile = "./contract_data.txt"
)

var (
	contractName     string = "demoContractNo" + strconv.Itoa(rand.Int())
	versionFrontPart        = "1.1.1."

	preTimestamp int64

	curNum       int64
	curBlockNum  int64
	curTimestamp int64

	count int = 0
)

func main() {
	var wg sync.WaitGroup

	flag.Parse()
	parseConfigJson(*configPath)

	inChan := make(chan int, *chanValue)
	defer close(inChan)
	closeChan := make(chan int)
	defer close(closeChan)

	if *consensusTest {
		// 计算平均共识时间
		wg.Add(1)
		go func() {
			var start time.Time
			var elapsed time.Duration

			// Begin to dial node
			client, err := cli.Dial(*rpcURL)
			if err != nil {
				fmt.Println("client connection error:", err.Error())
				os.Exit(1)
			}
			defer client.Close()
			heads := make(chan *types.Header, 1)
			sub, err := client.SubscribeNewHead(context.Background(), heads)
			if err != nil {
				fmt.Println("Failed to subscribe to head events", "err", err)
			}
			defer sub.Unsubscribe()

			handle, err := os.Create(consensusLogFile)
			if err != nil {
				panic(err)
			}
			defer handle.Close()
			w := bufio.NewWriter(handle)

			cur := time.Now()
		perf:
			for {
				select {
				case <-heads:
					curElapsed := time.Since(cur)
					fmt.Fprintf(w, "当前区块共识时间 %4.3f 秒\n", curElapsed.Seconds())
					fmt.Printf("当前区块共识时间 %4.3f 秒\n", curElapsed.Seconds())
					cur = time.Now()

					count++
					if count == 1 {
						start = time.Now()
					} else if count > *blockDuration {
						elapsed = time.Since(start)
						break perf
					}

					// 计算每个区块的tps
					if count == 1 {
						curBlockNum = getCurrentBlockNum()
						curNum, curTimestamp = getBlockTxNum(curBlockNum)
					} else {
						preTimestamp = curTimestamp

						curBlockNum = getCurrentBlockNum()
						curNum, curTimestamp = getBlockTxNum(curBlockNum)
						if curTimestamp != preTimestamp {
							fmt.Printf("当前区块tps：%4.2f tx/s\n\n", float64(curNum)/float64(curTimestamp-preTimestamp))
						}
					}

				case <-closeChan:
					panic("stopped, interuppted by signal...")
				default:

				}
			}
			fmt.Fprintf(w, "平均共识时间 %4.3f 秒\n", elapsed.Seconds()/float64(*blockDuration))
			w.Flush()

			wg.Done()
		}()
	}

	if *stressTest != 0 {
		// 等待newHead事件
		if *consensusTest {
			for count <= 2 {
				time.Sleep(1 * time.Second)
			}
		}

		wg.Add(1)
		go func() {
			handle, err := os.Create(simpleContractLogFile)
			if err != nil {
				panic(err)
			}
			defer handle.Close()
			w := bufio.NewWriter(handle)

			//Judging whether this contract exists or not
			if *stressTest == 1 {
				// 简单合约调用
				if !getContractByAddress(*contractAddress) {
					panic("-contractAddress [...] is not exist ...")
				}
			} else if *stressTest == 2 {
				if !getContractByAddress(*deployContractAddress) {
					panic("-deployContractAddress [...] is not exist ...")
				}
			} else {
				panic("-stressTest is invalid ...")
			}

			var tries int = 0
			startNum := getCurrentBlockNum()
		stressTest:
			for {
				tries++
				var str string
				if *stressTest == 1 {
					// 简单合约调用
					str = "invokeNotify(\"test1....\")"
				} else {
					// 复杂合约调用
					str = "cnsRegister(" + contractName + "," + versionFrontPart + strconv.Itoa(tries) + "," +
						*deployContractAddress + ")"
				}

				//fmt.Println(str, *registerContractNum, tries)
				fmt.Fprintln(w, str)

				err, _ = invoke(*contractAddress, *abiPath, str, *txType)
				//time.Sleep(2 * time.Millisecond)
				//inChan <- 1

				if tries >= *registerContractNum {
					// 查询成功注册合约总数
					var startTimestamp int64
					var endTimestamp int64

					if *realtimeTps {
						break stressTest
					}

					last := len(txHashList) - 1
					for getTxByHash(txHashList[last]) == false &&
						getTxByHash(txHashList[last-2]) == false &&
						getTxByHash(txHashList[last-4]) == false &&
						getTxByHash(txHashList[last-6]) == false {
						select {
						case <-closeChan:
							panic("stopped, interuppted by signal...")
						default:
						}
					}

					endNum := getCurrentBlockNum()

					sum, startTimestamp := getBlockTxNum(startNum)
					for i := startNum + 1; i <= endNum; i++ {
						var n int64
						n, endTimestamp = getBlockTxNum(i)
						sum += n
					}

					if *stressTest == 2 {
						registerContracts := getPerfResults().(int64)
						fmt.Printf("成功注册合约总数：%d\n", registerContracts)
						fmt.Fprintf(w, "成功注册合约总数：%d\n", registerContracts)
					}

					fmt.Println("###start", startTimestamp, "end", endTimestamp)
					fmt.Println("hash list", txHashList)
					fmt.Printf("注册合约tps：%f tx/s\n", float64(sum)/float64(endTimestamp-startTimestamp))
					fmt.Fprintf(w, "注册合约tps：%f tx/s\n", float64(sum)/float64(endTimestamp-startTimestamp))
					w.Flush()

					break
				}

				select {
				case <-closeChan:
					panic("stopped, interuppted by signal...")
				default:
					continue
				}

			}
			wg.Done()
		}()
	}
	/*
		var wg sync.WaitGroup
		wg.Add(1)
		// GetSendSpeed 获取发送速度
		go func() {
			now := time.Now()
			for {
				if time.Since(now).Seconds() >= 1 {
					select {
					case <-inChan:
						length := ReadChan(inChan)
						fmt.Printf("Send Speed:%d/s\n", length)
						now = time.Now()
					case <-closeChan:
						panic("too bad")
						wg.Done()
					}
				}
			}
		}()
	*/

	if *benchmark {
		wg.Add(1)
		// 计算内存使用率
		go func(interval time.Duration) {
			var totalSum float64
			var freeSum float64
			var usedPercentSum float64

		benchmarkMem:
			for {
				var count int = 10
				for ; count > 0; count-- {
					v, _ := mem.VirtualMemory()
					totalSum += float64(v.Total)
					freeSum += float64(v.Free)
					usedPercentSum += float64(v.UsedPercent)
					time.Sleep(100 * time.Millisecond)
				}

				fmt.Printf("Total: %v, Free:%v, UsedPercent:%4.2f%%\n",
					totalSum/float64(count), freeSum/float64(count), usedPercentSum/float64(count))

				time.Sleep(interval)

				select {
				case <-closeChan:
					break benchmarkMem
				default:
					continue
				}
			}
			wg.Done()
		}(1000 * time.Millisecond)

		wg.Add(1)
		// 统计cpu平均使用率
		go func(interval time.Duration) {
		benchmarkCpu:
			for {
				cpuUsageRates, err := cpu.Percent(interval, true)
				if err != nil {
					fmt.Println(err)
					return
				}

				var sum float64 = 0
				for _, v := range cpuUsageRates {
					sum += v
				}
				average := sum / float64(len(cpuUsageRates))

				fmt.Printf("Cpu usage average rate :%4.2f%%\n", average)

				time.Sleep(interval)

				select {
				case <-closeChan:
					break benchmarkCpu
				default:
					continue
				}
			}
			wg.Done()
		}(1000 * time.Millisecond)

		wg.Add(1)
		// 计算网络带宽
		go func(interval time.Duration) {
		benchmarkNet:
			for {
				stats1, err := net.IOCounters(false)
				if err != nil {
					fmt.Println(err)
				}
				time.Sleep(interval)
				stats2, err := net.IOCounters(false)
				if err != nil {
					fmt.Println(err)
				}
				// unit : bytes/s
				netIoSentSpeed := float64((stats2[0].BytesSent - stats1[0].BytesSent)) / interval.Seconds()
				netIoRecvSpeed := float64((stats2[0].BytesRecv - stats1[0].BytesRecv)) / interval.Seconds()

				fmt.Printf("Net send rate :%f bytes/s, recv rate :%f bytes/s\n", netIoSentSpeed, netIoRecvSpeed)

				select {
				case <-closeChan:
					break benchmarkNet
				default:
					continue
				}
			}
			wg.Done()
		}(1000 * time.Millisecond)
	}

	go Trap(closeChan)
	wg.Wait()
}
