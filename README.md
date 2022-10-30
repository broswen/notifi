# Notifi

![diagram.png](diagram.png)



### todo
- [x] custom errors 
- [ ] api response types, entity -> response
- [ ] bulk notifications, same content, many destinations?
- [x] poll postgres for scheduled notifications and submit to delivery queue
  - [ ] figure out how to partition/shard db polling for scaling (random number column with consistent hashing?)
- [x] prevent spam when db is down and notification is sent successfully
- [x] add api service tokens
  - [ ] count messages via service client id and prometheus
  - [x] do this through cloudflare access service tokens
- [x] add prometheus metrics
- [x] set up k8s
- [x] add tests
