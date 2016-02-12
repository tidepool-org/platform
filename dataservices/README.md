## Examples

### Login

POST /auth/login

```
curl -X POST -i -u <user-name> -d '' "http://localhost:8009/auth/login"
```

Response

```
HTTP/1.1 200 OK

x-tidepool-session-token: <your-token>

{"userid":"b676436f60","username":"<user-name>","emails":["<user-name>"],"roles":[""],"termsAccepted":"2016-02-02T11:26:27+13:00","emailVerified":true}
```

### Save dataset

POST /dataset

```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X POST -d '[{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]' http://localhost:8077/dataset
```

Response
```
{
  "Dataset": [
    {
      "deliveryType": "scheduled",
      "rate": 2,
      "duration": 21600000,
      "suppressed": null,
      "type": "basal",
      "deviceTime": "2014-06-11T06:00:00Z",
      "time": "2014-06-11T06:00:00Z",
      "timezoneOffset": 0,
      "conversionOffset": 0,
      "deviceId": "tools"
    }
  ],
  "Errors": ""
}
```

### Get dataset

GET /dataset

```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X GET http://localhost:8077/dataset
```


### Save dataset

GET /data

```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X GET http://localhost:8077/dataset
```