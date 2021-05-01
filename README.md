# go-su-bru

Go based su password bruteforcer

Still a work in progress - various aspects could be improved, but seems to work well

Unix only, and currently has a sole dependency on `golang.org/x/sys/unix`

DONT USE FOR ILLEGAL PURPOSES - if only because this is not exactly stealthy.

## Testing

Create a new user, and give them a password from wordlist.txt (this is originally from a boot2root). For example, using a password at about position 47k:

```
sudo useradd testuser
echo -e "motown\nmotown" | sudo passwd testuser
```

then to run (directly invoking the script):

`go run gosubru.go testuser ./wordlist.txt`

Additional options: if you specify a third param you can set the number of agents at once (the default is 256).

If you specify a *fourth* param you can specify the timeout limit for each su instance (default is 5 seconds):

`args: username passwordfile [batchsize - defaults 256] [command timeout - defaults 5 (seconds)]`

## building

nothing special about compiling this, just make sure you have the unix lib installed:

```bash
go get -u golang.org/x/sys/unix
go build -o gosubru gosubru.go
```
