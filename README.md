# seeChange

A mutex-free FIFO string buffer served over ZeroMQ, with Protobuf messages and a Go client that demonstrates it.

## Running it

### With Docker Compose

```
docker compose up --build
```

This builds both images, starts the server, then runs the client's demo against it. The client prints each request next to the server's reply and exits:

```
push:{value:"A"} -> push:{ok:true}
push:{value:"B"} -> push:{ok:true}
push:{value:"C"} -> push:{ok:true}
size:{}          -> size:{size:3}
pull:{}          -> pull:{value:"A" ok:true}
size:{}          -> size:{size:2}
pull:{}          -> pull:{value:"B" ok:true}
pull:{}          -> pull:{value:"C" ok:true}
pull:{}          -> pull:{}
```

The server keeps running after the client exits. Stop everything with Ctrl+C, or `docker compose down`. To stop the server as soon as the demo finishes, run `docker compose up --build --abort-on-container-exit` instead.

The last line is a pull from an empty buffer. Proto3 doesn't print zero values, so `value:""` and `ok:false` aren't shown.

### Without Docker

Requires Go 1.26. In one terminal:

```
go run ./cmd/server
```

In another:

```
go run ./cmd/client
```

The client dials `tcp://localhost:5555` by default. Set `SERVER_ADDR` to point it somewhere else.

### Tests

```
go test ./...
```

### Regenerating the Protobuf code

The generated code in `internal/pb` is committed, so you only need this after editing `proto/buffer.proto`. It requires `protoc` and `protoc-gen-go`:

```
protoc --go_out=. --go_opt=module=github.com/jkqp/seechange proto/buffer.proto
```

## Layout

```
cmd/server/       server binary: socket setup and the request loop
cmd/client/       client binary: the demo sequence
internal/buffer/  the channel-based FIFO buffer
internal/server/  Handle: turns a Request into a buffer call and a Response
internal/pb/      generated Protobuf code, shared by both binaries
proto/            the .proto source
```

The `main` packages are kept thin. The logic lives under `internal/`, where it can be unit tested without a socket.

## Design

### Buffer

The brief required channels instead of a mutex, so the decision was how to use them. I considered two options:

1. **Buffered channel as storage**: the channel itself holds the items.
2. **Owner goroutine**: one goroutine keeps the items in a slice, and other goroutines send it requests over channels.

I went with the buffered channel because it's less code and simpler. There's no extra goroutine to start or shut down, and `Size` is just `len(ch)`. Channels send values in the order they were added, so we get FIFO ordering for free. The channel has to be buffered, because an unbuffered channel can't store anything. It only hands data from one goroutine to another.

The trade-off is a fixed capacity. When the buffer is full, `Push` returns `false` instead of waiting, because a full channel can't accept any more items. The owner version stores items in a slice that grows, so it never fills up. The brief doesn't need that, so I chose the simpler design.

### Messages

`Request` and `Response` each wrap a `oneof`, with one case per operation. Every message on the socket has the same type, and the receiver switches on which case is set. The alternative was a separate message type per operation, plus some kind of header saying which one follows. A `oneof` gets the same result with no extra framing.

`Response` also has an `ErrorResponse` case. The server sends it for an empty request or for bytes that don't parse, so the client always gets a reply it can decode.

### ZeroMQ

**Socket pattern: REQ/REP.** Every operation is one request followed by one reply, which is exactly what REQ/REP enforces. The constraint is that a REP socket must reply before it can receive again. That's why the server answers malformed input with an `ErrorResponse` instead of returning an error, which would kill the server for every client.

**Library: `go-zeromq/zmq4`.** It's a pure-Go implementation, so there's no cgo and no libzmq to install. The binaries build with `CGO_ENABLED=0` and run on a minimal image. The trade-off is that it's less battle-tested than `pebbe/zmq4`, which wraps the C library. For this exercise, a simpler build was worth more.

**Concurrency.** The server handles one request at a time in a single loop, so the buffer isn't currently accessed concurrently over the network. It's still safe for concurrent use (see `TestConcurrentPush`). If the server moved to a ROUTER/DEALER pattern with several workers, the buffer wouldn't need to change.

### Code sharing

Both binaries import `internal/pb`, so the client and server use the same message types, generated from one `.proto` file. The packages live under `internal/` so they can't be imported from outside this module. If another project needed the buffer, it would move out of `internal/` at that point.

### Docker

One multi-stage `Dockerfile` builds both binaries, and the `APP` build argument picks which one. The two images differ only in the binary they contain, so two near-identical Dockerfiles would just drift apart. The build stage downloads modules before copying the source, so code changes don't invalidate the dependency layer. The run stage is `distroless/static` running as non-root: it has no shell and no package manager, and each image is about 17 MB.

In compose, the client reaches the server by its service name (`tcp://server:5555`), passed in through `SERVER_ADDR`. `depends_on` only orders container start-up. It doesn't wait for the server to be listening, so the client socket also retries its dial. The server handles `SIGTERM` and shuts down cleanly when compose stops it.

## Limitations

- The buffer capacity (10) and the server's listen address are hardcoded.
- One server goroutine handles every request in order. That's fine for this load, but it is a throughput ceiling.
- The client is a fixed demo, not an interactive tool.

## Conclusion 
I used Claude Code throughout this task, and how much I let it lead depended on the part.

For the buffer, I was comfortable letting Claude take the lead. The functions were straightforward, and not something I felt I needed to be coached through. I took the same approach for the buffer's unit tests, except for the concurrency test, where I was more hands-on. For that, I asked Claude to coach me with hints and questions rather than write the code, so I'd understand it myself. From the brief, and from my research into SeeChange's business, I inferred that concurrency is a core part of how they use Go.

I made more use of coaching for the handler, server and client, because I wasn't familiar with ZeroMQ or Protobuf before. Once I saw examples of data being unmarshalled, it clicked that this was the same thing I do with JSON in my day job, and the handler came together. The main thing I had to get my head around was the nested wrapper types Protobuf generates for a `oneof`, which is how the handler builds its responses. The server and client were a similar story. Once I'd seen a few examples and realised this was a request/reply exchange over TCP, I could skeleton them out myself.

Claude wrote the Dockerfile and compose file. I reviewed them against what I know from working with Kubernetes manifests on GCP, mostly adjusting CPU and replica counts. The compose file plays the same role locally that a Deployment and Service would on GKE. The client reaching the server by its service name works like Kubernetes Service DNS, and the `SIGTERM` handling matters for the same reason it does when a pod is shut down. The Dockerfile itself was the newer part for me, since at work a cloud platform team owns the build and deployment, and images are only ever built by CI pipelines. Multi-stage builds and distroless base images were the main things I had to read up on.