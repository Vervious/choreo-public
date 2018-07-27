README

This folder contains simulations of Dolevs, based on/copied from the other libraries in this repo.

It hooks a WebGL visualization onto these simulations, and also removes the grpc component (since we are running experiments locally). We may be able to port this to visualizing deployed clusters when needed.

## To Run:
```
cd /choreo-public/sim/web
python -m SimpleHTTPServer

go get github.com/Vervious/eventsource
cd /choreo-public/sim
go run *.go
```


## Simulation

Ignore the below, we use HTTP2/SSE for simplicity. The code supporting HTTP2/SSE is by no means production ready; they consist of a fork of a random library (github/Vervious/eventsource) and hacky client-side javascript code. This may require revisitation.

The visualization client uses grpc/grpc-web server-side streams to connect to the simulator server (or, perhaps, to a node in a deployed cluster. We could gossip network wide status updates for the visualizer to use.) We explored both vanilla websockets and HTTP2/SSE; both are viable lower-level transportation alternatives. We chose grpc-web for its growing community, adoption, and explicitness, though it isn't immediately clear how relevant its streaming capabilities are (and how they relate to HTTP2 server-push).


## Dependencies:

NOT go get google.golang.org/grpc
go get github.com/Vervious/eventsource


## Future

I would like to transpile go into javascript (gopherjs)
