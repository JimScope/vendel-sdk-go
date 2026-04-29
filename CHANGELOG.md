# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-04-29

### Added
- `Client.ListDevices(ctx, *ListDevicesOptions)` — list registered gateway devices with optional `page`, `per_page`, and `device_type` filters.
- `Client.ListMessages(ctx, *ListMessagesOptions)` — list SMS messages with optional `page`, `per_page`, `status`, `device_id`, `batch_id`, `recipient`, `from`, and `to` (ISO8601) filters.
- `Device` type and `PaginatedDevices` / `PaginatedMessages` paginated response aliases.
- `ListDevicesOptions` and `ListMessagesOptions` with pointer fields so callers can distinguish unset from zero values.
- New fields on `MessageStatus`: `FromNumber` (`from_number`), `Body` (`body`), `MessageType` (`message_type`), `SentAt` (`sent_at`), and `DeliveredAt` (`delivered_at`).

## [0.3.0] - Previous

### Added
- `Client.ListContacts` and `Client.ListContactGroups`.
- `Client.GetMessageStatus` and `Client.GetBatchStatus`.
- `Contact`, `ContactGroup`, `MessageStatus`, `BatchStatus`, and generic `PaginatedResponse[T]` types.

## [0.2.0] - Previous

### Added
- `Client.SendSMSTemplate` for template-based SMS sending with variable interpolation.
- `GroupIDs` field on `SendSMSRequest` and `SendSMSTemplateRequest` for sending to contact groups.
- Initial test suite and CI workflow.

## [0.1.0] - Initial release

### Added
- `Client` with `SendSMS` and `GetQuota`.
- `VendelError` and `QuotaError` typed errors with `IsAPIError` / `IsQuotaError` helpers.
- Webhook signature verification helpers.

[0.4.0]: https://github.com/JimScope/vendel-sdk-go/releases/tag/v0.4.0
[0.3.0]: https://github.com/JimScope/vendel-sdk-go/releases/tag/v0.3.0
[0.2.0]: https://github.com/JimScope/vendel-sdk-go/releases/tag/v0.2.0
[0.1.0]: https://github.com/JimScope/vendel-sdk-go/releases/tag/v0.1.0
