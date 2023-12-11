## Getting Started

To get started with this project, clone the repository and navigate into the project directory. Then, run this command:

```sh
go mod tidy
```

### config.yml

The `config.yml` file is used to configure the database connection. Here's an example:

```yml
database:
  addresses:
    - "private-tidb.clusters.tidb-cloud.com:4000
  username: "username"
  password: "password"
  dbName: "test"
  maxOpenConns: 1000
  maxIdleConns: 500
  connMaxIdleTime: 15m
  connMaxLifetime: 30m
  tls: false
```

### Database Migration

To run the migration files located in `db/migration`, use the following command:

```sh
go run main.go migrate
```

### Seeding the Database

To seed the database with accounts, use the following command:

```sh
go run main.go seed -n 100 -l debug -d 5m -v 1 -b 1000
```

Here's what each flag means:

* `-n 100`: Create 100 accounts.
* `-l debug`: Set the log level to debug (default is info).
* `-d 5m`: Set the duration to 5 minutes (this is essentially a timeout; default is 1 minute).
* `-v 1`: Number of virtual users. For example, `-v 3` means 3 virtual users will be creating accounts concurrently. If `-n 30` and `-v 3`, each virtual user will create 10 accounts.
* `-b 1000`: The balance for each created account.

### Load Testing

To perform a load test, use the following command:

```sh
go run main.go mem -v 1000 -l info -d 1m -s 10 -t 250ms -o 100 -q 10
```

Here's what each flag means:

* `-v 1000`: Number of virtual users. Unlike the seed command, each virtual user will create only 1 transfer at a time. 1000 virtual users will create 1000 transfers concurrently, then wait until each transfer is done before starting the next one.
* `-l info`: Set the log level to info (default is info).
* `-d 1m`: Set the duration to 1 minute (this is essentially a load test duration; default is 1 minute).
* `-s 10`: Size of each queue is 10 (maximum bulk insert is 10). If the queue is full, it will start a bulk insert to the database asynchronously.
* `-t 250ms`: Timeout for each queue is 250ms. If the queue is not full but the timeout is reached, it will start a bulk insert to the database asynchronously.
* `-o 100`: Number of account IDs that will be fetched from the database to use for transfers. For example, `-o 100` means all IDs used in transfers will come from these 100 accounts.
* `-q 10`: Number of queues in the pool.

the result will look like this:

```sh
13:38:12        info    load/load.go:107        Result: 1 collision, min: 96 ms, avg: 791 ms, max: 9815 ms
13:38:12        info    load/load.go:109        Total transfer: 105516
13:38:12        info    load/load.go:110        Transfer per second: 10551
```

min, avg, max is a end to end latency of transfer (not database latency).
Total transfer is a number of transfer that has been done.
Transfer per second is a number of transfer that has been done per second.
# bench-tidb-simplified
