package internal

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	api "dedb"
)

/*
 Base API level gRPC service
*/
type Service struct {
	repo repository
	pub  publisher
	log  zerolog.Logger
}

func (s *Service) Save(ctx context.Context, request *api.SaveRequest) (*api.SaveResponse, error) {
	err := s.repo.save(ctx, request.Events)
	if err != nil {
		return nil, err
	}

	s.pub.publish(ctx, request.Events)
	return &api.SaveResponse{}, nil
}

func (s *Service) GetDomain(ctx context.Context, request *api.GetDomainRequest) (*api.GetResponse, error) {
	events, err := s.repo.getDomain(ctx, request.Domain, request.DomainId, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}
	return &api.GetResponse{Events: events}, nil
}

func (s *Service) GetDomainIds(ctx context.Context, request *api.GetDomainIdsRequest) (*api.GetDomainIdsResponse, error) {
	ids, err := s.repo.getDomainIds(ctx, request.Domain, request.Offset, request.Limit)
	if err != nil {
		return nil, err
	}
	return &api.GetDomainIdsResponse{DomainIds: ids}, nil
}

func (s *Service) Subscribe(src api.DeDB_SubscribeServer) error {
	/*
		for {
			if s.shutdown {
				//s.notifyClientsOfRemoval()
				break
			}
			r, err := src.Recv()
			if err == io.EOF {
				s.log.Error().Err(err).Msgf("client broke connection")
				s.removeClient(src)
				return nil
			} else if err != nil {
				s.log.Error().Err(err).Msgf("error receiving from client")
				s.removeClient(src)
				return nil
			} else if r.RequestType == pb.SubscribeEventsRequest_ACK {
				s.ack(r)
			} else if !s.leader {
				s.notifyClientsOfRemoval()
			} else if r.RequestType == pb.SubscribeEventsRequest_CONNECT {
				s.addClient(src, r.ServiceName, r.EventNames)
			} else if r.RequestType == pb.SubscribeEventsRequest_DISCONNECT {
				s.removeClient(src)
			}
		}
	*/
	return nil
}

func (s *Service) Shutdown() {
	s.repo.shutdown()
}

func (s *Service) Start(config Config) error {
	s.log = log.With().Str("logger", "dedbService").Logger()
	s.log.Info().Msg("processing configuration")

	if config.RepoImpl == "redis" {
		r, err := NewRedisRepo(config)
		if err != nil {
			s.log.Error().Err(err).Msg("could not configure redis repo")
			return err
		} else {
			s.repo = r
		}
	} else {
		msg := fmt.Sprintf("repository %s not supported", config.RepoImpl)
		s.log.Error().Msgf(msg)
		return fmt.Errorf(msg)
	}

	if config.BrokerImpl == "redis" {
		p, err := NewRedisPublisher(config)
		if err != nil {
			s.log.Error().Err(err).Msg("could not configure redis publisher")
			return err
		} else {
			s.pub = p
		}
	} else {
		msg := fmt.Sprintf("broker %s not supported", config.BrokerImpl)
		s.log.Error().Msgf(msg)
		return fmt.Errorf(msg)
	}

	s.log.Info().Msg("service initialized")
	return nil
}
