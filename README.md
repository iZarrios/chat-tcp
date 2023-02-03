# Simpe Chat server using Transmission Control Protocol (TCP)

## Types of Commands

There are 5 commands that the user can do, which is typically used with the following pattern.

```bash
/cmd ...args
```

1. *`/nick`*  in which the user can change his nickname (which is default to be 'anon').
1. *`/join`*  in which the user can join a room (__this laso creates a room__).
1. *`/rooms`*  lists the currently availalbe rooms.
1. *`/msg`* the user can send a message in the currently joined room.
1. *`/quit`* the user quits the chat application.

## How to use

* note that 3000 is the port of my choice and can be changed.

using telnet

```bash
telnet localhost 3000
```

using netcat

```bash
nc localhost 3000
```
