package spserv

import "github.com/MaKcm14/vk-test/internal/controller/spserv/sprpc"

// SPServer defines the common interface for every Sub/Pub server implementation.
type SPServer interface {
	sprpc.PubSubServer
	Close()
}
