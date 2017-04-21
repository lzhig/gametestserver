游戏测试工具专用http接口

调用url: http://server/GetGameToken
post数据格式:
{
	"username": "mjtest2000",
	"password": "123456",
	"appid": "241d7b34d61b3aea4eb0e3a91d753ecb"
}

返回数据格式:
{
	"ret": 0,
	"userinfo": {
		"uid": 1005959,
		"username": "bWp0ZXN0MQ==",
		"app_id": "241d7b34d61b3aea4eb0e3a91d753ecb",
		"app_name": "\u4e8c\u4eba\u9ebb\u5c06",
		"timestamp": 1470795354,
		"user_ip": "10.63.41.12",
		"access_token": "a7cce98009e31f05e8b6ca492d8ec7ea",
		"extra": "",
		"guide_status": 2
	},
	"balance": {
		"cash": 999999967,
		"coin": 1000000,
		"nm": 0
	}
}
