package circuit_breaker

import (
	"context"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/social_graph"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
	"tweet/app_errors"
	"tweet/model"
	"tweet/tls"
)

type SocialGraphCircuitBreaker struct {
	circuitBreaker *gobreaker.CircuitBreaker
	tracer         trace.Tracer
}

func NewSocialGraphCircuitBreaker(tracer trace.Tracer) *SocialGraphCircuitBreaker {
	return &SocialGraphCircuitBreaker{
		circuitBreaker: CircuitBreaker(),
		tracer:         tracer,
	}
}

func CircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "SocialGraphCircuitBreaker",
			MaxRequests: 1,
			Timeout:     5 * time.Second,
			Interval:    0,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 0
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				log.Printf("Circuit Breaker '%s' changed from '%s' to '%s'\n", name, from, to)
			},
		},
	)
}

func (cb *SocialGraphCircuitBreaker) CheckVisibility(ctx context.Context, targetUser *social_graph.SocialGraphUsername) (bool, *app_errors.AppError) {
	cbCtx, span := cb.tracer.Start(ctx, "SocialGraphCircuitBreaker.CheckVisibility")
	defer span.End()

	conn, err := getgRPCConnection("social-graph:9001")
	defer conn.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	authUser := cbCtx.Value("authUser").(model.AuthUser)

	socialGraphService := social_graph.NewSocialGraphServiceClient(conn)
	cbCtx = metadata.AppendToOutgoingContext(cbCtx, "authUsername", authUser.Username)

	execute, err := cb.circuitBreaker.Execute(func() (interface{}, error) {
		response, err := socialGraphService.CheckVisibility(cbCtx, targetUser)

		if err != nil {
			return false, &app_errors.AppError{Code: 500, Message: err.Error()}
		}

		return response.Visibility, nil
	})
	if err != nil {
		return false, &app_errors.AppError{Code: 503, Message: err.Error()}
	}

	return execute.(bool), nil
}

func getgRPCConnection(address string) (*grpc.ClientConn, error) {
	creds := credentials.NewTLS(tls.GetgRPCClientTLSConfig())

	conn, err := grpc.DialContext(
		context.Background(),
		address,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)

	if err != nil {
		log.Fatalf("Failed to start gRPC connection: %v", err)
	}

	return conn, err
}
