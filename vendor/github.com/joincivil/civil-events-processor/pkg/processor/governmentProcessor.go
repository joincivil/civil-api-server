package processor

// TODO
// import (
//     "errors"
//     "fmt"
//     "math/big"
//     "strings"

//     "github.com/ethereum/go-ethereum/accounts/abi/bind"
//     log "github.com/golang/glog"

//     commongen "github.com/joincivil/civil-events-crawler/pkg/generated/common"
//     crawlermodel "github.com/joincivil/civil-events-crawler/pkg/model"

//     "github.com/joincivil/civil-events-processor/pkg/model"

//     cpersist "github.com/joincivil/go-common/pkg/persistence"
//     ctime "github.com/joincivil/go-common/pkg/time"
// )

// // NOTE for now only track ParameterSet and AppelateSet events

// // NewGovernmentEventProcessor is a convenience function to init a government processor
// func NewGovernmentEventProcessor(client bind.ContractBackend) *ParameterizerEventProcessor {
//     return &GovernmentEventProcessor{
//         client:             client
//     }
// }

// // GovernmentEventProcessor handles the processing of raw events into aggregated data
// type GovernmentEventProcessor struct {
//     client             bind.ContractBackend
// }

// func (p *ParameterizerEventProcessor) isValidGovernmentContractEventName(name string) bool {
//     name = strings.Trim(name, " _")
//     eventNames := commongen.EventTypesParameterizerContract()
//     return isStringInSlice(eventNames, name)
// }
