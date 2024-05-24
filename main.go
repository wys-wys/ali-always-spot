// This file is auto-generated, don't edit it. Thanks.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"

	"github.com/Mr-LvGJ/ali-always-spot/pkg/client"
	"github.com/Mr-LvGJ/ali-always-spot/pkg/setting"
)

var (
	AccessKey = pflag.String("access-key", "", "aliyun access key ")
	SecretKey = pflag.String("secret-key", "", "aliyun secret key ")
	RegionId  = pflag.String("region-id", "cn-hongkong", "region id")
)

func main() {
	pflag.Parse()
	setting.InitConfig(&setting.Config{AccessKey: AccessKey, SecretKey: SecretKey, RegionId: RegionId})
	client.SetupEcsClient()
	client.SetupVpcClient()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt, os.Kill)

	errCh := make(chan error)

	go func() {
		errCh <- run()
	}()

	select {
	case err := <-errCh:
		log.Fatalln("error occur, ", err)
	case s := <-c:
		log.Println("exit...", "signal", s.String())
		return
	}
}

func run() error {
	createCh := make(chan struct{})
	createDoneCh := make(chan struct{})
	// 1. Release EIP
	go func() {
		tk := time.NewTicker(30 * time.Minute)
		defer tk.Stop()

		// DescribeEipAddress
		// if not assosiate delete
		// else skip

	}()

	// 2. Loop describe instances
	go func() {
		tk := time.NewTicker(3 * time.Second)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				instances, err := client.DescribeInstances()
				if err != nil {
					log.Println(err)
					continue
				}
				if instances != nil {
					if len(instances.Instance) != 1 {
						log.Println("instance exist, skip...")
					} else {
						createCh <- struct{}{}
						<-createDoneCh
					}
				}
			}
		}
	}()

	// 3. Instance deleted, create new instance
	for {
		select {
		case <-createCh:
			// create instance
			log.Println("begin create new instance...")
			client.RunInstances()
			createDoneCh <- struct{}{}
		}
	}
}
