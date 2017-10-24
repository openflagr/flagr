/*
Package simplebox provides a simple, easy-to-use cryptographic API where all of
the hard decisions have been made for you in advance. The backing cryptography
is XSalsa20 and Poly1305, which are known to be secure and fast.

This package uses NaCl's secretbox under the hood, but also includes a simple
yet secure nonce generation strategy. A 24-byte random nonce is generated from
a secure source, used to encrypt a message, and prepended to the resulting
ciphertex. When it's time for decryption, the message is split back into nonce
and ciphertext, and the message is decrypted.

Thanks to the size of the nonce, the chance of a collision is negligible. For
example, after encrypting 2^64 messages, the odds of there having been a
repeated nonce is approximately 2^-64.

Note that although this strategy assures the confidentiality of your messages,
it doesn't provide any protection against messages being reordered and replayed
by an active adversary.

This idea is entirely based on the SimpleBox implementation included with
RbNaCl: https://github.com/cryptosphere/rbnacl/wiki/SimpleBox
*/
package simplebox
