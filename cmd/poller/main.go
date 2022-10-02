package main

func main() {

	// parse queue topic
	// parse brokers
	// parse dsn
	// parse poll interval

	// for every interval
	// scan db for notifications where the scheduled time is less than X minutes in the future
	// submit to queue
	// remove from db (or mark sent for a cleanup job)
	// store successful notifications in postgres
	//		store failed notification status
}
