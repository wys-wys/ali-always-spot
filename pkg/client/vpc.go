// This file is auto-generated, don't edit it. Thanks.
package client

import (
	"log/slog"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	vpc20160428 "github.com/alibabacloud-go/vpc-20160428/v6/client"

	"github.com/Mr-LvGJ/ali-always-spot/pkg/setting"
)

var vpcClient *vpc20160428.Client

func SetupVpcClient() (_result *vpc20160428.Client, _err error) {
	// 工程代码泄露可能会导致 AccessKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考。
	// 建议使用更安全的 STS 方式，更多鉴权访问方式请参见：https://help.aliyun.com/document_detail/378661.html。
	config := &openapi.Config{
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
		AccessKeyId: setting.C().AccessKey,
		// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
		AccessKeySecret: setting.C().SecretKey,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Vpc
	config.Endpoint = tea.String("vpc.cn-hongkong.aliyuncs.com")
	_result = &vpc20160428.Client{}
	_result, _err = vpc20160428.NewClient(config)
	if _err != nil {
		panic(_err)
	}
	vpcClient = _result
	return _result, _err
}

func getOrCreateVswitchId(zoneId string) (*string, error) {
	result, err := ecsClient.DescribeVSwitches(&ecs20140526.DescribeVSwitchesRequest{
		RegionId: setting.C().RegionId,
		ZoneId:   tea.String(zoneId),
	})
	if err != nil {
		return nil, err
	}
	if *result.Body.TotalCount == 0 {
		vpcClient.CreateDefaultVpc(&vpc20160428.CreateDefaultVpcRequest{
			RegionId: setting.C().RegionId,
		})

		vs, err := vpcClient.CreateDefaultVSwitch(&vpc20160428.CreateDefaultVSwitchRequest{
			RegionId: setting.C().RegionId,
			ZoneId:   tea.String(zoneId),
		})
		if err != nil {
			return nil, err
		}
		return vs.Body.VSwitchId, nil
	}
	return result.Body.VSwitches.VSwitch[0].VSwitchId, nil
}

func HasAvaliableEipAddress() ([]string, error) {
	var releaseIps []string
	resp, err := vpcClient.DescribeEipAddresses(&vpc20160428.DescribeEipAddressesRequest{
		RegionId: setting.C().RegionId,
	})
	if err != nil {
		slog.Error("DescribeEipAddress failed", "err", err)
		return nil, err
	}
	for _, eip := range resp.Body.EipAddresses.EipAddress {
		if *eip.Status == "Available" {
			releaseIps = append(releaseIps, *eip.AllocationId)
		}
	}
	return releaseIps, nil
}

func ReleaseEips(eips []string) error {
	for _, eip := range eips {
		cbpId, _ := getOrCreateCommonBandwidthPackages()
		vpcClient.RemoveCommonBandwidthPackageIp(&vpc20160428.RemoveCommonBandwidthPackageIpRequest{
			RegionId:           setting.C().RegionId,
			BandwidthPackageId: cbpId,
			IpInstanceId:       tea.String(eip),
		})
		_, err := vpcClient.ReleaseEipAddress(&vpc20160428.ReleaseEipAddressRequest{
			RegionId:     setting.C().RegionId,
			AllocationId: tea.String(eip),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func getOrCreateCommonBandwidthPackages() (*string, error) {
	resp, err := vpcClient.DescribeCommonBandwidthPackages(&vpc20160428.DescribeCommonBandwidthPackagesRequest{
		RegionId: setting.C().RegionId,
	})
	if err != nil {
		return nil, err
	}
	if *resp.Body.TotalCount == 0 {
		resp, err := vpcClient.CreateCommonBandwidthPackage(&vpc20160428.CreateCommonBandwidthPackageRequest{
			RegionId:           setting.C().RegionId,
			Bandwidth:          tea.Int32(1000),
			InternetChargeType: tea.String("PayByDominantTraffic"),
		})
		if err != nil {
			return nil, err
		}
		return resp.Body.BandwidthPackageId, nil
	}
	return resp.Body.CommonBandwidthPackages.CommonBandwidthPackage[0].BandwidthPackageId, nil
}

func getOrCreateEip() (*string, *string, error) {
	resp, err := vpcClient.DescribeEipAddresses(&vpc20160428.DescribeEipAddressesRequest{
		RegionId: setting.C().RegionId,
	})
	if err != nil {
		return nil, nil, err
	}
	for _, eip := range resp.Body.EipAddresses.EipAddress {
		if *eip.Status == "Available" {
			return eip.AllocationId, eip.IpAddress, nil
		}
	}
	res, err := vpcClient.AllocateEipAddress(&vpc20160428.AllocateEipAddressRequest{
		RegionId:           setting.C().RegionId,
		InstanceChargeType: tea.String("PostPaid"),
		InternetChargeType: tea.String("PayByTraffic"),
	})
	if err != nil {
		return nil, nil, err
	}
	return res.Body.AllocationId, res.Body.EipAddress, nil
}

func addEipToCBP(eipId, cbpId *string) error {
	resp, err := vpcClient.DescribeCommonBandwidthPackages(&vpc20160428.DescribeCommonBandwidthPackagesRequest{
		RegionId:           setting.C().RegionId,
		BandwidthPackageId: cbpId,
	})
	if err != nil {
		return err
	}
	for _, eip := range resp.Body.CommonBandwidthPackages.CommonBandwidthPackage[0].PublicIpAddresses.PublicIpAddresse {
		if *eip.AllocationId == *eipId {
			return nil
		}
	}

	_, err = vpcClient.AddCommonBandwidthPackageIp(&vpc20160428.AddCommonBandwidthPackageIpRequest{
		RegionId:           setting.C().RegionId,
		BandwidthPackageId: cbpId,
		IpInstanceId:       eipId,
	})
	return err
}

func associateEipAddress(insId, eipId *string) error {
	_, err := vpcClient.AssociateEipAddress(&vpc20160428.AssociateEipAddressRequest{
		RegionId:     setting.C().RegionId,
		InstanceId:   insId,
		AllocationId: eipId,
	})
	return err
}
