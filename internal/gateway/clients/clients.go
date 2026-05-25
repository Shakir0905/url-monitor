package clients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	analyticspb "github.com/shakir/url-monitor/proto/analytics"
	authpb "github.com/shakir/url-monitor/proto/auth"
	urlpb "github.com/shakir/url-monitor/proto/url"
)

type Clients struct {
	Auth      authpb.AuthServiceClient
	URL       urlpb.URLServiceClient
	Analytics analyticspb.AnalyticsServiceClient
}

func New(authAddr, urlAddr, analyticsAddr string) (*Clients, func(), error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	authConn, err := grpc.Dial(authAddr, dialOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("dial auth: %w", err)
	}

	urlConn, err := grpc.Dial(urlAddr, dialOpts...)
	if err != nil {
		authConn.Close()
		return nil, nil, fmt.Errorf("dial url: %w", err)
	}

	analyticsConn, err := grpc.Dial(analyticsAddr, dialOpts...)
	if err != nil {
		authConn.Close()
		urlConn.Close()
		return nil, nil, fmt.Errorf("dial analytics: %w", err)
	}

	cleanup := func() {
		authConn.Close()
		urlConn.Close()
		analyticsConn.Close()
	}

	return &Clients{
		Auth:      authpb.NewAuthServiceClient(authConn),
		URL:       urlpb.NewURLServiceClient(urlConn),
		Analytics: analyticspb.NewAnalyticsServiceClient(analyticsConn),
	}, cleanup, nil
}
