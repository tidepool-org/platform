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
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X POST -d '[{"deviceTime": "2014-06-11T06:00:00.000Z", "time": "2014-06-11T06:00:00.000Z", "timezoneOffset": 0, "conversionOffset": 0, "type": "basal", "deliveryType": "scheduled", "scheduleName": "Standard", "rate": 2, "duration": 21600000, "deviceId": "tools"}]' http://localhost:8077/dataset/<userid>
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
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X GET http://localhost:8077/dataset/<userid>
```

`type` (optional) : The Tidepool data type to search for. Only objects with a type field matching the specified type param will be returned.
can be /userid?type=smbg or a comma seperated list e.g /userid?type=smgb,cbg . If is a comma seperatedlist, then objects matching any of the sub types will be returned.

`subType` (optional) : The Tidepool data subtype to search for. Only objects with a subtype field matching the specified subtype param will be returned. can be /userid?subtype=physicalactivity or a comma seperated list e.g /userid?subtypetype=physicalactivity,steps . If is a comma seperatedlist, then objects matching any of the types will be returned.

`startDate` (optional) : Only objects with 'time' field equal to or greater than start date will be returned . Must be in ISO date/time format e.g. 2015-10-10T15:00:00.000Z

`endDate` (optional) : Only objects with 'time' field less than to or equal to start date will be returned . Must be in ISO date/time format e.g. 2015-10-10T15:00:00.000Z

e.g.
```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X GET 'http://localhost:8077/dataset/<userid>?type=smbg&subType=linked&startDate=2015-10-10T15:00:00.000Z&endDate=2015-10-10T15:00:00.000Z'
```

### Get datum

GET /data

```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X GET http://localhost:8077/data/<userid>/<datumid>
```