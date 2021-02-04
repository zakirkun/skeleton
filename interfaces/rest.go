package interfaces

import (
	"context"
	"fmt"
	"log"
	"net/http"

	configs "github.com/crowdeco/skeleton/configs"
	handlers "github.com/crowdeco/skeleton/handlers"
	middlewares "github.com/crowdeco/skeleton/middlewares"
	routes "github.com/crowdeco/skeleton/routes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

type (
	Rest struct{}
)

func (g *Rest) Run() {
	log.Printf("Starting REST Server on :%d", configs.Env.HtppPort)

	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	endpoint := fmt.Sprintf("0.0.0.0:%d", configs.Env.RpcPort)
	conn, err := grpc.DialContext(ctx, endpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}

	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	mux := http.NewServeMux()

	router := handlers.NewRouter()
	router.Add(routes.NewMuxRouter(conn))
	router.Add(routes.NewGRpcGateway(ctx, conn))

	middleware := handlers.NewMiddleware()
	middleware.Add(middlewares.NewAuth())

	log.Println("API Documentation is ready at /api/docs/ui")

	http.ListenAndServe(fmt.Sprintf(":%d", configs.Env.HtppPort), middleware.Attach(router.Handle(mux)))
}
