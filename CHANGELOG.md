# Merlin Release Notes

## v0.20.0 - 2022-04-18
* Upgrade pyyaml to >=5.4 by @karzuo in [#230](https://github.com/gojek/merlin/pull/230)
* Fix mapping of entity names to bigtable by @khorshuheng in [#223](https://github.com/gojek/merlin/pull/223)
* Column wise operator, filter and slice row by @tiopramayudi in [#228](https://github.com/gojek/merlin/pull/228)
* Add geohash related expression by @tiopramayudi in [#235](https://github.com/gojek/merlin/pull/235)
* Create tables from file & UI bugs fix by @karzuo in [#234](https://github.com/gojek/merlin/pull/234)
* Fix insert cache with NaN value by @tiopramayudi in [#236](https://github.com/gojek/merlin/pull/236)
* fix bug with loading table from file by @karzuo in [#238](https://github.com/gojek/merlin/pull/238)
* Fix slice row ui when delete entry in start or end index by @tiopramayudi in [#239](https://github.com/gojek/merlin/pull/239)
* Fix issue with built-in function by @tiopramayudi in [#241](https://github.com/gojek/merlin/pull/241)

## v0.19.0 - 2022-03-07
* Truncated bigtable name by @tiopramayudi in [#209](https://github.com/gojek/merlin/pull/209)
* Emit metric for redis and bigtable client by @tiopramayudi in [#210](https://github.com/gojek/merlin/pull/210)
* Use pod as label for cAdvisor container metrics by @ariefrahmansyah in [#213](https://github.com/gojek/merlin/pull/213)
* Use container label to filter out POD container by @ariefrahmansyah in [#214](https://github.com/gojek/merlin/pull/214)
* Cyclical Encoder + Column Ordering fix by @karzuo in [#208](https://github.com/gojek/merlin/pull/208)
* Fixed column update Algo and tests by @karzuo in [#216](https://github.com/gojek/merlin/pull/216)
* Fix int test by @karzuo in [#218](https://github.com/gojek/merlin/pull/218)
* Fix integration test by @karzuo in [#220](https://github.com/gojek/merlin/pull/220)
* Introducing list types for feast features retrieval by @ariefrahmansyah in [#212](https://github.com/gojek/merlin/pull/212)
* Use golang 1.17 by @pradithya in [#222](https://github.com/gojek/merlin/pull/222)
* Do not set resource annotation when queueResourcePercentage is empty by @tiopramayudi in [#224](https://github.com/gojek/merlin/pull/224)
* Initialize inference service's annotations by @ariefrahmansyah in [#225](https://github.com/gojek/merlin/pull/225)
* Support convertion from []interface to list type. by @ariefrahmansyah in [#226](https://github.com/gojek/merlin/pull/226)

## v0.18.0 - 2022-01-13
* Add time related built in function and cumulative value function by @tiopramayudi in [#197](https://github.com/gojek/merlin/pull/197)
* Load location using evaluated args value by @ariefrahmansyah in [#201](https://github.com/gojek/merlin/pull/201)
* Add Go's zoneinfo to standard transformer docker image by @ariefrahmansyah in [#202](https://github.com/gojek/merlin/pull/202)
* Add default jsonpath capability by @tiopramayudi in [#203](https://github.com/gojek/merlin/pull/203)
* Fix yaml specification on feast standard transformer page by @tiopramayudi in [#205](https://github.com/gojek/merlin/pull/205)
* Update google-cloud-sdk version by @pradithya in [#206](https://github.com/gojek/merlin/pull/206)
* Standard Transformer serving directly from Bigtable by @tiopramayudi in [#204](https://github.com/gojek/merlin/pull/204)
* Add unique identifier for euicheckbox by @tiopramayudi in [#207](https://github.com/gojek/merlin/pull/207)