package newsrooms

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/joincivil/go-common/pkg/newsroom"
	"github.com/patrickmn/go-cache"
	"time"
)

type Service interface {
	GetNewsroomByAddress(newsroomAddress string) (*newsroom.Newsroom, error)
}

type CachingService struct {
	base  *newsroom.Service
	cache *cache.Cache
}

func NewCachingService(base *newsroom.Service) *CachingService {
	// creates an in-memory cache where items expire every 2 minutes and purges expired items every 4 minutes
	c := cache.New(2*time.Minute, 4*time.Minute)

	return &CachingService{
		base:  base,
		cache: c,
	}
}

func (s *CachingService) GetNewsroomByAddress(newsroomAddress string) (*newsroom.Newsroom, error) {
	hit, found := s.cache.Get(newsroomAddress)
	if found {
		return hit.(*newsroom.Newsroom), nil
	}

	addr := common.HexToAddress(newsroomAddress)

	name, err := s.base.GetNewsroomName(addr)
	if err != nil {
		return nil, err
	}

	multisigAddress, err := s.base.GetOwner(addr)
	if err != nil {
		return nil, err
	}

	charter, err := s.base.GetCharter(addr)
	if err != nil {
		return nil, err
	}

	newsroom := &newsroom.Newsroom{
		Name:            name,
		ContractAddress: newsroomAddress,
		MultisigAddress: multisigAddress.String(),
		Charter:         charter,
	}
	s.cache.Set(newsroomAddress, newsroom, cache.DefaultExpiration)

	return newsroom, nil
}