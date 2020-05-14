
## DeviceParameter 

The `DeviceParameter` type is a subType of [`deviceEvent`](https://developer.tidepool.org/data-model/device-data/types/deviceEvent/). 

It gives details on the parameter, its current value and its previous value. 

```json
{
  "type": "deviceEvent",
  "subType": "deviceParameter",
  "time": "2020-01-20T08:17:07.920Z",
  "timezoneOffset": 60,
  "name": "Name of the Parameter",
  "value": "Current value of the parameter",
  "units": "%",
  "lastUpdateDate": "2020-01-20T08:10:00.000Z",
  "previousValue": "Previous value of the parameter",  
  "level": "1",
  "minValue": "0",
  "maxValue": "10",
  "processed": "yes",
  "linkedSubType": ["DeviceParameterName1", "DeviceParameterName2"]
} 
```

### name

Name of the device parameter.

```
QUICK SUMMARY
Required:
    platform: yes
```

### value

Value of the device parameter.

```
QUICK SUMMARY
Required:
    platform: yes
```

### units

Units that is used for this parameter. If this field is not set a default unit will be used. 

```
QUICK SUMMARY
Required:
    platform: no
```

### timezoneOffset

Derived from [Tidepool Commmon Fields](https://developer.tidepool.org/data-model/device-data/common.html#timezoneoffset)

```
QUICK SUMMARY
Required:
    platform: no
```

### lastUpdateDate 

The effective date the value was changed. Can be different than [`time`](https://developer.tidepool.org/data-model/device-data/common.html#time) that is the date of the upload. 

```
QUICK SUMMARY
Required:
    platform: yes
```

### previousValue

The previous value for the same device parameter that has been replaced by the new value defined in `value` 

```
QUICK SUMMARY
Required:
    platform: no
```

### level

The device parameter level, as it is defined in the Device System. As of now those levels are defined as one of the following values: 1, 2 or 3

```
QUICK SUMMARY
Required:
    platform: yes
```

### minValue 

Minimal value of the device parameter. 

```
QUICK SUMMARY
Required:
    platform: no
```

### maxValue 

Maximal value of the device parameter. 

```
QUICK SUMMARY
Required:
    platform: no
```

### processed 

Is it a parameter value that has been calculated by the System ? If the value is entered directly by the end-user, `processed` is set to `no`

```
QUICK SUMMARY
Required:
    platform: no
Enum:
    `yes`
    `no`    
```

### linkedSubType

List of SubTypes used to process the parameter.

```
QUICK SUMMARY
Required:
    platform: if processed = `yes`
```
