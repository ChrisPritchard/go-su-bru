# go-su-bru

Go based su password bruteforcer

Still a work in progress - various aspects could be improved, but seems to work well

Unix only, and currently has a sole dependency on `golang.org/x/sys/unix`

## Testing

Create a new user, and give them a password from adanalist.txt (this is originally from a boot2root). For example, using a password at about position 47k:

```
sudo useradd hakanbey
echo -e "123adanacurly\n123adanacurly" | sudo passwd hakanbey
```

then to run:

`go run subrute.go hakanbey ./adanalist.txt`

