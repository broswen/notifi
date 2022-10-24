# Notifi

![diagram.png](diagram.png)



### todo
- [x] custom errors 
- [ ] api response types, entity -> response
- [x] poll postgres for scheduled notifications and submit to delivery queue
  - [ ] figure out how to partition/shard db polling for scaling (random number column with consistent hashing?)
- [ ] prevent spam when db is down and notification is sent successfully
- [ ] add api service tokens, count messages per token in metrics
  - [ ] do this through cloudflare access service tokens
- [x] add prometheus metrics
- [x] set up k8s
- [ ] add tests