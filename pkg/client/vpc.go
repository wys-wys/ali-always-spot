// This file is auto-generated, don't edit it. Thanks.
package client

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
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
