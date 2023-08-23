# Notifi

![diagram.png](diagram.png)


## API

RESTful webservice that manages notifications.


`POST /api/notifications`

Submits a new notification to be processed.

`GET /api/notifications`

Returns a paginated list of notifications and their details. Add `?deleted=true` to show deleted notifications.

Use `limit` and `offset` query params to control pagination.


`GET /api/notifications/{id}`

Returns the details for a single notification.

`DELETE /api/notifications/{id}`

Marks a notification as deleted. This prevents delivery of scheduled notifications.

## Router

A simple redirection service that processes messages from the Kafka topic and routes them to the correct destination.

For notifications without a schedule, it submits them to the delivery queue to be delivered.

For notification with a schedule, it stores them in the outbox table to be polled and delivered at a later time.

## Poller

Polls the notifications for events with a schedule that haven't been deleted or delivered and are due within the next polling period.

Then it submits them to the delivery queue.

## Delivery

Receives notifications from the delivery queue and delivers them to the configured destinations.

### todo
- [x] custom errors 
- [ ] api response types, entity -> response
- [ ] bulk notifications, same content, many destinations?
- [ ] update repositories to accept a connection instead of pool
- [x] poll postgres for scheduled notifications and submit to delivery queue
  - [x] assign notifications to random partition key, assign partition range to each poller replica
  - [ ] use `select for update skip locked`
- [x] prevent spam when db is down and notification is sent successfully
- [x] add api service tokens
  - [ ] count messages via service client id and prometheus
  - [x] do this through cloudflare access service tokens
- [x] add prometheus metrics
- [x] set up k8s
- [x] add tests
- [ ] schemas for all events in protobuf
- [x] protect against duplicate delivery by storing a list of delivered notification ids in redis
  - set TTL on ids to store only for a certain period (1~ hour)
