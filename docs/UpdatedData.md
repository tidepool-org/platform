Here we document the data types that have to be created or udpated with new fields: 

- meal/wizard
- bolus
- physical activity
- deviceEvent - Alarm 
- food
- deviceEvent - Zen mode
- deviceEvent - Private Mode

## wizard 

The wizard object comes with an optional `recommended` structure that can be leveraged for our purpose. This structure is composed of 3 optional floating point value fields:
- carb: amount of insulin to cover the the total grams of carbohydrate input (`carbInput`)
- correction: amount of insulin recommended by the device to bring the PWD to their target blood glucose.
- net: total amount of recommended insulin

Here is an example of what can be sent with the related meaning:
- `recommended.net` is the system recommendation
- `bolus.normal` is the value delivered by the insulin pump.
- `bolus.expectedNormal` is the original value that has been requested to the insulin pump.
- `bolus.prescriptor` is a new field that describes the origin of the bolus. Details are defined in the below food section. 

And the additional field we would need:
- `inputTime` is a UTC string timestamp that defines at what time the patient has entered the meal. This field is optional. It takes the same format as `time` field.

```json
{
  "time": "2020-05-12T08:50:08.000Z",
  "timezoneOffset": 120,
  "deviceTime": "2020-05-12T08:50:08",
  "inputTime": "2020-05-12T08:45:08.000Z",
  "deviceId": "IdOfTheDevice",
  "type": "wizard",
  "carbInput": 50,
  "insulinOnBoard": 5.0,
  "recommended": {
    "net": 5
  },
  "units": "mg/dL",
  "bolus": {
    "time": "2020-05-12T08:50:08.000Z",
    "timezoneOffset": 120,
    "deviceTime": "2020-05-12T08:50:08",
    "deviceId": "IdOfTheDevice",
    "type": "bolus",
    "subType": "normal",
    "normal": 3.5,
    "expectedNormal": 4.0, 
    "prescriptor": "hybrid"
  }
}
```

## food 

As of now we don't have the information of the origin of the rescueCarbs value, is it a patient decision, is it a system recommendation, and in that case what was the recommendation vs the actual value.

Here we are introducing 2 new fields in the food object:
- `prescribedNutrition`: same structure as nutrition. It's an optional field. It gives the value that has been recommended by the system. 
- `prescriptor`: is the origin of the `rescuecarbs` object. This field is optional in most of the cases. 
    - range of values: `auto | manual | hybrid`
    - `auto`: nutrition and prescribedNutrition are equal
    - `manual`: prescribedNutrition is ignored
    - `hybrid`: nutrition and prescribedNutrition are __not equal__, `prescribedNutrition` is mandatory in that case. 

```json
{
  "type": "food",
  "meal": "rescuecarbs",
  "nutrition": {
    "carbohydrate" : {
      "net": 20,
      "units": "grams"
    }
  },
  "prescribedNutrition": {
    "carbohydrate" : {
      "net": 30,
      "units": "grams"
    }
  },
  "prescriptor": "hybrid",
  "meal": "rescuecarbs",
  "deviceId": "IdOfTheDevice",
  "deviceTime": "2020-05-12T06:50:08",
  "time": "2020-05-12T06:50:08.000Z",
  "timezoneOffset": 120
}

```

## bolus

3 types of bolus events are available as of now in the system:
- normal
- square
- dual/square

Here we are introducing 2 new fields in the bolus objects:
- `prescriptor`: same as above in `food`. This field is optional. 
- `insulinOnBoard`: amount of active insulin estimated by the system. This field will be accepted when `prescriptor` is either `auto` or `hybrid`. It will be ignored for `manual` entries.

```json
{
  "time": "2020-05-12T08:50:08.000Z",
  "timezoneOffset": 120,
  "deviceTime": "2020-05-12T08:50:08",
  "deviceId": "IdOfTheDevice",
  "type": "bolus",
  "subType": "normal",
  "normal": 3.5,
  "expectedNormal": 4.0, 
  "prescriptor": "hybrid"
}
```

## biphasic bolus

A `biphasic` bolus is a 2 parts bolus that is defined by the system. Below is the definition for this new type of bolus that leverages most of the fields from `normal` bolus. The subType associated to this type of bolus is `biphasic`.
We add the following fields:
- `eventId`: unique ID provided by the client that is used to link the 2 parts of the bolus.
- part: `1 | 2`. It's either the first part or the second part of the bolus. We will see that the first part of the bolus has to contain additional mandatory fields. 
- `normal` and `expectedNormal` are similar to what is defined in `normal` bolus. 
- `linkedBolus` defined the second part of the bolus at the time the first part is created. It's an estimated bolus that may be modified by the system. This section is mandatory for any `"part":1` object. 
  - `linkedBolus.normal`: the expected value for the second part of the biphasic bolus. The actual value is provided by the `"part":2` object.
  - `linkedBolus.duration`: the expected duration between the first and the second part of the biphasic bolus. The actual duration is provided by the `"part":2` object through the effective time of this second object. The duration structure is leveraged from structure already used in other objects such as physical activity.
- `prescriptor`: same as above in `food`. This field is optional. 

__Note #1__: this type of bolus can be used in the wizard object the same way we use the `normal` bolus.

__Note #2__: the `"part":2` object is not mandatory. The system can decide to cancel this second part of the bolus. 

```json
{
  "time": "2020-05-12T12:00:00.000Z",
  "timezoneOffset": 120,
  "deviceTime": "2020-05-12T12:00:08",
  "deviceId": "IdOfTheDevice",
  "type": "bolus",
  "subType": "biphasic",
  "eventId": "Bo123456789",
  "part": 1,
  "normal": 3.5,
  "expectedNormal": 4.0, 
  "linkedBolus": {
    "normal": 3.5,
    "duration": { 
    	  "value": 60,
    	  "units": "minutes"
    }
  },
  "prescriptor": "system"
}
{
  "time": "2020-05-12T12:50:00.000Z",
  "timezoneOffset": 120,
  "deviceTime": "2020-05-12T12:50:08",
  "deviceId": "IdOfTheDevice",
  "type": "bolus",
  "subType": "biphasic",
  "eventId": "Bo123456789",
  "part": 2,
  "normal": 3.5,
  "prescriptor": "system"
}
```

## physical activity

We need additional fields to get the time at which the physical activity is created, and the last time it was updated by the patient:
- `inputTime` is a UTC string timestamp that defines at what time the patient has entered the physical activity. This field is optional. It takes the same format as `time` field.
- `eventType`: type of event, either `start` or `end`
  - `start` defines the beginning of the event. The `duration` is the estimated one. The `time` field gives the actual start time of the event.
  - `stop` gives the end of the event. The `duration` is the actual duration. The `time` field gives the actual end time of the event. 
- `eventId`: unique ID provided by the client that is used to link stop and start events.

In the below example, the physical activity is entered on the handset at 8:00am. It starts at 8:50am for 60 minutes. Finally the stop event says that activity stopped at 9:40, that is tha actual duration was 50 minutes. This last information was entered at 10:00am.

```json
{
    "type": "physicalActivity",
    "reportedIntensity": "medium",
    "duration": { 
    	"value": 60,
    	"units": "minutes"
    },
    "eventType": "start",
    "eventId": "AP123456789",
    "deviceId": "DBLG1.1.6",
    "deviceTime": "2016-07-12T23:52:47",
    "inputTime": "2020-05-12T08:00:08.000Z",
    "time": "2020-05-12T08:50:08.000Z",
    "timezoneOffset": 60
}
{
    "type": "physicalActivity",
    "reportedIntensity": "medium",
    "duration": { 
    	"value": 50,
    	"units": "minutes"
    },
    "eventType": "stop",
    "eventId": "AP123456789",
    "deviceId": "DBLG1.1.6",
    "deviceTime": "2016-07-12T23:52:47",
    "inputTime": "2020-05-12T10:00:08.000Z",
    "time": "2020-05-12T09:40:08.000Z",
    "timezoneOffset": 60
}
```

## Alarm events
Leveraging the `deviceEvent` type with the already defined `alarm` subType. We add couple of fields to get more details on alarms and acknowledgement. 

- `alarmLevel`: `alarm | alert` 
- `alarmCode`: code of the alarm. This field is optional. 
- `alarmLabel`: label or description of the alarm. This field is optional. 
- `eventId`: unique Id of the event generated by the client system. This ID will be used to reconciliate data for the same event. 
- `eventType`: `start | stop` is the type of event for the given alarm.
  - `start`: alarm created by the system
  - `stop`: the system has received the patient acknowledge. 

For a given alarm that has been acknowledged by the patient, we will receive 2 deviceEvents of subType `alarm` with the same `eventId`:
- the first one gives the creation time on the system, `eventType`: `start`
- the second one gives the patient acknowledge, `eventType`: `stop`

```json
{
  "type": "deviceEvent",
  "subType": "alarm",
  "alarmType": "handset",
  "alarmLevel": "alarm", 
  "alarmCode": "123456",
  "alarmLabel": "Label of the alarm",
  "eventId": "123456789",
  "eventType": "alarm",
  "deviceId": "Id12345",
  "deviceTime": "2018-02-01T00:00:00",
  "time": "2020-05-12T08:50:08.000Z",
  "timezoneOffset": 60
}
```

## Zen mode && Confidential mode

Leveraging the `deviceEvent` type and creating 2 new subTypes with the same structure: `zen` and `confidential`.

- `subType`: `zen | confidential`
- `duration`: is a structured object that gives the duration of the event. 
- `eventType`: `start | stop` is the type of event for the given event.
  - `start`: event created by the system. The `duration` attached to this object is the expected duration of the event.
  - `stop`: the event is stopped. The `duration` attached to this object is the actual duration of the event.
- `eventId`: unique ID provided by the client that is used to link stop and start events.

```json
{
  "type": "deviceEvent",
  "subType": "zen",
  "eventType": "start", 
  "eventId": "Zen123456789",
  "duration": { 
    "value": 3,
    "units": "hours"
  },
  "deviceId": "Id12345",
  "deviceTime": "2018-02-01T00:00:00",
  "time": "2020-05-12T08:50:08.000Z",
  "timezoneOffset": 60
}
{
  "type": "deviceEvent",
  "subType": "confidential",
  "eventType": "start", 
  "eventId": "Conf123456789",
  "duration": { 
    "value": 180,
    "units": "minutes"
  },
  "deviceId": "Id12345",
  "deviceTime": "2018-02-01T00:00:00",
  "time": "2020-05-12T08:50:08.000Z",
  "timezoneOffset": 60
}
```
