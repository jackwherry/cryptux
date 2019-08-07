# Cryptux
E2E encrypted group chat in your terminal

In fewer than 200 lines of code, Cryptux implements a simple protocol for sending and receiving encrypted messages, using SecretBox from NaCl. A rogue server could record encrypted messages (and infer the lengths of the plaintext this way), the times and IP addresses of users, and the volume that each room receives, but message contents are completely obstructed from the server's view. People chatting in Cryptux are not aware of each other's IP or any identifying information unless their server has been compromised. 

You must have agreed on a password with the people you're chatting with over a secure channel outside of the application, because Cryptux does not implement public-key cryptography. 

**Example intended use case**: A couple of friends would like to exchange short pieces of important information over an encrypted channel. They have met in person and decided on a scheme for determining a password (such as a printed list of daily passwords). After making their own channel/room on a public server, they send each other their messages. They realize that a third party needs to be included and share the password sheet with that person. Now, all of them can decrypt each others' messages, but no one else can (assuming the passwords have been kept secret). 

## Installation
Assuming you have a recent version of Go installed, simply run the following:
```shell
$ go get github.com/jack-the-coder/cryptux/client
```
Then, simply run the executable at `$GOPATH/bin/client` like this:

```shell
$ ./client -id=room -pass=password
```
It defaults to `v.snazz.xyz` as its server, although you can specify `-server=whatever.com` on the command line. 

To set up your very own server, do so as follows:
```shell
$ go get github.com/jack-the-coder/cryptux/server
$ cd $GOPATH/bin
$ ./server
```
It runs on port 8000, which is hardcoded in both the server and the client. You'll want to run `./server` as a daemon in your init system or use `nohup` for temporary installations. 

## Security model
Needless to say, Cryptux does not follow a "standard" approach for E2E chat. In the name of simplicity and ease of use, a couple of security properties have been lost. For the intended application, I don't think that the way Cryptux handles its crypto is problematic.  

Cryptux generates a secret key from the user's password using Argon2id with a *hardcoded salt*. As I understand it, the benefits of using salts in password-hashing applications are to make it exceedingly difficult for an attacker to apply pre-computed rainbow tables and to prevent the attacker from being able to crack multiple passwords at once. In this application, however, the attacker's goal is not to figure out the password given the derived encryption key. If the attacker already has the encryption key, it's game over. My research has suggested so far that using a different salt per channel would not be of any use. If this is wrong, please open an issue or submit a pull request. 

**EDIT**: It turns out that using a hardcoded salt could theoretically allow an attacker to brute-force multiple channels at the same time. This isn't a useful property, in my opinion, so for simplicity's sake we're keeping the single-salt method. It has also come to my attention that the method Cryptux uses to generate keys should theoretically only be used for authentication, and the actual key should be determined using public-key encryption. The trouble with that is that the ability to add anyone to the communication instantly by giving them the password is broken. 

Once a 32-byte key has been derived from the password, Cryptux generates a 24-byte nonce and encrypts the user's message. The encrypted message is appended to the nonce, so that the first 24 bytes are the nonce and the rest is the ciphertext. The potential danger here is that there is no message padding: a rogue server or network operator could determine the length of the communication. 

As a result of the way that Cryptux works, multiple groups of users can be chatting in the same channel/room with different passwords. In this case, when the other group posts a message, an error is displayed instead of the message, since the client does not have the key for that communication. 

## Performance considerations
For better connections on spotty networks and to simplify the code, the client polls the server for the most recent message ten times a second. This could put a fair amount of load on the server, given that the number of connections per second is ten times the number of users accessing the server. However, there is no database or other I/O on the server and the CPU and memory requirements are pretty reasonable. I am expecting the biggest bottleneck to be networking on servers with slow connections and lots of users. 

Ping times between the clients and servers should be below 50 milliseconds or so to prevent message loss on high-traffic channels. 