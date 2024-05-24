// This file is auto-generated, don't edit it. Thanks.
package main

import (
	"log"
	"log/slog"
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
	// 1. Release available EIP
	go func() {
		tk := time.NewTicker(30 * time.Minute)
		defer tk.Stop()

		for {
			select {
			case <-tk.C:
				eips, err := client.HasAvaliableEipAddress()
				if err != nil {
					slog.Error("HasAvaliableEipAddress query failed", err)
					continue
				}
				err = client.ReleaseEips(eips)
				if err != nil {
					slog.Error("ReleaseEip failed", "err", err)
				}
			}
		}

		// DescribeEipAddress
		// if not assosiate delete
		// else skip

	}()

	// 2. Loop describe instances
	go func() {
		tk := time.NewTicker(30 * time.Second)
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
					if len(instances.Instance) != 0 {
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
			res, err := client.RunInstances()
			if err != nil {
				slog.Error("create instance failed", "err")
			} else {
				slog.Info("create instance success!!", "ins info", *res)
			}
			createDoneCh <- struct{}{}
		}
	}
}
