package utils

import (
	"fmt"
	"github.com/smartwalle/alipay/v3"
	"net/url"
)

var privateKey = "MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCphuWiGGJl7gshrTK1aMJ1osstCGEW3PdkATxFFx/UhNgH7SXD7IMoupYThQjzE1E8bN9TalGNMBEXzY4Jj0vZoJMEZeG8QY4Wqk+E66LVJV2hTUjSOy+pyDXDIY5Fe42X5Gd0BTKUs/p6NT7sMQWRdJWf2d1SfVUtgc5CSWzXJEpzatUGgo+/5bNJfSdM3MFpMEnFQ1qONP3+LxlumCZEE7JzQL8CYIzE+EduoGn/oSjFht/Ed67eOMVun8xY196l8VwQJ69oQ9npYD9CznphJMBg9RnK+PGFg7Hd/LfK+l7idadGIkaMWdy+wkArPHF+N3E1EjVDDrqEReNe2rNLAgMBAAECggEATLb/yXeRZ6DuJqIy9Ubf4M33dXx6OxdpWDb66CULwWcQP54AXgX0YIT3DwQ/xYfzVg3KjfvpxaN/yK67XPYck/IHNZFJMqDTxMvzSio7uIq9MRZBIumnqwQv2AqiUC0WKSsx6Z3EguhjK9RWkhmo9Ga2ouy7K+4NkpdQGE0T1NMZycw/wPc5LcwBq7sL+FiurFQAAM7Wq0W0l+71kLaQxOWaAUKSj5OVpWHDCHI1Oy2n+bIVVtBpy6HpEFCjQanu6BHk664HHH7l5L0C0UuW4K/GwDfPD7itMPKmkb5CccgDU+2rJuodVHyUbunqMSF2yAZ1UYPzZ8Ln7IxJkIqlYQKBgQDVF4t2xgdgINq8yggDu6wN5ZEM9XiPp+5HNpctGr4iChzQIpBR6SJiyRUC7vNnzRudgEpabKWEPy4i5GXHcZbHfOfxcyW3CHNaAVgxNa20T7KyybKBaA8b35V7Ca/p0MECW4nRTKwfRB4QBVdYx6aRh/tlR3B8yrfhDSHgZ+W58QKBgQDLqan92PmKVcXIIzWk8v59EmYHvl/Ewt1KzgUAc53JZSobekgri7PF/90VXbcgy2eJq6VxXpLGuPCuWLuxckEPRUYbuplGL/4xuHp8H83g/TKhO6tk71drPFy5vTyZxUVyAQen+l2X3U848sh09C75JViZLtXx7WU95SU4Fm+k+wKBgFR12JljNFktrIVXroWMRU3cx/lS8k4+SXuAb7s49lOXnoQAryNIPJDbErDu9RsXePKcftwIZDJeuHKsBItgwlqfb2+MLE630sDB96rJk+f8DuA+gbo4/IQXwq/Zzxfl3hqJHb8PnMlnvmKrO0u9FpBoTYR/JF7SGr/g7KR9idiBAoGAMa5itHTgcrl3tNm59VH9eJ8rWoo7LHFosB3PpIuPmxhdjDRpNI4wvYUr9lFVId/ckv3XLu+mGGn29GDa8G9xpXr9njgHudJtTM22u166xz6cwi4fIlEsXxFrgTfDd7NivGu55WUyvaAT+k1nTvheGRLeKQf+0rRZdR7X1HXMqE8CgYEAz+W8h838dg19fkExKdTsDbS/ROV0mlFDAMEJsBle2vwQDZv/79I8RwnDrZx6+bYWHE6HdDxEj65GrOiD5t7Ghtlxaiu8sXLdhto/Enc/AwRhNYCNTmB72WnwrmiBbj2OGV8xP5YRkwZFNZLe2GOLgseK19OU4FUXcUx9IPq3uOk="
var appID = "2021003186695004"

func GetPayment() {
	var client, err = alipay.New(appID, privateKey, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = client.LoadAliPayPublicKey(
		"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAoi+YDiVXNhK83QygxTvQYQeTVTnbSaKj5xLKOcVWYIOE6V0jW6o8ihCe5mk6eU3fU6UR38LSuYEb4ggU/cTC508HkAEHObEEZkiS9b61buKC8eNmcUtZqZA/coL39OPDoPNddzWbYwSGL8LGqMrXKTdH3pBrmSqialxbKLshTqC2Qq1Zo86tA2gYi+WPSIHhPBSSQwHF2qzX4okVLSXHGrHfhLmrzQDrWkGMcb+B0/hkYtM5N25EmUa9knWjZyADC35Wmdq4jmuDJmrDGtyDAQ4fNj5y5FCbrwjgYd2c+kZBsvGNJVbMCOvlgjuzzM8AzuITw2JC4CD8GHZTT82h6wIDAQAB")
	if err != nil {
		fmt.Println(err)
		return
	}
	// 将 key 的验证调整到初始化阶段

	var p = alipay.TradeWapPay{}
	p.NotifyURL = ""
	p.ReturnURL = ""
	p.Subject = "标题"
	p.OutTradeNo = "传递一个唯一单号"
	p.TotalAmount = "10.00"
	p.ProductCode = "QUICK_WAP_WAY"
	var url *url.URL
	url, err = client.TradeWapPay(p)
	if err != nil {
		fmt.Println(err)
	}

	var payURL = url.String()
	fmt.Println(payURL)

}
