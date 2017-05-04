go-cimd
=======

[CIMD](https://en.wikipedia.org/wiki/CIMD) server simulator 

Supported operations

- Login
- Logout
- Alive
- Submit message
- Deliver status report


### config
```
cimd_user: cimd_username
cimd_pw: cimd_password
port: 16001
greeting: Welcome to go-cimd
delivery_delay: 5
```

### build
```
$ make all
```
### run
```
$ ./go-cimd
```

### test with [ecimd2](https://github.com/VoyagerInnovations/ecimd2)
![image](http://g.recordit.co/27NJAT6gIC.gif)