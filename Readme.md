# Experiment: UPD server with confirmation

[WIP!]

Trial to make UDP server with confirming of message delivery 
using Golang as server language, and Php as a client.

## Purpose

Create some kind of UDP layer based protocol:
1. No TCP handshakes
2. Guaranty of delivery
3. Easy and lightweight
4. Fast as possible

## Steps

- [x] Make php -> go communication
- [ ] Split data to packages on client side
- [ ] Join packages to data on server side
- [ ] Receive confirmation
- [ ] Timeouts
- [ ] ...

## Possible use-cases

- UDP proxy for RabbitMQ