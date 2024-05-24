package client

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	"github.com/alibabacloud-go/tea/tea"

	"github.com/Mr-LvGJ/ali-always-spot/pkg/setting"
)

var (
	ecsClient *ecs20140526.Client
)

var zoneIds = []string{
	"cn-hongkong-b",
	"cn-hongkong-c",
	"cn-hongkong-d",
}

func SetupEcsClient() {
	config := &openapi.Config{
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
		AccessKeyId: setting.C().AccessKey,
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
		AccessKeySecret: setting.C().SecretKey,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Ecs
	config.Endpoint = tea.String("ecs.cn-hongkong.aliyuncs.com")
	_result, _err := ecs20140526.NewClient(config)
	if _err != nil {
		panic(_err)
	}
	ecsClient = _result
}

func DescribeInstances() (*ecs20140526.DescribeInstancesResponseBodyInstances, error) {
	resp, err := ecsClient.DescribeInstances(&ecs20140526.DescribeInstancesRequest{
		RegionId: setting.C().RegionId,
	})
	if err != nil {
		return nil, err
	}
	return resp.Body.Instances, nil
}

func GetImageId() (*string, error) {
	res, err := ecsClient.DescribeImages(&ecs20140526.DescribeImagesRequest{
		RegionId:        setting.C().RegionId,
		ImageOwnerAlias: tea.String("self"),
	})
	if err != nil {
		slog.Error("describe image failed...", "err", err)
		return nil, err
	}
	if len(res.Body.Images.Image) == 0 {
		slog.Info("self image not found...")
		res, err := ecsClient.DescribeImages(&ecs20140526.DescribeImagesRequest{
			RegionId:    setting.C().RegionId,
			ImageFamily: tea.String(" acs:debian_12_4_x64"),
		})
		if err != nil {
			slog.Error("describe image failed...", "err", err)
			return nil, err
		}
		return res.Body.Images.Image[0].ImageId, nil
	}
	return res.Body.Images.Image[0].ImageId, nil

}

func DescribePriceAndGetAvailableZone() (*string, error) {
	for _, zoneId := range zoneIds {
		res, err := ecsClient.DescribePrice(&ecs20140526.DescribePriceRequest{
			RegionId:     setting.C().RegionId,
			ZoneId:       tea.String(zoneId),
			ResourceType: tea.String("instance"),
			InstanceType: tea.String("ecs.n1.tiny"),
			SystemDisk: &ecs20140526.DescribePriceRequestSystemDisk{
				Size:     tea.Int32(20),
				Category: tea.String("cloud_efficiency"),
			},
			SpotStrategy: tea.String("SpotAsPriceGo"),
			SpotDuration: tea.Int32(0),
		})
		if err != nil {
			slog.Error("describe price failed...", "err", err)
			continue
		}
		price := res.Body.PriceInfo.Price.TradePrice
		if *price > 0.2 {
			slog.Warn("zone: ", zoneId, "price large than 0.2, ", price)
			continue
		}
		return &zoneId, nil
	}
	return nil, errors.New("can not find discount zone")
}

func getOrCreateSecurityGroupId() (*string, error) {
	result, err := ecsClient.DescribeSecurityGroups(&ecs20140526.DescribeSecurityGroupsRequest{
		RegionId: setting.C().RegionId,
	})
	if err != nil {
		slog.Error("DescribeSecurityGroups err	", err)
		return nil, err
	}

	if *result.Body.TotalCount == 0 {
		group, err := ecsClient.CreateSecurityGroup(&ecs20140526.CreateSecurityGroupRequest{
			RegionId: setting.C().RegionId,
		})
		if err != nil {
			slog.Error("CreateSecurityGroup err", err)
			return nil, err
		}
		return group.Body.SecurityGroupId, nil
	}

	return result.Body.SecurityGroups.SecurityGroup[0].SecurityGroupId, nil
}

func RunInstances() (*string, error) {
	var retInfo string
	zone, err := DescribePriceAndGetAvailableZone()
	if err != nil {
		slog.Error("DescribePriceAndGetAvailableZone failed", err)
		return nil, err
	}

	imageId, err := GetImageId()
	if err != nil {
		slog.Error("GetImageId failed", err)
		return nil, err
	}

	sgId, err := getOrCreateSecurityGroupId()
	if err != nil {
		slog.Error("getSecurityGroupId failed", err)
		return nil, err
	}

	vsId, err := getOrCreateVswitchId(*zone)
	if err != nil {
		slog.Error("getVswitchId failed", err)
		return nil, err
	}

	password := generateRandomString(10)

	req := &ecs20140526.RunInstancesRequest{
		RegionId:        setting.C().RegionId,
		ImageId:         imageId,
		InstanceType:    tea.String("ecs.n1.tiny"),
		SecurityGroupId: sgId,
		VSwitchId:       vsId,
		Password:        tea.String(password),
		ZoneId:          zone,
		SystemDisk: &ecs20140526.RunInstancesRequestSystemDisk{
			Size:     tea.String("20"),
			Category: tea.String("cloud_efficiency"),
		},
		Amount:       tea.Int32(1),
		SpotStrategy: tea.String("SpotAsPriceGo"),
		SpotDuration: tea.Int32(0),
	}
	slog.Info("RunInstances request", *req)

	ins, err := ecsClient.RunInstances(req)
	if err != nil {
		slog.Error("RunInstance failed", "err", err)
		return nil, err
	}
	insId := ins.Body.InstanceIdSets.InstanceIdSet[0]

	cbpId, err := getOrCreateCommonBandwidthPackages()
	if err != nil {
		slog.Error("getOrCreateCommonBandwidthPackages failed", err)
		return nil, err
	}

	eipId, ipAddr, err := getOrCreateEip()
	if err != nil {
		slog.Error("getOrCreateEip failed", err)
		return nil, err
	}

	err = addEipToCBP(eipId, cbpId)
	if err != nil {
		slog.Error("addEipToCBP failed", err)
		return nil, err
	}

	err = associateEipAddress(insId, eipId)
	if err != nil {
		slog.Error("associateEipAddress failed", err)
		return nil, err
	}

	retInfo = fmt.Sprintf("InstanceId：%s, Password：%s, EIP: %s", *insId, password, *ipAddr)

	return &retInfo, nil
}

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charSet[rand.Intn(len(charSet))]
	}
	return "L" + string(result)
}
