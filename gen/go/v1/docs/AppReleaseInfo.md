# AppReleaseInfo

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ActiveChannels** | [**[]AppChannel**](AppChannel.md) | The active channels | [optional] [default to null]
**AppId** | **string** | The application ID | [optional] [default to null]
**CreatedAt** | [**time.Time**](time.Time.md) | The time at which the release was created | [optional] [default to null]
**Editable** | **bool** | If the release is editable | [optional] [default to null]
**EditedAt** | [**time.Time**](time.Time.md) | The last time at which the release was changed | [optional] [default to null]
**PreflightChecks** | [**[]PreflightCheck**](PreflightCheck.md) | Release preflight checks | [optional] [default to null]
**Sequence** | **int64** | The app sequence number | [optional] [default to null]
**Version** | **string** | The vendor supplied version | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


