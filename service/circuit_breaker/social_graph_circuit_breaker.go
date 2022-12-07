package circuit_breaker

import (
	"context"
	"github.com/FTN-TwitterClone/grpc-stubs/proto/social_graph"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
			Name:        "SocialGraph",
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

	conn, err := tls.GetgRPCConnection("social-graph:9001")
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
		span.SetStatus(codes.Error, err.Error())
		return false, &app_errors.AppError{Code: 503, Message: err.Error()}
	}

	return execute.(bool), nil
}

func (cb *SocialGraphCircuitBreaker) GetMyFollowers(ctx context.Context) ([]*social_graph.SocialGraphUsername, *app_errors.AppError) {
	cbCtx, span := cb.tracer.Start(ctx, "SocialGraphCircuitBreaker.GetMyFollowers")
	defer span.End()

	conn, err := tls.GetgRPCConnection("social-graph:9001")
	defer conn.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	authUser := cbCtx.Value("authUser").(model.AuthUser)

	socialGraphService := social_graph.NewSocialGraphServiceClient(conn)
	cbCtx = metadata.AppendToOutgoingContext(cbCtx, "authUsername", authUser.Username)

	execute, err := cb.circuitBreaker.Execute(func() (interface{}, error) {
		response, err := socialGraphService.GetMyFollowers(cbCtx, new(empty.Empty))

		if err != nil {
			return false, &app_errors.AppError{Code: 500, Message: err.Error()}
		}

		return response.Usernames, nil
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 503, Message: err.Error()}
	}

	return execute.([]*social_graph.SocialGraphUsername), nil
}

func (cb *SocialGraphCircuitBreaker) GetTargetGroupUsers(ctx context.Context, targetGroup model.TargetGroup) ([]*social_graph.SocialGraphUsername, *app_errors.AppError) {
	cbCtx, span := cb.tracer.Start(ctx, "SocialGraphCircuitBreaker.GetTargetGroupUsers")
	defer span.End()

	conn, err := tls.GetgRPCConnection("social-graph:9001")
	defer conn.Close()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 500, Message: err.Error()}
	}

	socialGraphService := social_graph.NewSocialGraphServiceClient(conn)

	tg := social_graph.SocialGraphTargetUsersGroup{
		Town:   targetGroup.Town,
		Gender: targetGroup.Gender,
		MinAge: targetGroup.MinAge,
		MaxAge: targetGroup.MaxAge,
	}

	execute, err := cb.circuitBreaker.Execute(func() (interface{}, error) {
		response, err := socialGraphService.GetTargetGroupUser(cbCtx, &tg)

		if err != nil {
			return false, &app_errors.AppError{Code: 500, Message: err.Error()}
		}

		return response.Usernames, nil
	})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, &app_errors.AppError{Code: 503, Message: err.Error()}
	}

	return execute.([]*social_graph.SocialGraphUsername), nil
}
