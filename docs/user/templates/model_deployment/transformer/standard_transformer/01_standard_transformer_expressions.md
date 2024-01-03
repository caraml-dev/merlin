<!-- page-title: Standard Transformer Expressions -->
<!-- parent-page-title: Standard Transformer -->
# Standard Transformer Expressions

Standard Transformer provides several built-in functions that are useful for common ML use-cases. These built-in functions are accessible from within expression context.

| Categories | Functions                                                    |
| ---------- | -------------------------------------------------------------|
| Geospatial | [Geohash](#geohash)                                          |
| Geospatial | [S2ID](#s2id)                                                |
| Geospatial | [HaversineDistance](#haversinedistance)                      |
| Geospatial | [HaversineDistanceWithUnit](#haversinedistancewithunit)      |
| Geospatial | [PolarAngle](#polarangle)                                    |
| Geospatial | [GeohashDistance](#geohashdistance)                          |
| Geospatial | [GeohashAllNeighbors](#geohashallneighbors)                  |
| Geospatial | [GeohashNeighborForDirection](#geohashneighborfordirection)  |
| JSON       | [JsonExtract](#jsonextract)                                  |
| Statistics | [CumulativeValue](#cumulativevalue)                          |
| Time       | [Now](#now)                                                  |
| Time       | [DayOfWeek](#dayofweek)                                      |
| Time       | [IsWeekend](#isweekend)                                      |
| Time       | [FormatTimestamp](#formattimestamp)                          |
| Time       | [ParseTimestamp](#parsetimestamp)                            |
| Time       | [ParseDateTime](#parsedatetime)                              |
| Series     | [Get](#get)                                                  |
| Series     | [IsIn](#isin)                                                |
| Series     | [StdDev](#stddev)                                            |
| Series     | [Mean](#mean)                                                |
| Series     | [Median](#median)                                            |
| Series     | [Max](#max)                                                  |
| Series     | [MaxStr](#maxstr)                                            |
| Series     | [Min](#min)                                                  |
| Series     | [MinStr](#minstr)                                            |
| Series     | [Quantile](#quantile)                                        |
| Series     | [Sum](#sum)                                                  |
| Series     | [Flatten](#flatten)                                          |
| Series     | [Unique](#unique)                                            |


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

### HaversineDistanceWithUnit

HaversineDistanceWithUnit calculates Haversine distance of two points (given by their latitude and longitude) and given the distance unit

#### Input

| Name          | Description                                                             |
| -----------   | ----------------------------------------------------------------------- |
| Latitude 1    | Latitude of the first point, in form of JSONPath, array, or variable.   |
| Longitude 1   | Longitude of the first point, in form of JSONPath, array, or variable.  |
| Latitude 2    | Latitude of the second point, in form of JSONPath, array, or variable.  |
| Longitude 2   | Longitude of the second point, in form of JSONPath, array, or variable. |
| Distance Unit | Unit of distance measurement, supported unit `km` and `m`               |

#### Output

`The haversine distance between 2 points.`

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
  expression: HaversineDistanceWithUnit("$.pickup.latitude", "$.pickup.longitude", "$.dropoff.latitude", "$.dropoff.longitude", "m")
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

### GeohashDistance
GeohashDistance will calculate haversine distance between two geohash. It will convert a geohash into the center point (latitude, longitude) of that geohash and calculate haversine distance based on that point.

#### Input

| Name          | Description                                              |
|---------------|----------------------------------------------------------|
| Geohash 1     | First geohash, in form of JSONPath, array                |
| Geohash 2     | Second geohash, in form of JSONPath, array               |
| Distance Unit | Unit measurement of distance, supported unit `km` and `m`|

#### Output
`Haversine Distance between two geohash calculated from the center point of that geohash`

#### Example
```
Input:
{
  "pickup_geohash": "qqgggnwxx",
  "dropoff_geohash": "qqgggnweb"
}

Standard Transformer Config:
variables:
- name: geohash_distance
  expression: GeohashDistance("$.pickup_geohash", "$.dropoff_geohash", "m")
```

### GeohashAllNeighbors

GeohashAllNeighbors will find all neighbors of geohash from all directions

#### Input

| Name          | Description                                              |
|---------------|----------------------------------------------------------|
| Geohash 1     | Geohash , in form of JSONPath, array                     |

#### Output
`List of neighbors of given geohash`

#### Example
```
Input:
{
  "pickup_geohash": "qqgggnwxx",
  "dropoff_geohash": "qqgggnweb"
}

Standard Transformer Config:
variables:
- name: geohash_distance
  expression: GeohashAllNeighbors("$.pickup_geohash")
```

### GeohashNeighborForDirection

GeohashNeighborForDirection will find a neighbor of geohash given the direction

#### Input

| Name          | Description                                              |
|---------------|----------------------------------------------------------|
| Geohash 1     | Geohash , in form of JSONPath, array                     |
| Direction     | Direction of that neighbor relatively from geohash. List of accepted direction `north`, `northeast`, `northwest`, `south`, `southeast`, `southwest`, `west`, `east`|

#### Output
`Neighbor of given geohash`

#### Example
```
Input:
{
  "pickup_geohash": "qqgggnwxx",
  "dropoff_geohash": "qqgggnweb"
}

Standard Transformer Config:
variables:
- name: geohash_distance
  expression: GeohashNeighborForDirection("$.pickup_geohash", "north")
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

## Statistics

### CumulativeValue

CumulativeValue is a function that accumulates values based on the index and its predecessors. E.g., `[1, 2, 3] => [1, 1+2, 1+2+3] => [1, 3, 6]`.

#### Input

| Name   | Description       |
| ------ | ----------------- |
| Values | Array of numbers. |

#### Output

`Array of cumulative values.`

#### Example

```
Input:
{
  "fares": [10000, 20000, 50000]
}

Standard Transformer Config:
variables:
- name: cumulative_fares
  expression: CumulativeValue($.fares)

Output: `[10000, 30000, 80000]`
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
Input:
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
Input:
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


## Series Expression
Series expression is function that can be invoked by series (column) values in a table

### Get
`Get` will retrieve a row in series based on the given index

#### Input
| Name | Description |
|------|-------------|
| Index| Position of rows starts with 0 |

#### Output
Single series row

#### Examples
Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Users try to retrieve index 2 for series `avg_order_1_day`

Standard Transformer Config:
```
variables:
- name: total_order_1_day
  expression: yourTableName.Col("avg_order_1_day").Get(2)
```

Output: 4000

### IsIn
`IsIn` checks whether value in a row is part of the given array, the result will be a new series that has boolean type

#### Input
| Name | Description |
|------|-------------|
| Comparator| Array of value  |

#### Output
New Series that has boolean type and same dimension with original series

#### Examples
Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Standard Transformer Config:
```
variables:
- name: bool_series
  expression: yourTableName.Col("avg_order_1_day").IsIn([2000, 3000])
```
Output:
| bool_series |
|----------------|
| true |
| true |
| false |

### StdDev
`StdDev` is a function to calculate standard deviation from series values. The output will be single value

#### Input
No Input

#### Output
Single value with float type

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |


Standard Transformer Config:
```
variables:
- name: std_dev
  expression: yourTableName.Col("avg_cancellation_rate_30_day").StdDev()
```
Output: 0.0068475461947247

### Mean
`Mean` is a function to calculate mean value from series values. The output will be single value

#### Input
No Input

#### Output
Single value with float type

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Standard Transformer Config:
```
variables:
- name: mean
  expression: yourTableName.Col("avg_order_1_day").Mean()
```
Output: 3000

### Median

`Median` is a function to calculate median value from series values. The output will be single value

#### Input
No Input

#### Output
Single value with float type

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Standard Transformer Config:
```
variables:
- name: median
  expression: yourTableName.Col("avg_order_1_day").Median()
```
Output: 3000

### Max

`Max` is a function to find max value from series values. The output will be single value

#### Input
No Input

#### Output
Single value with float type

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Standard Transformer Config:
```
variables:
- name: max
  expression: yourTableName.Col("avg_order_1_day").Max()
```
Output: 4000

### MaxStr

`MaxStr` is a function to find max value from series values. The output will be single value in string type

#### Input
No Input

#### Output
Single value with string type

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Standard Transformer Config:
```
variables:
- name: max_str
  expression: yourTableName.Col("avg_order_1_day").MaxStr()
```
Output: "4000"

### Min

`Min` is a function to find minimum value from series values. The output will be single value in float type

#### Input
No Input

#### Output
Single value with float type

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Standard Transformer Config:
```
variables:
- name: min
  expression: yourTableName.Col("avg_order_1_day").Min()
```
Output: 2000

### MinStr

`MinStr` is a function to find minimum value from series values. The output will be single value in string type

#### Input
No Input

#### Output
Single value with string type

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Standard Transformer Config:
```
variables:
- name: min_str
  expression: yourTableName.Col("avg_order_1_day").MinStr()
```
Output: "2000"

### Quantile

`Quantile` is a function to returns the sample of x such that x is greater than or equal to the fraction p of samples

#### Input
Fraction in float type

#### Output
Single value with float type

#### Examples

Suppose users have table `yourTableName`

| rank |
|-------|
| 1 |
| 2 | 
| 3 |
| 4 |
| 5 |
| 6 | 
| 7 |
| 8 |
| 9 |
| 10 |

Standard Transformer Config:
```
variables:
- name: quantile_0.9
  expression: yourTableName.Col("rank").Quantile(0.9)
```
Output: 9

### Sum

`Sum` is a function to sum all the values in the seriess. The output will be single value in float type

#### Input
No Input

#### Output
Single value with float type

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | avg_order_1_day | avg_cancellation_rate_30_day |
|---------------|-----------------|------------------------------|
| 1 | 2000 | 0.02 |
| 2 | 3000 | 0.005 |
| 3 | 4000 | 0.006 |

Standard Transformer Config:
```
variables:
- name: sum
  expression: yourTableName.Col("avg_order_1_day").Sum()
```
Output: 9000

### Flatten
`Flatten` is a function to flatten all values in a series, this is suitable for series that has list type, for non list the result will be the same with the original seriess

#### Input
No Input

#### Output
New Series that the value already flatten

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | nearby_restaurant_ids |
|---------------|-----------------|
| 1 | [2, 3, 4] |
| 2 | [4, 5, 6] |
| 3 | [7, 8, 9] |

Standard Transformer Config:
```
variables:
- name: restaurant_ids
  expression: yourTableName.Col("nearby_restaurant_ids").Flatten()
```
Output:
| restaurant_ids |
|----------------|
| 2 |
| 3 |
| 4 |
| 4 |
| 5 |
| 6 |
| 7 |
| 8 |
| 9 |

### Unique
`Unique` is a function to return all values without duplication.
#### Input
No Input

#### Output
New Series that has unique value for each row

#### Examples

Suppose users have table `yourTableName`

| restaurant_id | rating |
|---------------|-----------------|
| 1 | [2, 2, 4] |
| 2 | [4, 5, 4] |
| 1 | [2, 2, 4] |

Standard Transformer Config:
```
variables:
- name: unique_restaurant_id
  expression: yourTableName.Col("restaurant_id").Unique()
```
Output:
| unique_restaurant_id |
|----------------|
| 1 |
| 2 |

```
variables:
- name: rating
  expression: yourTableName.Col("rating").Unique()
```
Output:
| rating |
|--------|
| [2, 2, 4] |
| [4, 5, 4] |