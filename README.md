***

## Metrics Query API's
###*Example Channel Data*

**Channel: water_level**

|time|depth|location|reporter|weather|width|
|----|-----|--------|--------|-------|-----|
|2015-12-06T23:13:28.63539517Z   |130 | us    |  kenny   | sunny    | 6   |
|2015-12-06T23:13:28.63539517Z   |153 | china |  yang    | sunny    | 1   |
|2015-12-06T23:13:36.45545326Z   |112 | china |  yang    | sunny    | 5   |
|2015-12-06T23:13:51.054705818Z  |167 | us    |  kenny   | cloudy    | 4   |
|2015-12-06T23:14:13.526779874Z  |109 | china |  kenny   | cloudy   | 8   |
|2015-12-06T23:14:38.867923309Z  |143 | us    |  yang    | cloud    | 9   |
|2015-12-06T23:14:43.356459782Z  |153 | us    |  yang    | sunny    | 10  |
|2015-12-06T23:14:47.312731741Z  |155 | china |  kenny   | sunny    | 3   |
|2015-12-06T23:14:50.094782447Z  |135 | us    |  yang    | sunny    | 3   |
|2015-12-06T23:14:53.467367472Z  |120 | china |  kenny   | cloudy   | 0   |
|2015-12-06T23:14:56.146985608Z  |null| china |  yang    | sunny    | 5   |


###*Time Range Expression Syntax*

All timestamps are unix milliseconds from epoch.

An example time range including start and end time would be:

```
time_range=1449436224077:1449436235379
```

Time range without end time would imply that the current timestamp is the end time.

```
time_range=1449436224077:
```

All time ranges should contain the start time.

###*Time Interval Expression Syntax*

Time intervals are expressed in the format of `\d+[usmhdw]`.

For example, `1u` means 1 microsecond, `1s` means 1 second, `2m` means 2 minutes, `12h` means 12 hours and `3d` means 3 days, `1w` means 1 week.

###*Tagging/Filtering Expression Syntax*

Supported operations are: `eq`, `ne`, `lt`, `gt`, `le`, `ge`. However, only `eq` and `ne` are supported by tagging expression so far.

Multiple expressions are comma separated.

**tagging expression example**

```
tags=location:eq:china,reporter:ne:kenny
```
**filtering expression example**

```
filters=width:gt:2,depth:le:130
```

Logical grouping on these expressions is not currently supported. Comma separated expressions will by default treated as logical `AND`.

###*Supported Summary Type*

avg/mean, min, max, median, sum, count, last, first, top_n, percentile_n

###*Summary Grouping Syntax*

Some times even the gauges need to be grouped by tags, example would be:

```
group_by=location,weather
```

The order in the `group_by` parameter matters.

Only tags can be grouped.

###*Query for Field Values*

This is used to query for a single value matching all the expressions, applied with summary_type.

Adding `group_by` will result in one value per group.

Since the value is aggregated according to `summary_type`, the returned value won't have a timestamp attached to it.

**request**

```
GET /channels/<channel_id>/query?field=<field>&tags=<tagging expression>&filters=<filtering expression>&summary_type=<summary_type>&group_by=<group_by>&time_range=<time_range expression>
```

`field` and `summary_type` are required.

`tags`, `filters`, `group_by` and `time_range` are optional.

**example**

```
GET /channels/1/query?field=depth&tags=reporter:eq:yang&filters=width:gt:4,width:le:10&group_by=location,weather&summary_type=avg&time_range=1449436224077:
```

Notice that the selected field is `depth`, but the filters are on width. They can be different, as long as they are all defined on the channel.

**response**

```
{
  "water_level": {
    "depth": {
      "avg": {
        "location=china,weather=sunny": { "value": 112 },
        "location=us,weather=cloudy": { "value": 143 },
        "location=us,weather=sunny": { "value": 153 }
      }
    }
  }
}
```

If queried without `group_by`, such as:

```
GET /channels/1/query?field=depth&tags=reporter:eq:yang&filters=width:gt:4,width:le:10&summary_type=avg&time_range=1449436224077:
```
**response**

```
{
  "water_level": {
    "depth": {
      "avg": { "value": 136 }
    }
  }
}
```

###*Query for Time Serials*

This is used to query for a serial of timestamped values with applied summaries for each interval.

**request**

```
GET /channels/<channel_id>/serials?field=<field>&tags=<tagging expression>&filters=<filtering expression>&summary_type=<summary_type>&group_by=<group_by>&time_range=<time_range expression>&time_interval=<interval expression>
```

`field`, `time_range`, `time_interval`, `summary_type` are required.

`tags`, `filters`, `group_by` are optional.

**example**

```
GET /channels/1/serials?field=depth&tags=reporter:eq:yang&filters=width:gt:4,width:le:10&group_by=location&summary_type=avg&time_range=1449436224077:1449436235379&time_interval=1s
```

**response**

```
{
  "water_level": {
    "depth": {
      "avg": {
        "location=china": [
          { "timestamp": 1449456213000, "value": null },
          { "timestamp": 1449456214000, "value": null },
          { "timestamp": 1449456215000, "value": 112 },
          { "timestamp": 1449456216000, "value": 109 },
          { "timestamp": 1449456217000, "value": null },
          { "timestamp": 1449456218000, "value": null },
          { "timestamp": 1449456219000, "value": null },
          { "timestamp": 1449456220000, "value": 155 },
          { "timestamp": 1449456221000, "value": 120 },
          { "timestamp": 1449456222000, "value": 173 }
        ],
        "location=us": [
          { "timestamp": 1449456213000, "value": null },
          { "timestamp": 1449456214000, "value": 130 },
          { "timestamp": 1449456215000, "value": null },
          { "timestamp": 1449456216000, "value": null },
          { "timestamp": 1449456217000, "value": 167 },
          { "timestamp": 1449456218000, "value": 143 },
          { "timestamp": 1449456219000, "value": 153 },
          { "timestamp": 1449456220000, "value": null },
          { "timestamp": 1449456221000, "value": null },
          { "timestamp": 1449456222000, "value": null }
        ]
      }
    }
  }
}
```

The timestamp returned in response is always unix milliseconds from epoch.

Most of the time, grouping is not necessary.

**example**

```
GET /channels/1/serials?field=depth&tags=reporter:eq:yang&filters=width:gt:4,width:le:10&summary_type=avg&time_range=1449436224077:1449436235379&time_interval=1s
```

**response**

```
{
  "water_level": {
    "depth": {
      "avg": [
        { "timestamp": 1449456213000, "value": null },
        { "timestamp": 1449456214000, "value": 130 },
        { "timestamp": 1449456215000, "value": 112 },
        { "timestamp": 1449456216000, "value": 109 },
        { "timestamp": 1449456217000, "value": 167 },
        { "timestamp": 1449456218000, "value": 143 },
        { "timestamp": 1449456219000, "value": 153 },
        { "timestamp": 1449456220000, "value": 155 },
        { "timestamp": 1449456221000, "value": 120 },
        { "timestamp": 1449456222000, "value": 173 }
      ]
    }
  }
}
```

###*Query for Raw Data*

This is usually for testing purposes.

**request**

```
GET /channels/<channel_id>/raw?fields=<fields>&tags=<tagging expression>&filters=<filtering expression>&time_range=<time_range expression>
```

`fields` and `time_range` are required.

`tags`, `filters`, `order`, `limit` and `offset` are optional.

`summary_type`, `time_interval` and `group_by` are omitted.

`fields` can include multiple fields to be returned at a time.

**example**

```
GET /channels/1/raw?fields=depth,width&tags=reporter:eq:yang&filters=width:gt:4,width:le:10&time_range=1449436224077:1449436235379&order=desc&limit=3&offset=3
```

**response**

```
{
  "water_level": {
    "depth": [
      { "timestamp": 1449456219387, "value": 153 },
      { "timestamp": 1449456218385, "value": 143 },
      { "timestamp": 1449456217385, "value": 167 }
    ],
    "width": [
      { "timestamp": 1449456219387, "value": 10 },
      { "timestamp": 1449456218385, "value": 9 },
      { "timestamp": 1449456217385, "value": 4 }
    ]
  }
}
```
