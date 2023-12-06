package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mateusmatinato/goexpert-clean-arch/configs"
	"github.com/mateusmatinato/goexpert-clean-arch/internal/event/handler"
	"github.com/mateusmatinato/goexpert-clean-arch/internal/infra/graph"
	"github.com/mateusmatinato/goexpert-clean-arch/internal/infra/grpc/pb"
	"github.com/mateusmatinato/goexpert-clean-arch/internal/infra/grpc/service"
	"github.com/mateusmatinato/goexpert-clean-arch/internal/infra/web/webserver"
	"github.com/mateusmatinato/goexpert-clean-arch/pkg/events"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// sqlite3
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	execMigrations(db)

	rabbitMQChannel := getRabbitMQChannel()

	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("OrderCreated", &handler.OrderCreatedHandler{
		RabbitMQChannel: rabbitMQChannel,
	})

	createOrderUseCase := NewCreateOrderUseCase(db, eventDispatcher)
	listOrderUseCase := NewListOrderUseCase(db)

	ws := webserver.NewWebServer(cfg.WebServerPort)
	orderPath := "/order"
	webOrderHandler := NewWebOrderHandler(db, eventDispatcher)
	ws.AddHandler(webserver.NewRoute(orderPath, webserver.POST, webOrderHandler.Create))
	ws.AddHandler(webserver.NewRoute(orderPath, webserver.GET, webOrderHandler.FindAll))
	fmt.Println("Starting web server on port", cfg.WebServerPort)
	go ws.Start()

	grpcServer := grpc.NewServer()
	orderService := service.NewOrderService(*createOrderUseCase, *listOrderUseCase)
	pb.RegisterOrderServiceServer(grpcServer, orderService)
	reflection.Register(grpcServer)

	fmt.Println("Starting gRPC server on port", cfg.GRPCServerPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCServerPort))
	if err != nil {
		panic(err)
	}
	go grpcServer.Serve(lis)

	srv := graphql_handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		CreateOrderUseCase: *createOrderUseCase,
		ListOrderUseCase:   *listOrderUseCase,
	}}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	fmt.Println("Starting GraphQL server on port", cfg.GraphQLServerPort)
	err = http.ListenAndServe(":"+cfg.GraphQLServerPort, nil)
	if err != nil {
		return
	}
}

func execMigrations(db *sql.DB) {
	instance, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fSrc, err := (&file.File{}).Open("./db/migrations")
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithInstance("file", fSrc, "sqlite3", instance)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}
}

func getRabbitMQChannel() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return ch
}
