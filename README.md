# go-su-bru

Go based su password bruteforcer

Still a work in progress - running into some sort of ulimit or file handler issue.

Currently has a sole dependency on `github.com/creack/pty`

## Testing

Create a new user, and give them a password from adanalist.txt (this is originally from a boot2root). For example, using a password at about position 47k:

```
sudo useradd hakanbey
echo -e "123adanacurly\n123adanacurly" | sudo passwd hakanbey
```

then to run:

`go run subrute.go hakanbey ./adanalist.txt`

currently at around 23k-33k into the list, it will crash with:

`fork/exec /bin/su: resource temporarily unavailable`

or various issues of that nature.
