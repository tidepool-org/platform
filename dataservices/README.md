## Running

This service requires the rest of the tidpool platform also running at this stage. Please see https://github.com/tidepool-org/tools#runservers

Running this service can be done with the command

```
make build
```

```
bin/dataservices
```

## Examples

### Login

POST /auth/login

Example Request
```
curl -X POST -i -u <user-name> -d '' "http://localhost:8009/auth/login"
```

Example Response
```
HTTP/1.1 200 OK

x-tidepool-session-token: <your-token>

{"userid":"b676436f60","username":"<user-name>","emails":["<user-name>"],"roles":[""],"termsAccepted":"2016-02-02T11:26:27+13:00","emailVerified":true}
```

### Save dataset

POST /dataset

Example Request
```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <session-token>" -X POST -d '[{"time": "2016-02-05T07:02:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T07:02:00", "deviceId": "IR1285-79-36047-15", "type": "basal", "deliveryType": "scheduled", "scheduleName": "DEFAULT", "rate": 0.1, "duration": 10800000, "payload": {"logIndices": [58 ] }, "index": 58 }, {"time": "2016-02-05T13:34:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T13:34:00", "deviceId": "IR1285-79-36047-15", "type": "deviceEvent", "subType": "status", "status": "suspended", "reason": {"suspended": "manual"}, "payload": {"logIndices": [24 ] }, "index": 24 }, {"time": "2016-02-05T10:02:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T10:02:00", "deviceId": "IR1285-79-36047-15", "type": "basal", "deliveryType": "scheduled", "scheduleName": "DEFAULT", "rate": 2, "duration": 12780000, "payload": {"logIndices": [57 ] }, "index": 57 }, {"time": "2016-02-05T13:35:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T13:35:00", "deviceId": "IR1285-79-36047-15", "type": "basal", "deliveryType": "suspend", "scheduleName": "DEFAULT", "rate": 0, "duration": 1260000, "payload": {"logIndices": [56 ] }, "index": 56 }, {"time": "2016-02-05T13:56:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T13:56:00", "deviceId": "IR1285-79-36047-15", "type": "deviceEvent", "subType": "status", "status": "resumed", "reason": {"resumed": "manual"}, "payload": {"logIndices": [24 ] }, "index": 24 }, {"time": "2016-02-05T14:04:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T14:04:00", "deviceId": "IR1285-79-36047-15", "type": "deviceEvent", "subType": "status", "status": "suspended", "reason": {"suspended": "manual"}, "payload": {"logIndices": [23 ] }, "index": 23 }, {"time": "2016-02-05T13:56:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T13:56:00", "deviceId": "IR1285-79-36047-15", "type": "basal", "deliveryType": "scheduled", "scheduleName": "DEFAULT", "rate": 2, "duration": 540000, "payload": {"logIndices": [55 ] }, "index": 55 }, {"time": "2016-02-05T15:29:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T15:29:00", "deviceId": "IR1285-79-36047-15", "type": "deviceEvent", "subType": "status", "status": "resumed", "reason": {"resumed": "manual"}, "payload": {"logIndices": [23 ] }, "index": 23 }, {"time": "2016-02-05T15:30:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T15:30:00", "deviceId": "IR1285-79-36047-15", "type": "deviceEvent", "subType": "status", "status": "suspended", "reason": {"suspended": "manual"}, "payload": {"logIndices": [22 ] }, "index": 22 }, {"time": "2016-02-05T15:50:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T15:50:00", "deviceId": "IR1285-79-36047-15", "type": "deviceEvent", "subType": "status", "status": "resumed", "reason": {"resumed": "manual"}, "payload": {"logIndices": [22 ] }, "index": 22 }, {"time": "2016-02-05T14:05:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T14:05:00", "deviceId": "IR1285-79-36047-15", "type": "basal", "deliveryType": "suspend", "scheduleName": "DEFAULT", "rate": 0, "duration": 6480000, "payload": {"logIndices": [54 ] }, "index": 54 }, {"time": "2016-02-05T15:53:00.000Z", "timezoneOffset": 0, "clockDriftOffset": 0, "conversionOffset": 0, "deviceTime": "2016-02-05T15:53:00", "deviceId": "IR1285-79-36047-15", "type": "basal", "deliveryType": "scheduled", "scheduleName": "DEFAULT", "rate": 1.75, "duration": 432000000, "payload": {"logIndices": [53 ] }, "index": 53, "annotations": [{"code": "animas/basal/flat-rate"} ] } ]' http://localhost:8077/dataset/<userid>
```

Example Response
```
{
  "Dataset": [
    {
      "deliveryType": "scheduled",
      "scheduleName": "DEFAULT",
      "rate": 0.1,
      "duration": 10800000,
      "suppressed": null,
      "id": "56dded32dd28e548fa00001a",
      "userId": "b676436f60",
      "deviceId": "IR1285-79-36047-15",
      "uploadId": "",
      "deviceTime": "2016-02-05T07:02:00",
      "time": "2016-02-05T07:02:00.000Z",
      "timezoneOffset": 0,
      "conversionOffset": 0,
      "clockDriftOffset": 0,
      "type": "basal",
      "payload": {
        "logIndices": [
          58
        ]
      },
      "annotations": null,
      "createdTime": "2016-03-08T10:05:54+13:00"
    },
    {
      "subType": "status",
      "status": "resumed",
      "reason": {
        "resumed": "manual"
      },
      "id": "56dded32dd28e548fa00002c",
      "userId": "b676436f60",
      "deviceId": "IR1285-79-36047-15",
      "uploadId": "",
      "deviceTime": "2016-02-05T15:50:00",
      "time": "2016-02-05T15:50:00.000Z",
      "timezoneOffset": 0,
      "conversionOffset": 0,
      "clockDriftOffset": 0,
      "type": "deviceEvent",
      "payload": {
        "logIndices": [
          22
        ]
      },
      "annotations": null,
      "createdTime": "2016-03-08T10:05:54+13:00"
    },
    {
      "deliveryType": "scheduled",
      "scheduleName": "DEFAULT",
      "rate": 1.75,
      "duration": 432000000,
      "suppressed": null,
      "id": "56dded32dd28e548fa000030",
      "userId": "b676436f60",
      "deviceId": "IR1285-79-36047-15",
      "uploadId": "",
      "deviceTime": "2016-02-05T15:53:00",
      "time": "2016-02-05T15:53:00.000Z",
      "timezoneOffset": 0,
      "conversionOffset": 0,
      "clockDriftOffset": 0,
      "type": "basal",
      "payload": {
        "logIndices": [
          53
        ]
      },
      "annotations": [
        {
          "code": "animas/basal/flat-rate"
        }
      ],
      "createdTime": "2016-03-08T10:05:54+13:00"
    },
    ....
  ],
  "Errors": ""
}
```

### Save blob

POST /blob

Example Request
```
curl -i -H "x-tidepool-session-token: <your-token>" -F "data=@my.blob" -X POST http://localhost:8077/blob/<userid>
```


### Get dataset

GET /dataset

Example Request
```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X GET http://localhost:8077/dataset/<userid>
```



`type` (optional) : The Tidepool data type to search for. Only objects with a type field matching the specified type param will be returned.
can be /userid?type=smbg or a comma seperated list e.g /userid?type=smgb,cbg . If is a comma seperatedlist, then objects matching any of the sub types will be returned.

`subType` (optional) : The Tidepool data subtype to search for. Only objects with a subtype field matching the specified subtype param will be returned. can be /userid?subtype=physicalactivity or a comma seperated list e.g /userid?subtypetype=physicalactivity,steps . If is a comma seperatedlist, then objects matching any of the types will be returned.

`startDate` (optional) : Only objects with 'time' field equal to or greater than start date will be returned . Must be in ISO date/time format e.g. 2015-10-10T15:00:00.000Z

`endDate` (optional) : Only objects with 'time' field less than to or equal to start date will be returned . Must be in ISO date/time format e.g. 2015-10-10T15:00:00.000Z

Example Request (with optional params)
```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X GET 'http://localhost:8077/dataset/<userid>?type=smbg&subType=linked&startDate=2015-10-10T15:00:00.000Z&endDate=2015-10-10T15:00:00.000Z'
```

```
{
  "Dataset": [
    {
      "_active": true,
      "_groupId": "85e9e57e20",
      "_id": "56dded320713525292566935",
      "_schemaVersion": 1,
      "annotations": [
        {
          "code": "animas/basal/flat-rate"
        }
      ],
      "createdTime": "2016-03-08T10:05:54+13:00",
      "deliveryType": "scheduled",
      "deviceId": "IR1285-79-36047-15",
      "deviceTime": "2016-02-05T15:53:00",
      "duration": 432000000,
      "id": "56dded32dd28e548fa000030",
      "payload": {
        "logIndices": [
          53
        ]
      },
      "rate": 1.75,
      "scheduleName": "DEFAULT",
      "time": "2016-02-05T15:53:00.000Z",
      "type": "basal",
      "uploadId": "",
      "userId": "b676436f60"
    },
    {
      "_active": true,
      "_groupId": "85e9e57e20",
      "_id": "56dded320713525292566934",
      "_schemaVersion": 1,
      "createdTime": "2016-03-08T10:05:54+13:00",
      "deliveryType": "suspend",
      "deviceId": "IR1285-79-36047-15",
      "deviceTime": "2016-02-05T14:05:00",
      "duration": 6480000,
      "id": "56dded32dd28e548fa00002e",
      "payload": {
        "logIndices": [
          54
        ]
      },
      "rate": 0,
      "scheduleName": "DEFAULT",
      "time": "2016-02-05T14:05:00.000Z",
      "type": "basal",
      "uploadId": "",
      "userId": "b676436f60"
    },
    ...
  ],
  "Errors": ""
}
```


### Get datum

GET /data

Example Request
```
curl -H "Content-Type: application/json" -H "x-tidepool-session-token: <your-token>" -X GET http://localhost:8077/data/<userid>/<datumid>
```