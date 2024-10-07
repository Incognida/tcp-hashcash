# world-of-wisdom

A simple TCP-server that uses [hashcash](https://en.wikipedia.org/wiki/Hashcash) algorithm to prevent DDOS 
in a challenge-response fashion.

The algorithm is following:
1. Client sends a request to the server.
2. Server sends a challenge to the client (which is hashcash).
3. Client tries to find a hash with 20 leading zeroed bits of the 160-bit hash of the challenge by constantly incrementing the counter (proof of work).
4. When the client finds the hash, it sends it back to the server.
5. Server checks if the hash has 20 zeros leading bits and if it does, it sends the response to the client.