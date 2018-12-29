

## Clifton Theme Based Johar window site

### Environment Vairable details

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


### Pushing app to cloud foundry

1. Update the environment variables in manaifest.yml 
2. Create a mysql service instance

```
cf create-service p-mysql 100mb mydb
```

3. push the app

```
cf push -f manifest.yml --no-start
```

4. Bind the app to the mysql service instnace

```
cf bind-service johari mydb 
```

5. start the app

```
cf start johari
```




### Build and run locally

1. Make sure you have your $GOPATH set as per https://github.com/golang/go/wiki/SettingGOPATH

2. Build

```
go get github.com/tools/godep
go get github.com/brucemjenks/johari
godep go install
```

3. Set relevant environment varilabe and cd into the johari root path

4.  Make sure you have a mysql DB to use and set the DBURL environment variable.

5. Then execute the binary

```
:> johari 
VCAP_APPLICATION ENV variable not found
Using url http://localhost:8080 for callback
```


### Testing

run [ginkgo](https://github.com/onsi/ginkgo) to execute tests 

```
ginkgo 
```

---

Please fork or submit pull requests to contribute to this project.



