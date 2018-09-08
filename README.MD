

## Clifton Theme Based Johar window site


| Environment Variable  | Description |
| ------------- | ------------- |
| DBURL  | Optional varialble that provides a mysql database connection string.  By default application will get the connection string from VCAP_SERVICES environmental variable.  Eample string is `root:root@tcp(192.168.64.101:3306)/johari?parseTime=true`  |
| CLIENTID  | Oauth Client ID  |
| CLIENTSECRET  | Oauth Client Secret  |
| AUTH_URL  | Oauth Authorization URL  |
| TOKEN_URL  | Oauth Token URL  |
| OAUTHURLPARAMS  | Custom Params to add to the [AuthCodeURL](https://godoc.org/golang.org/x/oauth2#Config.AuthCodeURL).  exmaple would be `&hd=mydomain.io&access_type=offline` |
| OAUTHDOMAIN  | The Auth Domain users are allowed to connect from.  for example `mydomain.io`  |
| SESSION_NAME  | Cookie Session Name  |
| COOKIE_STORE_KEY  | Cookie Store Key  |
| PRIVATE_ENCRYPTION_KEY | 32 character encyption key |





