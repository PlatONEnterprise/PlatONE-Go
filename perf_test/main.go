package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	types "github.com/PlatONnetwork/PlatON-Go/core/types"
	cli "github.com/PlatONnetwork/PlatON-Go/ethclient"
)

var (
	// 公共参数
	rpcURL     = flag.String("url", "ws://127.0.0.1:6790", "节点url")
	configPath = flag.String("configPath", "", "配置文件")
	// 性能测试参数
	contractAddress = flag.String("contractAddress", "", "合约地址，用于合约压测,当地址不为空时，启用合约压测")
	abiPath         = flag.String("abiPath", "./", "待测合约的abi文件相对路径")
	funcParams      = flag.String("funcParams", "", "待测合约的接口及参数")
	txType          = flag.Int("txType", 0, "指定发送的交易类型")
	benchmark       = flag.Bool("benchmark", false, "是否开启benchmark")
	blockDuration   = flag.Int("blockDuration", 1, "性能测试的区块区间数")
	chanValue       = flag.Uint("chanValue", 100, "每秒最大压力")
)

func main() {
	var wg sync.WaitGroup

	flag.Parse()

	client, err := cli.Dial(*rpcURL)
	if err != nil {
		fmt.Println("client connection error:", err.Error())
		os.Exit(1)
	}
	defer client.Close()
	heads := make(chan *types.Header, 16)
	sub, err := client.SubscribeNewHead(context.Background(), heads)
	if err != nil {
		fmt.Println("Failed to subscribe to head events", "err", err)
	}
	defer sub.Unsubscribe()

	var count int = 0
	var start time.Time
	var elapsed time.Duration

perf:
	for {
		select {
		case head := <-heads:
			fmt.Println("new tx root hash", head.TxHash.Hex())
			count++
			if count == 1 {
				start = time.Now()
			} else if count > *blockDuration {
				elapsed = time.Since(start)
				break perf
			}
		}
	}

	fmt.Printf("平均共识时间 %4.3f 秒", elapsed.Seconds()/float64(*blockDuration))

	inChan := make(chan int, *chanValue)
	defer close(inChan)
	closeChan := make(chan int)
	defer close(closeChan)

	go func() {
		for {
			// 简单合约调用
			err = invoke(*contractAddress, *abiPath, *funcParams, *txType)
			if err != nil {
				panic(err.Error())
			}
			inChan <- 1
		}
	}()

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

	if *benchmark {
		wg.Add(1)
		// 计算内存使用率
		go func() {
			var totalSum float64
			var freeSum float64
			var usedPercentSum float64

			for count := 100; count > 0; count-- {
				v, _ := mem.VirtualMemory()
				totalSum += float64(v.Total)
				freeSum += float64(v.Free)
				usedPercentSum += float64(v.UsedPercent)
				time.Sleep(100 * time.Millisecond)
			}

			fmt.Printf("Total: %v, Free:%v, UsedPercent:%4.2f%%\n",
				totalSum/100.0, freeSum/100.0, usedPercentSum/100.0)
			wg.Done()
		}()

		wg.Add(1)
		// 统计cpu平均使用率
		go func(interval time.Duration) {
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
			wg.Done()
		}(1000 * time.Millisecond)

		wg.Add(1)
		// 计算网络带宽
		go func() {
			stats1, err := net.IOCounters(false)
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(100 * time.Millisecond)
			stats2, err := net.IOCounters(false)
			if err != nil {
				fmt.Println(err)
			}
			// unit : bytes/s
			netIoSentSpeed := float64((stats2[0].BytesSent - stats1[0].BytesSent) / 10.0)
			netIoRecvSpeed := float64((stats2[0].BytesRecv - stats1[0].BytesRecv) / 10.0)

			fmt.Printf("Net send rate :%f bytes/s, recv rate :%f bytes/s\n", netIoSentSpeed, netIoRecvSpeed)
			wg.Done()
		}()
	}

	go Trap(closeChan)

	wg.Wait()
}
