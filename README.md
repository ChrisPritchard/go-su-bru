# go-su-bru

Go based su password bruteforcer - uses goroutines to start multiple pseudoterminals, then just run password candidates against su in each terminal over and over.

Unix only, and currently has a sole dependency on `golang.org/x/sys/unix`

DONT USE FOR ILLEGAL PURPOSES - if only because this is not exactly stealthy.

## Usage

Directly invoking the script:

`go run gosubru.go testuser ./wordlist.txt`

Additional options: if you specify a third param you can set the number of agents at once (the default is 256).

If you specify a *fourth* param you can specify the timeout limit for each su instance (default is 5 seconds):

`args: username passwordfile [batchsize - defaults 256] [command timeout - defaults 5 (seconds)]`

## Building

nothing special about compiling this, just make sure you have the unix lib installed:

```bash
go get -u golang.org/x/sys/unix
go build -o gosubru gosubru.go
```
