# Standard Transformer UDFs

Standard Transformer provides several built-in functions that are useful for common ML use-cases. These built-in functions are accessible from within expression context.

| Categories | Functions                               |
| ---------- | --------------------------------------- |
| Geospatial | [Geohash](#geohash)                     |
| Geospatial | [S2ID](#s2id)                           |
| Geospatial | [HaversineDistance](#haversinedistance) |
| Geospatial | [PolarAngle](#polarangle)               |
| JSON       | [JsonExtract](#jsonextract)             |
| Time       | [Now](#now)                             |
| Time       | [DayOfWeek](#dayofweek)                 |
| Time       | [IsWeekend](#isweekend)                 |
| Time       | [FormatTimestamp](#formattimestamp)     |
| Time       | [ParseTimestamp](#parsetimestamp)       |
| Time       | [ParseDateTime](#parsedatetime)         |

## Geospatial

### Geohash

Geohash calculates geohash of `latitude` and `longitude` with the given `precision`.

#### Input

| Name      | Description                                                       |
| --------- | ----------------------------------------------------------------- |
| Latitude  | Latitude of the object, in form of JSONPath, array, or variable.  |
| Longitude | Longitude of the object, in form of JSONPath, array, or variable. |
| Precision | Character precision in integer.                                   |

#### Output

`Geohash of location with the given precision.`

#### Example

```
Input:
{
  "latitude": 1.0,
  "longitude": 2.0
}

Standard Transformer Config:
variables:
- name: geohash
  expression: Geohash("$.latitude", "$.longitude", 12)

Output: `"s01mtw037ms0"`
```

### S2ID

S2ID calculates S2ID cell of `latitude` and `longitude` with the given `level`.

#### Input

| Name      | Description                                                       |
| --------- | ----------------------------------------------------------------- |
| Latitude  | Latitude of the object, in form of JSONPath, array, or variable.  |
| Longitude | Longitude of the object, in form of JSONPath, array, or variable. |
| Level     | S2ID level in integer.                                            |

#### Output

`S2ID cell of the location in certain level.`

#### Example

```
Input:
{
  "latitude": 1.0,
  "longitude": 2.0
}

Standard Transformer Config:
variables:
- name: s2id
  expression: S2ID("$.latitude", "$.longitude", 12)

Output: `"1154732743855177728"`
```

### HaversineDistance

HaversineDistance calculates Haversine distance of two points (given by their latitude and longitude).

#### Input

| Name        | Description                                                             |
| ----------- | ----------------------------------------------------------------------- |
| Latitude 1  | Latitude of the first point, in form of JSONPath, array, or variable.   |
| Longitude 1 | Longitude of the first point, in form of JSONPath, array, or variable.  |
| Latitude 2  | Latitude of the second point, in form of JSONPath, array, or variable.  |
| Longitude 2 | Longitude of the second point, in form of JSONPath, array, or variable. |

#### Output

`The haversine distance between 2 points in kilometer.`

#### Example

```
Input:
{
  "pickup": {
    "latitude": 1.0,
    "longitude": 2.0
  },
  "dropoff": {
    "latitude": 1.2,
    "longitude": 2.2
  }
}

Standard Transformer Config:
variables:
- name: haversine_distance
  expression: HaversineDistance("$.pickup.latitude", "$.pickup.longitude", "$.dropoff.latitude", "$.dropoff.longitude")
```

### PolarAngle

PolarAngle calculates polar angles between two points (given by their latitude and longitude) in radian.

#### Input

| Name        | Description                                                             |
| ----------- | ----------------------------------------------------------------------- |
| Latitude 1  | Latitude of the first point, in form of JSONPath, array, or variable.   |
| Longitude 1 | Longitude of the first point, in form of JSONPath, array, or variable.  |
| Latitude 2  | Latitude of the second point, in form of JSONPath, array, or variable.  |
| Longitude 2 | Longitude of the second point, in form of JSONPath, array, or variable. |

#### Output

`The polar angles between 2 points in radian.`

#### Example

```
Input:
{
  "pickup": {
    "latitude": 1.0,
    "longitude": 2.0
  },
  "dropoff": {
    "latitude": 1.2,
    "longitude": 2.2
  }
}

Standard Transformer Config:
variables:
- name: polar_angle
  expression: PolarAngle("$.pickup.latitude", "$.pickup.longitude", "$.dropoff.latitude", "$.dropoff.longitude")
```

## JSON

### JsonExtract

Given a JSON string as value, you can use JsonExtract to extract JSON value from that JSON string.

#### Input

| Name              | Description                                                                          |
| ----------------- | ------------------------------------------------------------------------------------ |
| Parent's JSONPath | Path to JSON key that its value is a JSON string to be extracted.                    |
| Nested's JSONPath | Path to JSON key inside of JSON string which extracted from Parent's JSONPath above. |

#### Output

`JSON value within a JSON string pointed by the first JSONPath argument.`

#### Example

```
Input:
{
  "details": "{\"merchant_id\": 9001}"
}

Standard Transformer Config:
variables:
- name: merchant_id
  valueType: STRING
  expression: JsonExtract("$.details", "$.merchant_id")

Output: `"9001"`
```

## Time

### Now

Return current local timestamp.

#### Input

`None`

#### Output

`Current local timestamp.`

#### Example

```
Standard Transformer Config:
variables:
- name: currentTime
  expression: Now()
```

### DayOfWeek

Return number representations of the day in a week, given the timestamp and timezone.

SUNDAY(0), MONDAY(1), TUESDAY(2), WEDNESDAY(3), THURSDAY(4), FRIDAY(5), SATURDAY(6).

#### Input

| Name      | Description                                                                                      |
| --------- | ------------------------------------------------------------------------------------------------ |
| Timestamp | Unix timestamp value in integer or string format. It accepts JSONPath, arrays, or variable.      |
| Timezone  | Timezone value in string. For example, `Asia/Jakarta`. It accepts JSONPath, arrays, or variable. |

#### Output

`Day number.`

#### Example

```
{
  "timestamp": "1637605459"
}

Standard Transformer Config:
variables:
- name: day_of_week
  expression: DayOfWeek("$.timestamp", "Asia/Jakarta")

Output: `2`
```

### IsWeekend

Return 1 if given timestamp is weekend (Saturday or Sunday), otherwise 0.

#### Input

| Name      | Description                                                                                      |
| --------- | ------------------------------------------------------------------------------------------------ |
| Timestamp | Unix timestamp value in integer or string format. It accepts JSONPath, arrays, or variable.      |
| Timezone  | Timezone value in string. For example, `Asia/Jakarta`. It accepts JSONPath, arrays, or variable. |

#### Output

`1 if weekend, 0 if not`

#### Example

```
{
  "timestamp": "1637445044",
  "timezone": "Asia/Jakarta"
}

Standard Transformer Config:
variables:
- name: is_weekend
  expression: IsWeekend("$.timestamp", "$.timezone")

Output: `1`
```

### FormatTimestamp

FormatTimestamp converts timestamp in given location into formatted date time string.

#### Input

| Name      | Description                                                                                             |
| --------- | ------------------------------------------------------------------------------------------------------- |
| Timestamp | Unix timestamp value in integer or string format. It accepts JSONPath, arrays, or variable.             |
| Timezone  | Timezone value in string. For example, `Asia/Jakarta`. It accepts JSONPath, arrays, or variable.        |
| Format    | Targetted date time format. It follows Golang date time format (https://pkg.go.dev/time#pkg-constants). |

#### Output

`Date time.`

#### Examples

```
Input:
{
  "timestamp": "1637691859"
}

Standard Transformer Config:
variables:
- name: datetime
  expression: FormatTimestamp("$.timestamp", "Asia/Jakarta", "2006-01-02")

Output: `"2021-11-24"`
```

### ParseTimestamp

ParseTimestamp converts timestamp in integer or string format to time.

#### Input

| Name      | Description                                                                                 |
| --------- | ------------------------------------------------------------------------------------------- |
| Timestamp | Unix timestamp value in integer or string format. It accepts JSONPath, arrays, or variable. |

#### Output

`Parsed timestamp.`

#### Examples

```
Input:
{
  "timestamp": "1619541221"
}

Standard Transformer Config:
variables:
- name: parsed_timestamp
  expression: ParseTimestamp("$.timestamp")

Output: `"2021-04-27 16:33:41 +0000 UTC"`
```

### ParseDateTime

ParseDateTime converts datetime given with specified format layout (e.g. RFC3339) into time.

#### Input

| Name      | Description                                                                                         |
| --------- | --------------------------------------------------------------------------------------------------- |
| Date time | Date time value in string format. It accepts JSONPath, arrays, or variable.                         |
| Timezone  | Timezone value in string. For example, `Asia/Jakarta`. It accepts JSONPath, arrays, or variable.    |
| Format    | Date time input format. It follows Golang date time format (https://pkg.go.dev/time#pkg-constants). |

#### Output

`Parsed date time.`

#### Examples

```
Input:
{
  "datetime": "2021-11-30 15:00:00",
  "location": "Asia/Jayapura"
}

Standard Transformer Config:
variables:
- name: parsed_datetime
  expression: ParseDateTime("$.datetime", "$.location", "2006-01-02 15:04:05")

Output: `"2021-11-30 15:00:00 +0900 WIT"`
```
