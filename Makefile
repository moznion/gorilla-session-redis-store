benchmark:
	go test -bench=. -benchmem

benchmark-moznion/redistore:
	go test -benchmem -test.bench MoznionRedistore -cpuprofile moznion-redistore.prof

benchmark-rbcervilla/redisstore:
	go test -benchmem -test.bench RbcervillaRedisstore -cpuprofile rbcervilla-redisstore.prof

