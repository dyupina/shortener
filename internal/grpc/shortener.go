package grpc

import (
	"context"
	"errors"
	models "shortener/internal/domain/models/json"
	pb "shortener/internal/domain/models/proto"
	"shortener/internal/handlers"
	"shortener/internal/repository"
)

// GRPCServer implements gRPC service.
type GRPCServer struct {
	pb.UnimplementedURLShortenerServer
	con *handlers.Controller
}

// NewGRPCServer creates and returns a new instance of the GRPCServer.
func NewGRPCServer(con *handlers.Controller) *GRPCServer {
	return &GRPCServer{
		con: con,
	}
}

// ShortenURL handles requests to create a shortened URL from an incoming URL.
func (s *GRPCServer) ShortenURL(ctx context.Context, req *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	shortID, err := s.con.URLService.ShortenURL(req.OriginalUrl, req.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateURL) {
			return &pb.ShortenURLResponse{ShortUrl: "", ErrorMessage: "URL already exists"}, nil
		}
		return nil, err
	}
	return &pb.ShortenURLResponse{ShortUrl: shortID, ErrorMessage: ""}, nil
}

// GetOriginalURL restores the original URL from a shortened identifier.
func (s *GRPCServer) GetOriginalURL(ctx context.Context, req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	originalURL, isDeleted, err := s.con.URLService.GettingOriginalURL(req.ShortId)
	if err != nil {
		return &pb.GetOriginalURLResponse{
			OriginalUrl:  "",
			IsDeleted:    false,
			ErrorMessage: "Bad Request",
		}, nil
	}

	return &pb.GetOriginalURLResponse{
		OriginalUrl:  originalURL,
		IsDeleted:    isDeleted,
		ErrorMessage: "",
	}, nil
}

// APIShortenURL provides an API for creating a shortened URL from an incoming request.
func (s *GRPCServer) APIShortenURL(ctx context.Context, req *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	shortID, err := s.con.URLService.ShortenURL(req.OriginalUrl, req.UserId)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateURL) {
			return &pb.ShortenURLResponse{
				ShortUrl:     "",
				ErrorMessage: "URL already exists",
			}, nil
		}
		return nil, err
	}

	return &pb.ShortenURLResponse{
		ShortUrl:     shortID,
		ErrorMessage: "",
	}, nil
}

// APIShortenBatchURL handles batch requests for creating shortened URLs from an incoming request.
func (s *GRPCServer) APIShortenBatchURL(ctx context.Context, req *pb.APIShortenBatchURLRequest) (*pb.APIShortenBatchURLResponse, error) {
	var urls []models.BatchRequestEntity
	for _, url := range req.Urls {
		urls = append(urls, models.BatchRequestEntity{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.OriginalUrl,
		})
	}

	batchResponse, err := s.con.URLService.APIShortenBatchURL(req.UserId, urls)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateURL) {
			return &pb.APIShortenBatchURLResponse{
				Results:      nil,
				ErrorMessage: "One or more URLs already exist",
			}, nil
		}
		return nil, err
	}

	var results []*pb.BatchResponseEntity
	for _, result := range batchResponse {
		results = append(results, &pb.BatchResponseEntity{
			CorrelationId: result.CorrelationID,
			ShortUrl:      result.ShortURL,
		})
	}

	return &pb.APIShortenBatchURLResponse{
		Results:      results,
		ErrorMessage: "",
	}, nil
}

// APIGetUserURLs handles requests to retrieve all URLs associated with a user.
func (s *GRPCServer) APIGetUserURLs(ctx context.Context, req *pb.APIGetUserURLsRequest) (*pb.APIGetUserURLsResponse, error) {
	urls, exist := s.con.URLService.APIGetUserURLs(req.UserId)

	if !exist {
		s.con.Logger.Debugf("(APIGetUserURLs) StatusUnauthorized userID %s", req.UserId)
		return &pb.APIGetUserURLsResponse{
			Urls:         nil,
			Exists:       false,
			ErrorMessage: "User not found",
		}, nil
	}

	if len(urls) == 0 {
		s.con.Logger.Debug("(APIGetUserURLs) StatusNoContent")
		return &pb.APIGetUserURLsResponse{
			Urls:         nil,
			Exists:       true,
			ErrorMessage: "",
		}, nil
	}

	var grpcURLs []*pb.UserURL
	for _, url := range urls {
		grpcURLs = append(grpcURLs, &pb.UserURL{
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return &pb.APIGetUserURLsResponse{
		Urls:         grpcURLs,
		Exists:       true,
		ErrorMessage: "",
	}, nil
}

// DeleteUserURLs handles HTTP requests to delete URLs belonging to a user.
func (s *GRPCServer) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	if req.UserId == "" {
		return &pb.DeleteUserURLsResponse{
			Success:      false,
			ErrorMessage: "Unauthorized",
		}, nil
	}

	resultCh, err := s.con.URLService.DeleteUserURLs(req.UserId, req.UrlIds)
	if err != nil {
		return &pb.DeleteUserURLsResponse{
			Success:      false,
			ErrorMessage: "Failed to delete URLs",
		}, nil
	}

	go func() {
		for res := range resultCh {
			s.con.Logger.Infof("Deleted short URL: %s", res)
		}
	}()

	return &pb.DeleteUserURLsResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

// Ping checks the connection to the data storage.
func (s *GRPCServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	err := s.con.URLService.PingHandler()
	if err != nil {
		return &pb.PingResponse{
			Success:      false,
			ErrorMessage: "Database connection error",
		}, nil
	}

	return &pb.PingResponse{
		Success:      true,
		ErrorMessage: "",
	}, nil
}

// Statistics returns the number of users and the number of shortened URLs in the service.
func (s *GRPCServer) Statistics(ctx context.Context, req *pb.StatisticsRequest) (*pb.StatisticsResponse, error) {
	trustedSubnet := s.con.Config.TrustedSubnet
	if trustedSubnet == "" {
		s.con.Logger.Debugf("Access Denied (empty trusted_subnet)")
		return &pb.StatisticsResponse{
			UsersCount:   0,
			UrlsCount:    0,
			ErrorMessage: "Access Denied (empty trusted_subnet)",
		}, nil
	}

	clientIP := req.ClientIp
	if !s.con.IsIPInSubnet(clientIP, trustedSubnet) {
		s.con.Logger.Debugf("Access Denied (IP not in specified subnet)")
		return &pb.StatisticsResponse{
			UsersCount:   0,
			UrlsCount:    0,
			ErrorMessage: "Access Denied (IP not in specified subnet)",
		}, nil
	}

	stats := s.con.URLService.Statistics()

	return &pb.StatisticsResponse{
		UsersCount:   int64(stats.Users),
		UrlsCount:    int64(stats.URLs),
		ErrorMessage: "",
	}, nil
}
