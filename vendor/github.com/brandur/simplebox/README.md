# simplebox

 [![build status](https://secure.travis-ci.org/brandur/simplebox.png)](https://travis-ci.org/brandur/simplebox) [![GoDoc](https://godoc.org/github.com/brandur/simplebox?status.png)](https://godoc.org/github.com/brandur/simplebox)

Package simplebox provides a simple, easy-to-use cryptographic API where all of
the hard decisions have been made for you in advance. The backing cryptography
is XSalsa20 and Poly1305, which are known to be secure and fast.

This package is a Golang port of the [RbNaCl module of the same name][rbnacl].

## Installation and Usage

```
go get github.com/brandur/simplebox
```

Please see [godoc for usage information and examples][godoc].

[godoc]: https://godoc.org/github.com/brandur/simplebox
[rbnacl]: https://github.com/cryptosphere/rbnacl/wiki/SimpleBox
